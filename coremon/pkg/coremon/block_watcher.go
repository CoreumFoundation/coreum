// Package coremon implements the service.
package coremon

import (
	"context"
	"sync"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/pkg/errors"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/coremon/pkg/tmclient"
)

type TmBlockWatcher interface {
	StartWatching(latestSyncedBlock uint64)
	LastSyncedBlock() uint64
	IsSynced() bool
	WaitForSync()
	Close() error
}

type BlockGetter interface {
	NewBlockDataChan() <-chan NewBlockData
	Close()
}

type NewBlockHandlerFn func(ev NewBlockData) error

func NewTmBlockWatcher(
	ctx context.Context,
	chainID string,
	tmRPC string,
	protoCodec *codec.ProtoCodec,
	parallelBlockFetchJobs int,
	blockDataHandler NewBlockHandlerFn,
) (watcher TmBlockWatcher, err error) {
	tmClient, err := tmclient.NewRPCClient(tmRPC)
	if err != nil {
		return nil, err
	}

	w := &tmBlockWatcher{
		rootCtx:  ctx,
		chainID:  chainID,
		tmRPC:    tmRPC,
		tmClient: tmClient,

		parallelBlockFetchJobs: parallelBlockFetchJobs,
		blockDataHandler:       blockDataHandler,

		handlerWG:            new(sync.WaitGroup),
		protoCodec:           protoCodec,
		latestSyncedBlockMux: new(sync.RWMutex),
		isSynced:             make(chan struct{}, 1),
		isClosing:            make(chan struct{}, 1),
		isDone:               make(chan struct{}, 1),

		logger: logger.Get(ctx).With(zap.String("svc", "block_watcher")),
	}

	return w, nil
}

type Status struct {
	LastBlock     uint64    `json:"last_block"`
	LastBlockTime time.Time `json:"last_block_time"`
}

type tmBlockWatcher struct {
	rootCtx  context.Context
	chainID  string
	tmRPC    string
	tmClient tmclient.TendermintClient

	blockGetter            BlockGetter
	parallelBlockFetchJobs int
	blockDataHandler       NewBlockHandlerFn

	handlerWG            *sync.WaitGroup
	latestSyncedBlock    uint64
	latestSyncedBlockMux *sync.RWMutex
	protoCodec           *codec.ProtoCodec

	isSynced  chan struct{}
	isClosing chan struct{}
	isDone    chan struct{}

	logger *zap.Logger
}

func (w *tmBlockWatcher) Close() (err error) {
	w.logger.Info("TmBlockWatcher exits")

	defer func() {
		w.logger.Info("TmBlockWatcher exited")
	}()

	close(w.isClosing)

	w.handlerWG.Wait()

	deadlineT := time.NewTimer(3 * time.Second)
	select {
	case <-w.isDone:
		// closed upon the latest block processed, no more blocks will be processed
	case <-deadlineT.C:
		// no new block in 3s, consider this stale
	}

	return
}

func (w *tmBlockWatcher) WaitForSync() {
	select {
	case <-w.isSynced:
	case <-w.isClosing:
	}
}

func (w *tmBlockWatcher) IsSynced() bool {
	select {
	case <-w.isSynced:
		return true
	default:
		return false
	}
}

func (w *tmBlockWatcher) LastSyncedBlock() uint64 {
	w.latestSyncedBlockMux.RLock()
	defer w.latestSyncedBlockMux.RUnlock()

	return w.latestSyncedBlock
}

