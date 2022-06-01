package zstress

import (
	"context"
	"fmt"
	"math/big"
	"runtime"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/parallel"
	"github.com/pkg/errors"
	"github.com/xlab/pace"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
)

// StressConfig contains config for benchmarking the blockchain
type StressConfig struct {
	// ChainID is the ID of the chain to connect to
	ChainID string

	// NodeAddress is the address of a cored node RPC endpoint, in the form of host:port, to connect to
	NodeAddress string

	// Accounts is the list of private keys used to send transactions during benchmark
	Accounts []cored.Secp256k1PrivateKey

	// NumOfTransactions to send from each account
	NumOfTransactions int
}

type tx struct {
	AccountIndex int
	TxIndex      int
	From         cored.Wallet
	To           cored.Wallet
	TxBytes      []byte
}

// Stress runs a benchmark test
func Stress(ctx context.Context, config StressConfig) error {
	numOfAccounts := len(config.Accounts)
	log := logger.Get(ctx)
	startTs := time.Now()
	client := cored.NewClient(config.ChainID, config.NodeAddress)

	signedTxPace := pace.New("signed tx", 10*time.Second, zapPaceReporter(log))

	log.Info(
		"Preparing signed transactions...",
		zap.Int("amount", numOfAccounts*config.NumOfTransactions),
	)

	var signedTxs [][][]byte
	err := parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		queue := make(chan tx)
		results := make(chan tx)
		for i := 0; i < runtime.NumCPU(); i++ {
			spawn(fmt.Sprintf("signer-%d", i), parallel.Continue, func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case tx, ok := <-queue:
						if !ok {
							return nil
						}
						tx.TxBytes = must.Bytes(client.TxBankSend(tx.From, tx.To, cored.Balance{Amount: big.NewInt(1), Denom: "core"}))
						select {
						case <-ctx.Done():
							return ctx.Err()
						case results <- tx:
						}
					}
				}
			})
		}
		spawn("enqueue", parallel.Continue, func(ctx context.Context) error {
			for i := 0; i < numOfAccounts; i++ {
				fromPrivateKey := config.Accounts[i]
				toPrivateKeyIndex := i + 1
				if toPrivateKeyIndex >= numOfAccounts {
					toPrivateKeyIndex = 0
				}
				toPrivateKey := config.Accounts[toPrivateKeyIndex]

				accNum, accSeq, err := client.GetNumberSequence(fromPrivateKey.Address())
				if err != nil {
					return errors.WithStack(fmt.Errorf("fetching account number and sequence failed: %w", err))
				}

				tx := tx{
					AccountIndex: i,
					From:         cored.Wallet{Name: "sender", Key: fromPrivateKey, AccountNumber: accNum, AccountSequence: accSeq},
					To:           cored.Wallet{Name: "receiver", Key: toPrivateKey},
				}

				for j := 0; j < config.NumOfTransactions; j++ {
					tx.TxIndex = j
					select {
					case <-ctx.Done():
						return ctx.Err()
					case queue <- tx:
					}
					tx.From.AccountSequence++
				}
			}
			return nil
		})
		spawn("integrate", parallel.Exit, func(ctx context.Context) error {
			signedTxs = make([][][]byte, numOfAccounts)

			defer func() {
				signedTxPace.Pause()
			}()

			for i := 0; i < numOfAccounts; i++ {
				signedTxs[i] = make([][]byte, config.NumOfTransactions)
			}
			for i := 0; i < numOfAccounts; i++ {
				for j := 0; j < config.NumOfTransactions; j++ {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case result := <-results:
						signedTxs[result.AccountIndex][result.TxIndex] = result.TxBytes
						signedTxPace.StepN(1)
					}
				}
			}
			return nil
		})
		return nil
	})
	if err != nil {
		return err
	}
	log.Info("Transactions prepared")

	broadcastTxPace := pace.New("sent tx", 10*time.Second, zapPaceReporter(log))

	log.Info("Broadcasting transactions...")
	err = parallel.Run(ctx, func(ctx context.Context, spawn parallel.SpawnFn) error {
		for i, accountTxs := range signedTxs {
			accountTxs := accountTxs
			spawn(fmt.Sprintf("account-%d", i), parallel.Continue, func(ctx context.Context) error {
				log := logger.Get(ctx)
				j := 0
				for {
					if err := ctx.Err(); err != nil {
						return err
					}
					tx := accountTxs[j]
					res, err := client.Broadcast(ctx, tx)
					if err != nil {
						// TODO(xlab): make sure we're abort there, as broadcasting the tx queue for an acconut
						// makes no sense if one nonce slot has been skipped.
						log.Error("Sending transaction failed", zap.Error(err))
						continue
					}

					broadcastTxPace.Step(1)

					log.Debug("Transaction broadcasted", zap.String("txHash", res.TxHash))

					j++
					if j >= config.NumOfTransactions {
						return nil
					}
				}
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	broadcastTxPace.Pause()

	log.Info(
		"Benchmark finished",
		zap.Duration("dur", time.Since(startTs)),
	)

	return nil
}
