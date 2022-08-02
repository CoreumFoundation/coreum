package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/rs/zerolog/log"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cmd/cored/cosmoscmd"
)

func main() {
	logger := server.ZeroLogWrapper{Logger: log.Logger}
	network, err := preProcessFlags()
	if err != nil {
		logger.Error("Error processing chain id flag", "err", err)
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
		if cmd.Name() == "start" {
			cmd.PersistentFlags().String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
		}
	}
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
