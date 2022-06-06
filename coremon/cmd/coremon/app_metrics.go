package main

import (
	"os"
	"time"

	cli "github.com/jawher/mow.cli"

	closer "github.com/CoreumFoundation/coreum-tools/pkg/closer"
	statsd_metrics "github.com/CoreumFoundation/coreum/coremon/pkg/statsd_metrics"
)

func initMetrics(c *cli.Cmd) {
	var (
		statsdPrefix   *string
		statsdAddr     *string
		statsdDisabled *string
	)

	initStatsdOptions(
		c,
		&statsdPrefix,
		&statsdAddr,
		&statsdDisabled,
	)

	if toBool(*statsdDisabled) {
		statsd_metrics.Disable()
	} else {
		go func() {
			for {
				hostname, _ := os.Hostname()
				err := statsd_metrics.Init(*statsdAddr, checkStatsdPrefix(*statsdPrefix), &statsd_metrics.StatterConfig{
					EnvName:  *envName,
					HostName: hostname,
				})

				if err != nil {
					time.Sleep(time.Minute)
					continue
				}

				break
			}

			closer.Bind(func() {
				statsd_metrics.Close()
			})
		}()
	}

}
