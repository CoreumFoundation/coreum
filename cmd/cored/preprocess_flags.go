package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

func preProcessFlags() (app.Network, error) {
	// define flags
	const flagHelp = "help"
	flagSet := pflag.NewFlagSet("pre-process", pflag.ExitOnError)
	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	flagSet.String(flags.FlagHome, app.DefaultNodeHome, "Directory for config and data")
	// Dummy flag to turn off printing usage of this flag set
	flagSet.BoolP(flagHelp, "h", false, "")
	chainID := flagSet.String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
	//nolint:errcheck // since we have set ExitOnError on flagset, we don't need to check for errors here
	flagSet.Parse(os.Args[1:])
	var shouldPrintHelp bool
	// we consider the issued command to be a help command. in that case we will ignore if
	// the network is disabled
	// TODO: remove this check after all chains are enabled.
	if flagSet.Changed(flagHelp) || len(os.Args) == 1 {
		shouldPrintHelp = true
	}

	// get chain config
	network, err := app.NetworkByChainID(app.ChainID(*chainID))
	// skip checking network is disabled error if help must be printed.
	// this is introduced only because some chains are disabled.
	// TODO: remove this check after all chains are enabled.
	if ignoreErr := errors.Is(err, app.ErrDisableNetwork) && shouldPrintHelp; err != nil && !ignoreErr {
		return app.Network{}, err
	}

	app.ChosenNetwork = network

	// overwrite home flag
	if flagSet.Changed(flags.FlagHome) {
		err = appendStringFlag(os.Args, flags.FlagHome, *chainID)
		if err != nil {
			return app.Network{}, err
		}
	} else {
		appendedHome := filepath.Join(app.DefaultNodeHome, *chainID)
		os.Args = append(os.Args, fmt.Sprintf("--%s=%s", flags.FlagHome, appendedHome))
	}

	return network, nil
}

func appendStringFlag(args []string, flag string, newVal string) error {
	for pos, arg := range args {
		if !(arg == "--"+flag || strings.HasPrefix(arg, "--"+flag+"=")) {
			continue
		}

		splits := strings.SplitN(arg, "=", 2)
		if len(splits) == 2 {
			args[pos] = splits[0] + "=" + filepath.Join(splits[1], newVal)
			return nil
		}

		if pos+1 > len(args) {
			return errors.Errorf("missing arg value for flag %s", flag)
		}

		args[pos+1] = filepath.Join(args[pos+1], newVal)
		return nil
	}

	return errors.New("flag not found")
}
