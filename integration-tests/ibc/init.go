package ibc

import (
	"context"
	"flag"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

type IBCChains struct {
	Coreum  IBCChain
	Gaia    IBCChain
	Osmosis IBCChain
}

var (
	chains IBCChains
)

func InitChain(ctx context.Context, address, fundingMnemonic string) integrationtests.Chain {
	queryCtx, queryCtxCancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
	defer queryCtxCancel()

	gaiaGRPClient, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaChainSettings := integrationtests.QueryCommonSettings(queryCtx, gaiaGRPClient)
	gaiaChainSettings.GasPrice = sdk.ZeroDec()
	gaiaChainSettings.GasAdjustment = 1.3
	gaiaChainSettings.CoinType = sdk.CoinType // gaia coin type

	return integrationtests.NewChain(
		gaiaGRPClient,
		gaiaChainSettings,
		fundingMnemonic)
}

func init() {
	var (
		gaiaAddress         string
		gaiaFundingMnemonic string

		osmosisAddress         string
		osmosisFundingMnemonic string
	)

	flag.StringVar(&gaiaAddress, "gaia-address", "localhost:9080", "Address of gaia node started by znet")
	flag.StringVar(&gaiaFundingMnemonic, "gaia-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flag.StringVar(&osmosisAddress, "osmosis-address", "localhost:9070", "Address of osmosis node started by znet")
	flag.StringVar(&osmosisFundingMnemonic, "osmosis-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")

	integrationtests.Init()

	gaiaChain := InitChain(context.Background(), gaiaAddress, gaiaFundingMnemonic)
	osmosisChain := InitChain(context.Background(), osmosisAddress, osmosisFundingMnemonic)

	chains = IBCChains{
		Gaia:    IBCChain{gaiaChain},
		Osmosis: IBCChain{osmosisChain},
	}
}