func (w *tmBlockWatcher) runInitialSync(ctx context.Context, latestSyncedBlock uint64) error {
	var upperBound uint64

	height, ts, err := w.tmClient.GetLatestBlockHeight(ctx)
	if err != nil {
		err = errors.Wrap(err, "failed to get latest block info from chain daemon")
		return err
	}

	if !ts.IsZero() {
		w.logger.Sugar().Infof("Block Sync: At block height %d while chain is at %d (%s)",
			latestSyncedBlock, height, ts.Format(time.RFC3339),
		)
	} else {
		w.logger.Sugar().Infof("Block Sync: At block height %d while chain is at %d",
			latestSyncedBlock, height,
		)
	}

	upperBound = uint64(height)

	if latestSyncedBlock >= upperBound {
		// already synced
		return nil
	}

	// TODO: inital sync logic there if needed to access past blocks

	w.latestSyncedBlockMux.Lock()
	w.latestSyncedBlock = uint64(height)
	w.latestSyncedBlockMux.Unlock()

	return nil
}

func (w *tmBlockWatcher) StartWatching(latestSyncedBlock uint64) {
	// initial sync: wait until getting up to chain height
	if err := w.runInitialSync(w.rootCtx, latestSyncedBlock); err != nil {
		w.logger.With(zap.Error(err)).Fatal("failed to run initial block sync")
		return
	}

	w.latestSyncedBlockMux.RLock()
	w.blockGetter = w.initBlockGetter(w.latestSyncedBlock+1, w.parallelBlockFetchJobs)
	w.latestSyncedBlockMux.RUnlock()

	newBlocks := w.blockGetter.NewBlockDataChan()

	// signal that the syncing is officially done
	close(w.isSynced)

	w.logger.Info("Block Sync: Initial sync done. Continuing to poll TmRPC for the new blocks.")

	for {
		select {
		case <-w.isClosing:
			// no more new blocks
			return
		case <-w.isDone:
			// no more new blocks
			return
		case block, ok := <-newBlocks:
			if !ok {
				// no more new blocks
				return
			}

			if err := w.handleNewBlockData(block); err != nil {
				if err == ErrShuttingDown {
					return
				}

				w.logger.With(zap.Error(err)).Sugar().Fatalf("failed to sync recent block, stopped at %d", w.LastSyncedBlock())
				return
			}
		}
	}
}

var ErrShuttingDown = errors.New("shutting down")

type NewBlockData struct {
	Block        *tmtypes.Block
	BlockResults *ctypes.ResultBlockResults
}

func (w *tmBlockWatcher) handleNewBlockData(data NewBlockData) (err error) {
	if err = w.blockDataHandler(data); err != nil {
		return err
	}

	w.latestSyncedBlockMux.Lock()
	w.latestSyncedBlock = uint64(data.Block.Height)
	w.latestSyncedBlockMux.Unlock()

	select {
	case <-w.isClosing:
		// this is the latest block handled before exit
		close(w.isDone)

		if err == nil {
			err = ErrShuttingDown
		}
	default:
	}

	return err
}

func (w *tmBlockWatcher) initBlockGetter(initHeight uint64, parallelJobs int) BlockGetter {
	getter := &blockGetter{
		tmClient: w.tmClient,

		jobs:    parallelJobs,
		jobMux:  new(sync.RWMutex),
		jobCond: sync.NewCond(new(sync.Mutex)),
		height:  initHeight,

		newBlocksC:   make(chan NewBlockData, parallelJobs*100),
		newBlocksMap: make(map[uint64]NewBlockData, parallelJobs*100),
		closeC:       make(chan struct{}, 1),

		logger: w.logger,
	}

	w.logger.Debug("initBlockGetter ready to announce and pull blocks")

	go getter.announceBlocks(initHeight)
	go getter.pullBlocks()

	return getter
}

type blockGetter struct {
	tmClient tmclient.TendermintClient

	jobs    int
	jobMux  *sync.RWMutex
	jobCond *sync.Cond
	height  uint64

	newBlocksC   chan NewBlockData
	newBlocksMap map[uint64]NewBlockData
	closeC       chan struct{}

	logger *zap.Logger
}

