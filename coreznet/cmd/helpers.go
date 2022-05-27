package cmd

import (
	"os"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/spf13/pflag"
)

// ConfigureLoggerWithCLI configures logger based on CLI flags
// FIXME (wojtek): Move it to logger library
func ConfigureLoggerWithCLI(verboseDefault bool) {
	verbose := verboseDefault
	if len(os.Args) > 1 {
		flags := pflag.NewFlagSet("verbose", pflag.ContinueOnError)
		flags.ParseErrorsWhitelist.UnknownFlags = true
		flags.BoolVarP(&verbose, "verbose", "v", verbose, "Turns on verbose logging")
		// Dummy flag to turn off printing usage of this flag set
		flags.BoolP("help", "h", false, "")

		_ = flags.Parse(os.Args[1:])
	}

	if !verbose {
		logger.VerboseOff()
	}
}
