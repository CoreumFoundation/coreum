package coremon

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/pace"
	"github.com/CoreumFoundation/coreum/coremon/pkg/statsd_metrics"
)

func NewBlockHandlerWithMetrics(
	ctx context.Context,
	chainID string,
) NewBlockHandlerFn {
	log := logger.Get(ctx)

	metricTags := statsd_metrics.Tags{
		"svc":      "coremon",
		"chain_id": chainID,
	}

	blockTimestamps := make(map[uint64]time.Time)

	newBlockHandlerPace := pace.New("blocks synced", 1*time.Minute, pace.ZapReporter(log))
	txsInBlocksPace := pace.New("tx throughput", 1*time.Minute, pace.ZapReporter(log))

	txThroughputReporting := pace.New("", 15*time.Second, func(_ string, timeframe time.Duration, value int) {
		// throughputReal is the tx throughput measured relative to the real-world time clock
		throughputReal := float64(value) / (float64(timeframe) / float64(time.Second))

		statsd_metrics.Report(func(s statsd_metrics.Statter, tagSpec []string) {
			s.Gauge("block_reports.tx_tp_real", throughputReal, tagSpec)
		}, metricTags)
	})

	blocksPaceReporting := pace.New("", 15*time.Second, func(_ string, timeframe time.Duration, value int) {
		// blocksPace is the block produce speed measured relative to the real-world time clock
		blocksPace := float64(value) / (float64(timeframe) / float64(time.Second))

		statsd_metrics.Report(func(s statsd_metrics.Statter, tagSpec []string) {
			s.Gauge("block_reports.blocks_pace", blocksPace, tagSpec)
		}, metricTags)
	})

	return func(data NewBlockData) error {
		blockNumber := uint64(data.Block.Height)
		txsInBlock := len(data.Block.Txs)

		latency := time.Now().Sub(data.Block.Time)
		log.Debug("Got new block",
			zap.Uint64("height", blockNumber),
			zap.Duration("latency", latency),
		)

		var (
			// blockTimeDiff is the difference between two finalized block timestamps,
			// enough to compute avg blocktime in the metrics postprocessing.
			blockTimeDiff time.Duration

			// txTroughputAbs is the absolute throughput, based on num of transactions
			// included in the block that was finalized in blockTimeDiff.
			txTroughputAbs float64
		)

		blockTimestamps[blockNumber] = data.Block.Time
		if prevBlockTimestamp, ok := blockTimestamps[blockNumber-1]; ok {
			blockTimeDiff = data.Block.Time.Sub(prevBlockTimestamp)
			txTroughputAbs = float64(txsInBlock) / (float64(blockTimeDiff) / float64(time.Second))
		}

		newBlockHandlerPace.Step(1)
		blocksPaceReporting.Step(1)
		txsInBlocksPace.Step(txsInBlock)
		txThroughputReporting.Step(txsInBlock)

		statsd_metrics.Report(func(s statsd_metrics.Statter, tagSpec []string) {
			s.Gauge("block_reports.height", blockNumber, tagSpec)
			s.Timing("block_reports.ingest_latency", latency, tagSpec)

			if txsInBlock > 0 {
				s.Gauge("block_reports.tx_per_block", txsInBlock, tagSpec)
				s.Count("block_reports.tx_total", txsInBlock, tagSpec)
			}

			if blockTimeDiff > 0 {
				s.Timing("block_reports.blocktime_diff", blockTimeDiff, tagSpec)

				if txTroughputAbs > 0 {
					s.Gauge("block_reports.tx_tp_abs", txTroughputAbs, tagSpec)
				}
			}
		}, metricTags)

		return nil
	}
}
