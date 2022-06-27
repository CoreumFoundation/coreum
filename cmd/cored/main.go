package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cored/pkg/config"
)

func main() {
	network, _ := config.GetNetworkByChainID(string(config.Devnet))
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		network.AddressPrefix,
		app.DefaultNodeHome,
		string(network.ChainID),
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
	)
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
