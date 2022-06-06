package coremon

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/pace"
)

func NewBlockHandlerWithMetrics(
	ctx context.Context,
) NewBlockHandlerFn {
	log := logger.Get(ctx)

	newBlockHandlerPace := pace.New("blocks synced", 1*time.Minute, pace.ZapReporter(log))
	txInBlocksPace := pace.New("tx seen", 1*time.Minute, pace.ZapReporter(log))

	return func(data NewBlockData) error {
		blockNumber := uint64(data.Block.Height)

		latency := time.Now().Sub(data.Block.Time)
		log.Debug("Got new block",
			zap.Uint64("height", blockNumber),
			zap.Duration("latency", latency),
		)

		newBlockHandlerPace.Step(1)
		txInBlocksPace.Step(len(data.Block.Txs))

		return nil
	}
}
