package coremon

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	influxwrite "github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/pace"
	"github.com/CoreumFoundation/coreum/coremon/pkg/statsd_metrics"
)

func NewBlockHandlerWithMetrics(
	ctx context.Context,
	chainID string,
	influxWriteAPI influxdb2api.WriteAPI,
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
			s.Gauge("report.observed_txs_throughput", throughputReal, tagSpec)
		}, metricTags)
	})

	blocksPaceReporting := pace.New("", 15*time.Second, func(_ string, timeframe time.Duration, value int) {
		// blocksPace is the block produce speed measured relative to the real-world time clock
		blocksPace := float64(value) / (float64(timeframe) / float64(time.Second))

		statsd_metrics.Report(func(s statsd_metrics.Statter, tagSpec []string) {
			s.Gauge("report.observed_blocks_pace", blocksPace, tagSpec)
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
			s.Timing("report.ingest_latency", latency, tagSpec)
			s.Gauge("report.observed_height", blockNumber, tagSpec)

			if txsInBlock > 0 {
				s.Count("report.observed_txs_total", txsInBlock, tagSpec)
			}
		}, metricTags)

		pointsToWrite := make([]*influxwrite.Point, 0, 1)
		{
			// assemble 'coremon_block_report' measurement point with all fields about particular block

			p := influxdb2.NewPointWithMeasurement("coremon_block_report")
			p = p.SetTime(data.Block.Time)
			p = p.AddField("height", data.Block.Height)

			if txsInBlock > 0 {
				p = p.AddField("txs", txsInBlock)
			}

			if blockTimeDiff > 0 {
				p = p.AddField("time_diff", float64(blockTimeDiff)/float64(time.Millisecond))

				if txTroughputAbs > 0 {
					p = p.AddField("txs_throughput", txTroughputAbs)
				}
			}

			allTags := metricTags.WithBaseTags()
			for k, v := range allTags {
				p = p.AddTag(k, v)
			}

			pointsToWrite = append(pointsToWrite, p)
		}

		if len(pointsToWrite) > 0 {
			writeInfluxPoints(influxWriteAPI, pointsToWrite)
		}

		return nil
	}
}

func writeInfluxPoints(
	writeAPI influxdb2api.WriteAPI,
	points []*influxwrite.Point,
) {
	defer func() {
		writeAPI.Flush()
	}()

	for _, point := range points {
		writeAPI.WritePoint(point)
	}
}
