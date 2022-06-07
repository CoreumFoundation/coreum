package main

import (
	"os"
	"time"

	cli "github.com/jawher/mow.cli"
	"go.uber.org/zap"

	closer "github.com/CoreumFoundation/coreum-tools/pkg/closer"
	statsd_metrics "github.com/CoreumFoundation/coreum/coremon/pkg/statsd_metrics"
)

func initMetrics(c *cli.Cmd) {
	var (
		statsdPrefix  *string
		statsdAddr    *string
		statsdEnabled *string
	)

	initStatsdOptions(
		c,
		&statsdPrefix,
		&statsdAddr,
		&statsdEnabled,
	)

	if toBool(*statsdEnabled) {
		appLogger.Info("statsd reporter is enabled", zap.String("target", *statsdAddr))
		go func() {
			for {
				hostname, _ := os.Hostname()
				err := statsd_metrics.Init(*statsdAddr, checkStatsdPrefix(*statsdPrefix), &statsd_metrics.StatterConfig{
					EnvName:  *envName,
					HostName: hostname,
				})

				if err != nil {
					appLogger.With(zap.Error(err)).Warn("failed to init statsd reporter")
					time.Sleep(time.Minute)
					continue
				}

				break
			}

			closer.Bind(func() {
				statsd_metrics.Close()
			})
		}()
	} else {
		statsd_metrics.Disable()
	}

}
