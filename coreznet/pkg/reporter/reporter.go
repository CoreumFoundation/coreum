package reporter

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"go.uber.org/zap"
)

// New creates new reporter
func New(title string, reportingPeriod time.Duration) *Reporter {
	return &Reporter{
		title:  title,
		period: reportingPeriod,
	}
}

// Reporter creates a metrics and reports them every period
type Reporter struct {
	title    string
	period   time.Duration
	counters []func(scale float64) zap.Field
}

// Run is the code to be run in a goroutine, it periodically reports the metrics
func (r *Reporter) Run(ctx context.Context) error {
	log := logger.Get(ctx)
	scale := float64(time.Second) / float64(r.period)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.period):
		}

		fields := make([]zap.Field, 0, len(r.counters))
		for _, c := range r.counters {
			fields = append(fields, c(scale))
		}
		log.Info(r.title, fields...)
	}
}

// Throughput reports the actual number of items per second being processed
func (r *Reporter) Throughput(label string) func(step uint64) {
	var counter uint64

	r.counters = append(r.counters, func(scale float64) zap.Field {
		return zap.Float64(label, float64(atomic.SwapUint64(&counter, 0))/scale)
	})

	return func(step uint64) {
		atomic.AddUint64(&counter, step)
	}
}

// Progress reports the percentage of items processed so far
func (r *Reporter) Progress(label string, total uint64) func(step uint64) {
	totalFloat := float64(total)
	var all uint64

	r.counters = append(r.counters, func(scale float64) zap.Field {
		return zap.Float64(label, 100.*float64(atomic.LoadUint64(&all))/totalFloat)
	})

	return func(step uint64) {
		atomic.AddUint64(&all, step)
	}
}