func (b *blockGetter) announceBlocks(startHeight uint64) {
	height := startHeight

	for {
		select {
		case <-b.closeC:
			close(b.newBlocksC)
			return
		default:
			b.jobCond.L.Lock()

			ev, found := b.newBlocksMap[height]
			for !found {
				b.jobCond.Wait()

				ev, found = b.newBlocksMap[height]

				select {
				case <-b.closeC:
					close(b.newBlocksC)
					return
				default:
					continue
				}
			}

			delete(b.newBlocksMap, height)
			b.jobCond.L.Unlock()

			b.newBlocksC <- ev

			height++
			continue
		}
	}
}

func (b *blockGetter) pullBlocks() {
	clientCtx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	wg := new(sync.WaitGroup)
	defer wg.Wait()

	for i := 0; i < b.jobs; i++ {
		wg.Add(1)
		go func(jobID int) {
			defer wg.Done()

			getHeightToFetch := func() uint64 {
				b.jobMux.Lock()
				defer b.jobMux.Unlock()

				h := b.height
				b.height++

				return h
			}

			// step 1: obtain new height to fetch
			height := getHeightToFetch()

			for {
				select {
				case <-b.closeC:
					return
				default:
					jobLog := b.logger.With(
						zap.Int("job", jobID),
						zap.Uint64("height", height),
					)

					// step 2: try to fetch it until cancelled
					newBlock, err := b.fetchBlockByNum(clientCtx, height)
					if err != nil {
						// TODO: ignore future block errors
						jobLog.With(zap.Error(err)).Warn("failed to fully fetch block, retry in 1s")

						time.Sleep(1 * time.Second)
						continue
					}

					var tooFarIntoFuture = false

					b.jobCond.L.Lock()
					b.newBlocksMap[height] = newBlock
					b.jobCond.Signal()

					if len(b.newBlocksMap) > 1024*b.jobs {
						tooFarIntoFuture = true
					}
					b.jobCond.L.Unlock()

					// job sleeps until backlog is too high
					for tooFarIntoFuture {
						time.Sleep(200 * time.Millisecond)

						b.jobCond.L.Lock()
						tooFarIntoFuture = len(b.newBlocksMap) > 1024*b.jobs
						b.jobCond.L.Unlock()
					}

					// step 3: assign a new height to the job
					height = getHeightToFetch()
				}
			}
		}(i)
	}
}

func (b *blockGetter) fetchBlockByNum(ctx context.Context, height uint64) (NewBlockData, error) {
	blockC := make(chan *ctypes.ResultBlock, 1)
	blockResultsC := make(chan *ctypes.ResultBlockResults, 1)
	errC := make(chan error, 4)

	retryOpts := []retry.Option{
		retry.Attempts(10),
		retry.MaxDelay(5 * time.Second),
	}

	go func() {
		defer close(blockC)

		if err := retry.Do(func() error {
			block, err := b.tmClient.GetBlock(ctx, int64(height))
			if err != nil {
				err = errors.Wrapf(err, "failed to get block info (%d) from chain daemon, will retry", height)
				return err
			}

			blockC <- block
			return nil
		}, retryOpts...); err != nil {
			errC <- err
		}
	}()

	go func() {
		defer close(blockResultsC)

		if err := retry.Do(func() error {
			blockResults, err := b.tmClient.GetBlockResults(ctx, int64(height))
			if err != nil {
				err = errors.Wrapf(err, "failed to get block results (%d) from chain daemon", height)
				return err
			}

			blockResultsC <- blockResults
			return nil
		}, retryOpts...); err != nil {
			errC <- err
		}
	}()

	block := <-blockC
	blockResults := <-blockResultsC
	select {
	case err := <-errC:
		return NewBlockData{}, err
	default:
		close(errC)
	}

	return NewBlockData{
		Block:        block.Block,
		BlockResults: blockResults,
	}, nil
}

func (b *blockGetter) NewBlockDataChan() <-chan NewBlockData {
	return b.newBlocksC
}

func (b *blockGetter) Close() {
	if b.jobCond != nil {
		b.jobCond.Signal()
	}

	close(b.closeC)
}
