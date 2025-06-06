package cosmoscmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/CoreumFoundation/coreum/v6/app"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

// OverwriteDefaultChainIDFlags searches for the DefaultChainID flag and replaces its value of the current default.
func OverwriteDefaultChainIDFlags(parentCmd *cobra.Command) {
	for _, cmd := range parentCmd.Commands() {
		if flag := cmd.LocalFlags().Lookup(flags.FlagChainID); flag != nil {
			flag.DefValue = string(app.DefaultChainID)
		}

		OverwriteDefaultChainIDFlags(cmd)
	}
}

// PreProcessFlags prepares the initial flags config for the cli.
func PreProcessFlags() (config.NetworkConfig, error) {
	// define flags
	const flagHelp = "help"
	flagSet := pflag.NewFlagSet("pre-process", pflag.ExitOnError)
	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	flagSet.String(flags.FlagHome, app.DefaultNodeHome, "Directory for config and data")
	// Dummy flag to turn off printing usage of this flag set
	help := flagSet.BoolP(flagHelp, "h", false, "")
	chainID := flagSet.String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
	//nolint:errcheck // since we have set ExitOnError on flagset, we don't need to check for errors here
	flagSet.Parse(os.Args[1:])
	// get chain config
	network, err := config.NetworkConfigByChainID(constant.ChainID(*chainID))
	if err != nil {
		return config.NetworkConfig{}, err
	}
	// we consider the issued command to be a help command if no args are provided.
	// in that case we will not check the chain-id and will return
	if len(os.Args) == 1 || *help {
		return network, nil
	}

	// overwrite home flag
	if flagSet.Changed(flags.FlagHome) {
		err := appendStringFlag(os.Args, flags.FlagHome, *chainID)
		if err != nil {
			return config.NetworkConfig{}, err
		}
	} else {
		appendedHome := filepath.Join(app.DefaultNodeHome, *chainID)
		os.Args = append(os.Args, fmt.Sprintf("--%s=%s", flags.FlagHome, appendedHome))
	}

	return network, nil
}

func appendStringFlag(args []string, flag, newVal string) error {
	for pos, arg := range args {
		if arg != "--"+flag && !strings.HasPrefix(arg, "--"+flag+"=") {
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

func removeFlag(args []string, flag string) []string {
	newArgs := make([]string, 0, len(args))
	var prevArgRemoved bool
	for _, arg := range args {
		if arg == flag {
			prevArgRemoved = true
			continue
		}
		if prevArgRemoved {
			prevArgRemoved = false
			if !strings.HasPrefix(arg, "-") {
				continue
			}
		}
		if strings.HasPrefix(arg, flag+"=") {
			continue
		}

		newArgs = append(newArgs, arg)
	}

	return newArgs
}
