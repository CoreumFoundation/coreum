package main

import (
	"fmt"
	"os"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cmd/cored/cosmoscmd"
)

const coreumEnvPrefix = "CORED"

func main() {
	network, err := cosmoscmd.PreProcessFlags()
	if err != nil {
		fmt.Printf("Error processing chain id flag, err: %s", err)
		os.Exit(1)
	}
	network.SetSDKConfig()
	app.ChosenNetwork = network

	rootCmd := cosmoscmd.NewRootCmd()
	cosmoscmd.OverwriteDefaultChainIDFlags(rootCmd)
	rootCmd.PersistentFlags().String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
	if err := svrcmd.Execute(rootCmd, coreumEnvPrefix, app.DefaultNodeHome); err != nil {
		fmt.Printf("Error executing cmd, err: %s", err)
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}
