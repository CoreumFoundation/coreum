//go:build integrationtests

package ibc

import (
	"context"
	"flag"
	"os"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

var (
	gaiaGRPCAddress     string
	gaiaRPCAddress      string
	gaiaFundingMnemonic string

	osmosisGRPCAddress     string
	osmosisRPCAddress      string
	osmosisFundingMnemonic string
)

func parseIBCFlags() {
	flagSet := flag.NewFlagSet("ibc", flag.ExitOnError)
	flagSet.StringVar(&gaiaGRPCAddress, "gaia-grpc-address", "localhost:9080", "GRPC address of gaia node started by znet")
	flagSet.StringVar(&gaiaRPCAddress, "gaia-rpc-address", "http://localhost:26557", "RPC address of gaia node started by znet")
	flagSet.StringVar(&gaiaFundingMnemonic, "gaia-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flagSet.StringVar(&osmosisGRPCAddress, "osmosis-grpc-address", "localhost:9070", "GRPC address of osmosis node started by znet")
	flagSet.StringVar(&osmosisRPCAddress, "osmosis-rpc-address", "http://localhost:26457", "RPC address of osmosis node started by znet")
	flagSet.StringVar(&osmosisFundingMnemonic, "osmosis-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flagSet.Parse(os.Args)
}

func init() {
	parseIBCFlags()
	ctx := context.Background()
	queryCtx, queryCtxCancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
	defer queryCtxCancel()

	// ********** Gaia **********
	gaiaGRPClient, err := grpc.Dial(gaiaGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaSettings := integrationtests.QueryCommonSettings(queryCtx, gaiaGRPClient)
	gaiaSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	gaiaSettings.GasAdjustment = 1.5
	gaiaSettings.CoinType = sdk.CoinType // gaia coin type
	gaiaSettings.RPCAddress = gaiaRPCAddress

	gaiaRPClient, err := sdkclient.NewClientFromNode(gaiaRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaChain := integrationtests.NewChain(
		gaiaGRPClient,
		gaiaRPClient,
		gaiaSettings,
		gaiaFundingMnemonic)
	integrationtests.ChainsHolder.Gaia = gaiaChain
	// ********** Osmosis **********

	osmosisGRPClient, err := grpc.Dial(osmosisGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChainSettings := integrationtests.QueryCommonSettings(queryCtx, osmosisGRPClient)
	osmosisChainSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	osmosisChainSettings.GasAdjustment = 1.5
	osmosisChainSettings.CoinType = sdk.CoinType // osmosis coin type
	osmosisChainSettings.RPCAddress = osmosisRPCAddress

	osmosisRPClient, err := sdkclient.NewClientFromNode(osmosisRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChain := integrationtests.NewChain(
		osmosisGRPClient,
		osmosisRPClient,
		osmosisChainSettings,
		osmosisFundingMnemonic)
	integrationtests.ChainsHolder.Osmosis = osmosisChain
}
