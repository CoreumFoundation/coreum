package main

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cosmtypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	cli "github.com/jawher/mow.cli"
	"go.uber.org/zap"

	closer "github.com/CoreumFoundation/coreum-tools/pkg/closer"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/coremon/pkg/coremon"
)

func processCmd(c *cli.Cmd) {
	var (
		chainID                *string
		tendermintRPC          *string
		parallelBlockFetchJobs *int

		influxEnabled  *bool
		influxEndpoint *string
		influxDBName   *string
		influxUser     *string
		influxPassword *string
	)

	initCosmosOptions(
		c,
		&chainID,
		&tendermintRPC,
		&parallelBlockFetchJobs,
	)

	initInfluxOptions(
		c,
		&influxEnabled,
		&influxEndpoint,
		&influxDBName,
		&influxUser,
		&influxPassword,
	)

	c.Before = func() {
		initMetrics(c)

		appLogger.Info("CoreMon Block Watching routine starts")
	}

	c.Action = func() {
		defer closer.Close()

		closer.Bind(func() {
			rootCancelFn()
		})

		var (
			influxClient   influxdb2.Client
			influxWriteAPI influxdb2api.WriteAPI
		)

		if *influxEnabled {
			influxClient = influxdb2.NewClient(*influxEndpoint, *influxUser+":"+*influxPassword)
			influxWriteAPI = influxClient.WriteAPI("", *influxDBName)

			errorsCh := influxWriteAPI.Errors()
			go func() {
				for err := range errorsCh {
					appLogger.With(zap.Error(err)).Warn("InfluxDB write error")
				}
			}()

			closer.Bind(func() {
				influxClient.Close()
			})
		}

		interfaceRegistry := cosmtypes.NewInterfaceRegistry()
		std.RegisterInterfaces(interfaceRegistry)
		protoCodec := codec.NewProtoCodec(interfaceRegistry)

		blockWatcher, err := coremon.NewTmBlockWatcher(
			rootCtx,
			*chainID,
			*tendermintRPC,
			protoCodec,
			*parallelBlockFetchJobs,
			coremon.NewBlockHandlerWithMetrics(rootCtx, *chainID, influxWriteAPI),
		)
		must.OK(err)

		// Launch chain block watcher routine
		//

		go blockWatcher.StartWatching(0)
		closer.Bind(func() {
			blockWatcher.Close()
		})

		//
		// Wait till Ctrl+C
		//

		closer.Hold()
	}
}
