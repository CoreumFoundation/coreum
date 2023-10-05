package integration

import (
	"context"
	"testing"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmlightclienttypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

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

	sender := c.MustConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.MustConvertToBech32Address(recipientAddress)
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

// ExecuteTimingOutIBCTransfer executes IBC transfer which should time out.
func (c ChainContext) ExecuteTimingOutIBCTransfer(
	ctx context.Context,
	t *testing.T,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := c.MustConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.MustConvertToBech32Address(recipientAddress)
	t.Logf("Sending timing out IBC transfer from %s, to %s, %s.", sender, receiver, coin.String())

	recipientChannelID := c.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, recipientChainContext.ChainSettings.ChainID)

	tmQueryClient := tmservice.NewServiceClient(recipientChainContext.ClientContext)
	latestBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	require.NoError(t, err)
	var headerTime time.Time
	if latestBlockRes.SdkBlock != nil {
		headerTime = latestBlockRes.GetSdkBlock().GetHeader().Time
	} else {
		// TODO: remove this "if condition" once all the connected chains have migrated to cosmos sdk v0.47.
		// Block is deprecated in favor of SdkBlock.
		headerTime = latestBlockRes.GetBlock().GetHeader().Time
	}

	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:       ibctransfertypes.PortID,
		SourceChannel:    recipientChannelID,
		Token:            coin,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutTimestamp: uint64(headerTime.Add(-5 * time.Second).UnixNano()),
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
) error {
	t.Helper()

	t.Logf("Waiting for account %s balance, expected amount: %s.", c.MustConvertToBech32Address(address), expectedBalance.String())
	bankClient := banktypes.NewQueryClient(c.ClientContext)
	retryCtx, retryCancel := context.WithTimeout(ctx, 30*time.Second)
	defer retryCancel()
	err := retry.Do(retryCtx, 100*time.Millisecond, func() error {
		requestCtx, requestCancel := context.WithTimeout(retryCtx, 5*time.Second)
		defer requestCancel()

		// We intentionally query all balances instead of single denom here to include this info inside error message.
		balancesRes, err := bankClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{
			Address: c.MustConvertToBech32Address(address),
		})
		if err != nil {
			return err
		}

		if balancesRes.Balances.AmountOf(expectedBalance.Denom).String() != expectedBalance.Amount.String() {
			return retry.Retryable(errors.Errorf("balance of %s is not as expected, all balances: %s", expectedBalance.String(), balancesRes.Balances.String()))
		}

		return nil
	})
	if err == nil {
		t.Logf("Received expected balance of %s.", expectedBalance.String())
	}

	return err
}

// AwaitForIBCChannelID returns the first opened channel of the IBC connected chain peer.
func (c ChainContext) AwaitForIBCChannelID(ctx context.Context, t *testing.T, port, peerChainID string) string {
	t.Helper()

	t.Logf("Getting %s chain channel with port %s on %s chain.", peerChainID, port, c.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer retryCancel()

	ibcChannelClient := ibcchanneltypes.NewQueryClient(c.ClientContext)

	var channelID string
	require.NoError(t, retry.Do(retryCtx, 500*time.Millisecond, func() error {
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
		return ibcclienttypes.Height{}, sdkerrors.Wrapf(cosmoserrors.ErrInvalidHeight, "invalid height type. expected type: %T, got: %T",
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

	require.NoError(t, retry.Do(retryCtx, 500*time.Millisecond, func() error {
		clientID, connectionID, err = c.getIBCClientAndConnectionIDs(retryCtx, peerChainID)
		if err != nil {
			return retry.Retryable(errors.Errorf("client and connection are not ready yet, %s", err))
		}
		return nil
	}))

	return clientID, connectionID
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
			return "", "", errors.Errorf("failed to find client %s connection on the chain %s", clientID, peerChainID)
		}

		return clientID, channelsRes.ConnectionPaths[0], nil
	}

	return "", "", errors.Errorf("failed to find client and connection on the %s", peerChainID)
}
