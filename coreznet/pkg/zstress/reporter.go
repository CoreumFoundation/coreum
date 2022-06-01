package zstress

import (
	"strconv"
	"time"

	"github.com/xlab/pace"
	"go.uber.org/zap"
)

// zapPaceReporter reports using our own zap logger and stops reporting when flow of events is stoped.
func zapPaceReporter(log *zap.Logger) pace.ReporterFunc {
	var previous float64
	var stalled time.Time
	return func(label string, timeframe time.Duration, value float64) {
		switch {
		case value == 0 && previous == 0:
			return // don't report anything
		case value == 0 && previous != 0:
			dur := timeframe
			if !stalled.IsZero() {
				dur = time.Since(stalled)
				n := dur / timeframe
				if dur-n*timeframe < 10*time.Millisecond {
					dur = n * timeframe
				}
			} else {
				stalled = time.Now().Add(-dur)
			}
			log.Sugar().Infof("%s: stalled for %v", label, dur)
			return
		default:
			previous = value
			stalled = time.Time{}
		}
		floatFmt := func(f float64) string {
			return strconv.FormatFloat(f, 'f', 3, 64)
		}
		switch timeframe {
		case time.Second:
			log.Sugar().Infof("%s: %s/s in %v", label, floatFmt(value), timeframe)
		case time.Minute:
			log.Sugar().Infof("%s: %s/m in %v", label, floatFmt(value), timeframe)
		case time.Hour:
			log.Sugar().Infof("%s: %s/h in %v", label, floatFmt(value), timeframe)
		case 24 * time.Hour:
			log.Sugar().Infof("%s: %s/day in %v", label, floatFmt(value), timeframe)
		default:
			log.Sugar().Infof("%s %s in %v (pace: %s/s)", floatFmt(value), label,
				timeframe, floatFmt(value/(float64(timeframe)/float64(time.Second))))
		}
	}
}
