package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cored/pkg/config"
)

func main() {
	net, err := config.NetworkByChainID(string(config.Mainnet))
	if err != nil {
		panic(err)
	}
	rootCmd, _ := cosmoscmd.NewRootCmd(
		app.Name,
		net.AddressPrefix,
		app.DefaultNodeHome,
		app.Name,
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
	)
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
