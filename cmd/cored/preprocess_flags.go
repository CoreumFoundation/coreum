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
	flagSet := pflag.NewFlagSet("pre-process", pflag.ExitOnError)
	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	flagSet.String(flags.FlagHome, app.DefaultNodeHome, "Directory for config and data")
	chainID := flagSet.String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
	//nolint:errcheck // since we have set ExitOnError on flagset, we don't need to check for errors here
	flagSet.Parse(os.Args[1:])

	// get chain config
	network, err := app.NetworkByChainID(app.ChainID(*chainID))
	if err != nil {
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
