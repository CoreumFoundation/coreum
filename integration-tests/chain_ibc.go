package integrationtests

import (
	"context"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	ibctmlightclienttypes "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	cosmosrelayer "github.com/cosmos/relayer/v2/relayer"
	cosmosrelayercosmoschain "github.com/cosmos/relayer/v2/relayer/chains/cosmos"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
)

// ExecuteIBCTransfer executes IBC transfer transaction.
func (c ChainContext) ExecuteIBCTransfer(
	ctx context.Context,
	t *testing.T,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := c.ConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.ConvertToBech32Address(recipientAddress)
	t.Logf("Sending IBC transfer sender: %s, receiver: %s, amount: %s.", sender, receiver, coin.String())

	recipientChannelID := c.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, recipientChainContext.ChainSettings.ChainID)
	height, err := c.GetLatestConsensusHeight(
		ctx,
		ibctransfertypes.PortID,
		recipientChannelID,
	)
	require.NoError(t, err)

	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: recipientChannelID,
		Token:         coin,
		Sender:        sender,
		Receiver:      receiver,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: height.RevisionNumber,
			RevisionHeight: height.RevisionHeight + 1000,
		},
	}

	return c.BroadcastTxWithSigner(
		ctx,
		c.TxFactory().WithSimulateAndExecute(true),
		senderAddress,
		&ibcSend,
	)
}

// AwaitForBalance queries for the balance with retry and timeout.
func (c ChainContext) AwaitForBalance(
	ctx context.Context,
	t *testing.T,
	address sdk.AccAddress,
	expectedBalance sdk.Coin,
) {
	t.Helper()

	t.Logf("Waiting for account %s balance, expected amount: %s.", c.ConvertToBech32Address(address), expectedBalance.String())
	bankClient := banktypes.NewQueryClient(c.ClientContext)
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	require.NoError(t, retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(retryCtx, 5*time.Second)
		defer requestCancel()

		// We intentionally query all balances instead of single denom here to include this info inside error message.
		balancesRes, err := bankClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{
			Address: c.ConvertToBech32Address(address),
		})
		if err != nil {
			return err
		}

		if balancesRes.Balances.AmountOf(expectedBalance.Denom).String() != expectedBalance.Amount.String() {
			return retry.Retryable(errors.Errorf("%s balance is still not equal to expected, all balances: %s", expectedBalance.Denom, balancesRes.String()))
		}

		return nil
	}))

	t.Logf("Received expected balance of %s.", expectedBalance.Denom)
}

// AwaitForIBCChannelID returns the first opened channel of the IBC connected chain peer.
func (c ChainContext) AwaitForIBCChannelID(ctx context.Context, t *testing.T, port, peerChainID string) string {
	t.Helper()

	t.Logf("Getting %s chain channel with port %s on %s chain.", peerChainID, port, c.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer retryCancel()

	ibcChannelClient := ibcchanneltypes.NewQueryClient(c.ClientContext)

	var channelID string
	require.NoError(t, retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
		defer requestCancel()

		ibcChannelsRes, err := ibcChannelClient.Channels(requestCtx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range ibcChannelsRes.Channels {
			if ch.PortId != port || ch.State != ibcchanneltypes.OPEN {
				continue
			}

			channelClientStateRes, err := ibcChannelClient.ChannelClientState(requestCtx, &ibcchanneltypes.QueryChannelClientStateRequest{
				PortId:    ch.PortId,
				ChannelId: ch.ChannelId,
			})
			if err != nil {
				return err
			}

			var clientState ibctmlightclienttypes.ClientState
			err = c.ClientContext.Codec().Unmarshal(channelClientStateRes.IdentifiedClientState.ClientState.Value, &clientState)
			if err != nil {
				return err
			}

			if clientState.ChainId == peerChainID {
				channelID = ch.ChannelId
				return nil
			}
		}

		return retry.Retryable(errors.Errorf("waiting for the %s channel on the %s to open", peerChainID, c.ChainSettings.ChainID))
	}))

	t.Logf("Got %s chain channel on %s chain, channelID:%s ", peerChainID, c.ChainSettings.ChainID, channelID)

	return channelID
}

