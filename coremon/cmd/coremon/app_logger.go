package main

import (
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
)

func initLogger(env *string, appLogger **zap.Logger) {
	var (
		logFormat  *string
		logVerbose *string
	)

	initLoggerOptions(
		&logFormat,
		&logVerbose,
	)

	*appLogger = logger.New(logger.Config{
		Format:  logger.Format(*logFormat),
		Verbose: toBool(*logVerbose),
	}).With(zap.String("env", *env))
}
