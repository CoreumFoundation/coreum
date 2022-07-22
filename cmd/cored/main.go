package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cmd/cored/cosmoscmd"
)

func main() {
	logger := server.ZeroLogWrapper{Logger: log.Logger}
	network, err := preProcessChainIDFlag()
	if err != nil {
		logger.Error("error processing chain id flag", "err", err)
		os.Exit(1)
	}
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		network.AddressPrefix(),
		app.DefaultNodeHome,
		string(network.ChainID()),
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
	)

	rootCmd.AddCommand(initCmd(app.DefaultNodeHome))

	for _, cmd := range rootCmd.Commands() {
		if isStringInList(cmd.Name(), "start", "collect-gentxs") {
			cmd.PersistentFlags().String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
		}

		// error out if the start command tries to connect to Mainnet, since it is not yet ready.
		if isStringInList(cmd.Name(), "start", "init") {
			cmd.PreRunE = chainCobraRunE(checkChainIDNotMain, cmd.PreRunE)
		}
	}
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}

func isStringInList(str string, list ...string) bool {
	for _, l := range list {
		if str == l {
			return true
		}
	}
	return false
}

func checkChainIDNotMain(cmd *cobra.Command, args []string) error {
	chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	if chainID == string(app.Mainnet) {
		return errors.Errorf("mainnet is not yet ready, use --chain-id=%s for devnet", string(app.Devnet))
	}

	return nil
}

func chainCobraRunE(list ...func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range list {
			if fn != nil {
				err := fn(cmd, args)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}