// GetLatestConsensusHeight returns the latest consensus height  for provided IBC port and channelID.
func (c ChainContext) GetLatestConsensusHeight(ctx context.Context, portID, channelID string) (ibcclienttypes.Height, error) {
	queryClient := ibcchanneltypes.NewQueryClient(c.ClientContext)
	req := &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	clientRes, err := queryClient.ChannelClientState(ctx, req)
	if err != nil {
		return ibcclienttypes.Height{}, err
	}

	var clientState exported.ClientState
	if err := c.ClientContext.InterfaceRegistry().UnpackAny(clientRes.IdentifiedClientState.ClientState, &clientState); err != nil {
		return ibcclienttypes.Height{}, err
	}

	clientHeight, ok := clientState.GetLatestHeight().(ibcclienttypes.Height)
	if !ok {
		return ibcclienttypes.Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid height type. expected type: %T, got: %T",
			ibcclienttypes.Height{}, clientHeight)
	}

	return clientHeight, nil
}

// AwaitForIBCClientAndConnectionIDs returns the clientID and channel for the peer chain.
func (c ChainContext) AwaitForIBCClientAndConnectionIDs(ctx context.Context, t *testing.T, peerChainID string) (string, string) {
	t.Helper()

	t.Logf("Waiting for IBC client and connection for the chain %s, on the chain: %s.", peerChainID, c.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	var (
		clientID, connectionID string
		err                    error
	)

	require.NoError(t, retry.Do(retryCtx, time.Second, func() error {
		clientID, connectionID, err = c.getIBCClientAndConnectionIDs(retryCtx, peerChainID)
		if err != nil {
			return retry.Retryable(errors.Errorf("client and connection are not ready yet, %s", err))
		}
		return nil
	}))

	return clientID, connectionID
}

// CreateIBCChannelsAndConnect creates two new channels for the provided ports on provided chains and connects them.
func CreateIBCChannelsAndConnect(
	ctx context.Context,
	t *testing.T,
	srcChain Chain,
	srcChainPort string,
	dstChain Chain,
	dstChainPort string,
	channelVersion string,
	channelOrder ibcchanneltypes.Order,
) {
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

	require.NoError(t, relayerSrcChain.CreateOpenChannels(
		ctx,
		relayerDstChain,
		3,
		5*time.Second,
		srcChainPort, dstChainPort,
		channelOrderString, channelVersion,
		false,
		"",
	))
}

func setupRelayerChain(
	ctx context.Context,
	t *testing.T,
	log *zap.Logger,
	chain Chain,
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
	relayerSrcChainKeyInfo, err := relayerSrcChainProvider.AddKey(relayerKeyName, chain.ChainSettings.CoinType)
	require.NoError(t, err)
	_, relayerKeyBytes, err := bech32.DecodeAndConvert(relayerSrcChainKeyInfo.Address)
	require.NoError(t, err)

	chain.Faucet.FundAccounts(ctx, t, FundedAccount{
		Address: relayerKeyBytes,
		Amount:  chain.NewCoin(sdk.NewInt(2000000)),
	})

	relayerChain := cosmosrelayer.NewChain(log, relayerSrcChainProvider, false)
	relayerChain.PathEnd = &cosmosrelayer.PathEnd{
		ChainID:      relayerChain.ChainID(),
		ClientID:     clientID,
		ConnectionID: connectionID,
	}
	return relayerChain
}

func (c ChainContext) getIBCClientAndConnectionIDs(ctx context.Context, peerChainID string) (string, string, error) {
	ibcClientClient := ibcclienttypes.NewQueryClient(c.ClientContext)
	ibcChannelClient := ibcconnectiontypes.NewQueryClient(c.ClientContext)

	clientStatesRes, err := ibcClientClient.ClientStates(ctx, &ibcclienttypes.QueryClientStatesRequest{
		Pagination: &query.PageRequest{Limit: query.MaxLimit},
	})
	if err != nil {
		return "", "", err
	}

	for i := range clientStatesRes.ClientStates {
		var clientState ibctmlightclienttypes.ClientState
		err = c.ClientContext.Codec().Unmarshal(clientStatesRes.ClientStates[i].ClientState.Value, &clientState)
		if err != nil {
			return "", "", err
		}

		if clientState.ChainId != peerChainID {
			continue
		}

		clientID := clientStatesRes.ClientStates[i].ClientId
		channelsRes, err := ibcChannelClient.ClientConnections(ctx, &ibcconnectiontypes.QueryClientConnectionsRequest{
			ClientId: clientID,
		})
		if err != nil {
			return "", "", err
		}
		if len(channelsRes.ConnectionPaths) != 1 {
			return "", "", errors.Errorf("can't find client %s connection on the chain %s", clientID, peerChainID)
		}

		return clientID, channelsRes.ConnectionPaths[0], nil
	}

	return "", "", errors.Errorf("can't find client and connection on the %s", peerChainID)
}
