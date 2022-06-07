package reporter

import (
	"context"
	"math"
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

// UInt32 creates metric of type uint32
func (r *Reporter) UInt32(label string, total uint32) func(step uint32) {
	progressLabel := label + "Progress"
	totalFloat := float64(total)
	var duringPeriod uint32
	var value, all float64

	r.counters = append(r.counters, func(scale float64) zap.Field {
		value = float64(atomic.SwapUint32(&duringPeriod, 0))
		return zap.Float64(label, value/scale)
	}, func(scale float64) zap.Field {
		all += value
		return zap.Float64(progressLabel, math.Round(100*all/totalFloat))
	})

	return func(step uint32) {
		atomic.AddUint32(&duringPeriod, step)
	}
}
