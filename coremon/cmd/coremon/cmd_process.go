package main

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cosmtypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	cli "github.com/jawher/mow.cli"

	closer "github.com/CoreumFoundation/coreum-tools/pkg/closer"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/coremon/pkg/coremon"
)

func processCmd(c *cli.Cmd) {
	var (
		chainID                *string
		tendermintRPC          *string
		parallelBlockFetchJobs *int
	)

	initCosmosOptions(
		c,
		&chainID,
		&tendermintRPC,
		&parallelBlockFetchJobs,
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

		interfaceRegistry := cosmtypes.NewInterfaceRegistry()
		std.RegisterInterfaces(interfaceRegistry)
		protoCodec := codec.NewProtoCodec(interfaceRegistry)

		blockWatcher, err := coremon.NewTmBlockWatcher(
			rootCtx,
			*chainID,
			*tendermintRPC,
			protoCodec,
			*parallelBlockFetchJobs,
			coremon.NewBlockHandlerWithMetrics(rootCtx),
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
