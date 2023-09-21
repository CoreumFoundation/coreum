//go:build integrationtests

package ibc

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	cosmosrelayer "github.com/cosmos/relayer/v2/relayer"
	cosmosrelayercosmoschain "github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
)

// ConvertToIBCDenom returns the IBC denom based on the channelID and denom.
func ConvertToIBCDenom(channelID, denom string) string {
	return ibctransfertypes.ParseDenomTrace(
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}

// CreateIBCChannelsAndConnect creates two new channels for the provided ports on provided chains and connects them.
func CreateIBCChannelsAndConnect(
	ctx context.Context,
	t *testing.T,
	srcChain integration.Chain,
	srcChainPort string,
	dstChain integration.Chain,
	dstChainPort string,
	channelVersion string,
	channelOrder ibcchanneltypes.Order,
) func() {
	t.Helper()

	log := zaptest.NewLogger(t)

	const relayerKeyName = "relayer-key"

	srcClientID, srcConnectionID := srcChain.AwaitForIBCClientAndConnectionIDs(ctx, t, dstChain.ChainSettings.ChainID)
	relayerSrcChain := setupRelayerChain(ctx, t, log, srcChain, relayerKeyName, srcClientID, srcConnectionID)

	dstClientID, dstConnectionID := dstChain.AwaitForIBCClientAndConnectionIDs(ctx, t, srcChain.ChainSettings.ChainID)
	relayerDstChain := setupRelayerChain(ctx, t, log, dstChain, relayerKeyName, dstClientID, dstConnectionID)

	var channelOrderString string
	switch channelOrder {
	case ibcchanneltypes.UNORDERED:
		channelOrderString = "UNORDERED"
	case ibcchanneltypes.ORDERED:
		channelOrderString = "ORDERED"
	default:
		t.Fatalf("Unsupported chennel order type:%d", channelOrder)
	}

	pathName := fmt.Sprintf("%s-%s", srcChain.ChainSettings.ChainID, dstChain.ChainSettings.ChainID)
	require.NoError(t, relayerSrcChain.CreateOpenChannels(
		ctx,
		relayerDstChain,
		3,
		5*time.Second,
		srcChainPort, dstChainPort,
		channelOrderString, channelVersion,
		false,
		"",
		pathName,
	))
	closerFunc := func() {
		require.NoError(t, relayerSrcChain.CloseChannel(ctx, relayerDstChain, 5, 5*time.Second, srcChain.ChainSettings.ChainID, srcChainPort, "", pathName))
	}
	return closerFunc
}

func setupRelayerChain(
	ctx context.Context,
	t *testing.T,
	log *zap.Logger,
	chain integration.Chain,
	relayerKeyName string,
	clientID, connectionID string,
) *cosmosrelayer.Chain {
	t.Helper()

	relayerSrcChainConfig := cosmosrelayercosmoschain.CosmosProviderConfig{
		Key:            relayerKeyName,
		ChainName:      chain.ChainSettings.ChainID,
		ChainID:        chain.ChainSettings.ChainID,
		RPCAddr:        chain.ChainSettings.RPCAddress,
		AccountPrefix:  chain.ChainSettings.AddressPrefix,
		KeyringBackend: "test",
		GasAdjustment:  1.2,
		GasPrices:      fmt.Sprintf("%s%s", chain.ChainSettings.GasPrice, chain.ChainSettings.Denom),
		Debug:          false,
		Timeout:        "20s",
		OutputFormat:   "indent",
		SignModeStr:    "direct",
	}

	relayerSrcChainProvider, err := relayerSrcChainConfig.NewProvider(log, t.TempDir(), false, chain.ChainSettings.ChainID)
	require.NoError(t, err)
	require.NoError(t, relayerSrcChainProvider.Init(ctx))
	relayerSrcChainKeyInfo, err := relayerSrcChainProvider.AddKey(relayerKeyName, chain.ChainSettings.CoinType, string(hd.Secp256k1Type))
	require.NoError(t, err)
	_, relayerKeyBytes, err := bech32.DecodeAndConvert(relayerSrcChainKeyInfo.Address)
	require.NoError(t, err)

	chain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: relayerKeyBytes,
		Amount:  chain.NewCoin(sdkmath.NewInt(2000000)),
	})

	relayerChain := cosmosrelayer.NewChain(log, relayerSrcChainProvider, false)
	relayerChain.PathEnd = &cosmosrelayer.PathEnd{
		ChainID:      relayerChain.ChainID(),
		ClientID:     clientID,
		ConnectionID: connectionID,
	}
	return relayerChain
}
