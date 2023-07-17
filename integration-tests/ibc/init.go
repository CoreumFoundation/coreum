//go:build integrationtests

package ibc

import (
	"context"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

func init() {
	ctx := context.Background()
	queryCtx, queryCtxCancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
	defer queryCtxCancel()

	// ********** Gaia **********
	gaiaGRPClient, err := grpc.Dial(integrationtests.GaiaGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaSettings := integrationtests.QueryCommonSettings(queryCtx, gaiaGRPClient)
	gaiaSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	gaiaSettings.GasAdjustment = 1.5
	gaiaSettings.CoinType = sdk.CoinType // gaia coin type
	gaiaSettings.RPCAddress = integrationtests.GaiaRPCAddress

	gaiaRPClient, err := sdkclient.NewClientFromNode(integrationtests.GaiaRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaChain := integrationtests.NewChain(
		gaiaGRPClient,
		gaiaRPClient,
		gaiaSettings,
		integrationtests.GaiaFundingMnemonic)
	integrationtests.ChainsHolder.Gaia = gaiaChain
	// ********** Osmosis **********

	osmosisGRPClient, err := grpc.Dial(integrationtests.OsmosisGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChainSettings := integrationtests.QueryCommonSettings(queryCtx, osmosisGRPClient)
	osmosisChainSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	osmosisChainSettings.GasAdjustment = 1.5
	osmosisChainSettings.CoinType = sdk.CoinType // osmosis coin type
	osmosisChainSettings.RPCAddress = integrationtests.OsmosisRPCAddress

	osmosisRPClient, err := sdkclient.NewClientFromNode(integrationtests.OsmosisRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChain := integrationtests.NewChain(
		osmosisGRPClient,
		osmosisRPClient,
		osmosisChainSettings,
		integrationtests.OsmosisFundingMnemonic)
	integrationtests.ChainsHolder.Osmosis = osmosisChain
}
