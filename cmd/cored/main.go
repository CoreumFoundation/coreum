package main

import (
	"os"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/cmd/cored/cosmoscmd"
)

func main() {
	logger := log.NewNopLogger()
	network, err := cosmoscmd.PreProcessFlags()
	if err != nil {
		logger.Error("Error processing chain id flag", "err", err)
		os.Exit(1)
	}
	network.SetSDKConfig()
	app.ChosenNetwork = network

	rootCmd := cosmoscmd.NewRootCmd()
	cosmoscmd.OverwriteDefaultChainIDFlags(rootCmd)
	rootCmd.PersistentFlags().String(flags.FlagChainID, string(app.DefaultChainID), "The network chain ID")
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		var errCode server.ErrorCode
		if errors.As(err, errCode) {
			os.Exit(errCode.Code)
		}

		os.Exit(1)
	}
}
