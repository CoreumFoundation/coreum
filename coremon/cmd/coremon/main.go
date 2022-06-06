package main

import (
	"context"
	"os"

	cli "github.com/jawher/mow.cli"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

var app = cli.App("coremon", "Daemon for cosmos chain monitoring and accurate stats exporting.")

// Global options for the app
var (
	envName *string

	appLogger, _          = zap.NewProduction()
	rootCtx, rootCancelFn = context.WithCancel(context.Background())
)

func main() {
	// Allows to set env variables from .env file
	readEnv()

	initGlobalOptions(&envName)
	initLogger(envName, &appLogger)

	app.Before = prepareApp
	app.Command("process", "Start chain blocks processing", processCmd)

	_ = app.Run(os.Args)
}

func prepareApp() {
	rootCtx = logger.WithLogger(rootCtx, appLogger)
}
