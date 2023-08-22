package main

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/cmd/cored/cosmoscmd"
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
		var errCode server.ErrorCode
		if errors.As(err, errCode) {
			os.Exit(errCode.Code)
		}

		os.Exit(1)
	}
}
