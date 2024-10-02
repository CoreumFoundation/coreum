package integration

import (
	"bytes"
	"context"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmlightclienttypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
)

// ExecuteIBCTransfer executes IBC transfer transaction.
func (c ChainContext) ExecuteIBCTransfer(
	ctx context.Context,
	t *testing.T,
	txf client.Factory,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	t.Helper()

	return c.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		txf,
		senderAddress,
		coin,
		recipientChainContext,
		recipientChainContext.MustConvertToBech32Address(recipientAddress),
		"",
	)
}

// ExecuteIBCTransferWithMemo is similar to ExecuteIBCTransfer method
// but it allows passing memo and allows specifying the recipient as string.
func (c ChainContext) ExecuteIBCTransferWithMemo(
	ctx context.Context,
	t *testing.T,
	txf client.Factory,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress string,
	memo string,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := c.MustConvertToBech32Address(senderAddress)

	recipientChannelID := c.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		recipientChainContext,
	)
	height, err := c.GetLatestConsensusHeight(
		ctx,
		ibctransfertypes.PortID,
		recipientChannelID,
	)
	require.NoError(t, err)

	t.Logf("Sending IBC transfer sender: %s, receiver: %s, channel: %s amount: %s, memo: %s.",
		sender, recipientAddress, recipientChannelID, coin.String(), memo)
	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: recipientChannelID,
		Token:         coin,
		Sender:        sender,
		Receiver:      recipientAddress,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: height.RevisionNumber,
			RevisionHeight: height.RevisionHeight + 400000,
		},
		Memo: memo,
	}

	return c.BroadcastTxWithSigner(
		ctx,
		txf,
		senderAddress,
		&ibcSend,
	)
}

// ExecuteTimingOutIBCTransfer executes IBC transfer which should time out.
func (c ChainContext) ExecuteTimingOutIBCTransfer(
	ctx context.Context,
	t *testing.T,
	txf client.Factory,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := c.MustConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.MustConvertToBech32Address(recipientAddress)
	t.Logf("Sending timing out IBC transfer from %s, to %s, %s.", sender, receiver, coin.String())

	recipientChannelID := c.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		recipientChainContext,
	)

	tmQueryClient := cmtservice.NewServiceClient(recipientChainContext.ClientContext)
	latestBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	require.NoError(t, err)
	var headerTime time.Time
	if latestBlockRes.SdkBlock != nil {
		headerTime = latestBlockRes.GetSdkBlock().GetHeader().Time
	} else {
		headerTime = latestBlockRes.GetBlock().GetHeader().Time // we keep it to keep the compatibility with old versions
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
		txf,
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

	t.Logf(
		"Waiting for account %s balance, expected amount: %s.",
		c.MustConvertToBech32Address(address),
		expectedBalance.String(),
	)
	bankClient := banktypes.NewQueryClient(c.ClientContext)

	err := c.AwaitState(ctx, func(ctx context.Context) error {
		// We intentionally query all balances instead of single denom here to include this info inside error message.
		balancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
			Address: c.MustConvertToBech32Address(address),
		})
		if err != nil {
			return err
		}

		if balancesRes.Balances.AmountOf(expectedBalance.Denom).String() != expectedBalance.Amount.String() {
			return retry.Retryable(errors.Errorf(
				"balance of %s is not as expected, all balances: %s",
				expectedBalance.String(),
				balancesRes.Balances.String()),
			)
		}

		return nil
	})
	if err == nil {
		t.Logf("Received expected balance of %s.", expectedBalance.String())
	}

	return err
}

// AwaitForIBCChannelID returns the last opened channel of the IBC connected chain peer with specified port.
func (c ChainContext) AwaitForIBCChannelID(
	ctx context.Context,
	t *testing.T,
	port string,
	peerChain ChainContext,
) string {
	t.Helper()

	t.Logf("Getting %s chain channel with port %s on %s chain.",
		peerChain.ChainSettings.ChainID, port, c.ChainSettings.ChainID)

	var connectedChannelIDs []string

	err := c.AwaitState(ctx, func(ctx context.Context) error {
		// Reset slice in case previous iteration failed.
		connectedChannelIDs = []string{}

		openChannelsMap, err := c.getAllOpenChannels(ctx)
		if err != nil {
			return errors.Errorf("failed to query open channels on: %s: %s", c.ChainSettings.ChainID, err)
		}

		peerOpenChannelsMap, err := peerChain.getAllOpenChannels(ctx)
		if err != nil {
			return errors.Errorf("failed to query open channels on: %s: %s", peerChain.ChainSettings.ChainID, err)
		}

		for chID, ch := range openChannelsMap {
			if ch.PortId != port {
				continue
			}

			// Counterparty channel on a peer chain should exist and match a current chain channel.
			peerCh, ok := peerOpenChannelsMap[ch.Counterparty.ChannelId]
			if !ok || peerCh.Counterparty.ChannelId != chID {
				continue
			}
			// Peer chain might have different port ID. E.g., in case of IBC transfer from WASM smart contract
			// source port is wasm.<src-chain-smart-contract>, but destination is wasm.<dst-chain-smart-contract>.
			peerPort := peerCh.PortId

			expectedPeerChainName, err := c.getIBCCounterpartyChainName(ctx, chID, port)
			if err != nil {
				return errors.Wrapf(err, "counterparty chain name query failed for: %s", c.ChainSettings.ChainID)
			}
			expectedChainName, err := peerChain.getIBCCounterpartyChainName(ctx, peerCh.ChannelId, peerPort)
			if err != nil {
				return errors.Wrapf(err, "counterparty chain name query failed for: %s", peerChain.ChainSettings.ChainID)
			}

			// Chains names should match.
			if expectedChainName != c.ChainSettings.ChainID || expectedPeerChainName != peerChain.ChainSettings.ChainID {
				continue
			}
			connectedChannelIDs = append(connectedChannelIDs, ch.ChannelId)
		}

		if len(connectedChannelIDs) == 0 {
			return errors.New("no open channels found")
		}
		return nil
	})
	require.NoError(t, err)

	// Intentionally return channel with the last id because the last channel is more likely to be appropriate one
	// especially on devnet or testnet where channels are recreated frequently.
	sort.Slice(connectedChannelIDs, func(i, j int) bool {
		iChID, err := parseNumericChannelID(connectedChannelIDs[i])
		require.NoError(t, err)

		jChID, err := parseNumericChannelID(connectedChannelIDs[j])
		require.NoError(t, err)

		return iChID > jChID
	})

	t.Logf("Got %s chain channels on %s chain, channelIDs:%s. Using channelID: %s ",
		peerChain.ChainSettings.ChainID, c.ChainSettings.ChainID,
		strings.Join(connectedChannelIDs, ","), connectedChannelIDs[0])

	return connectedChannelIDs[0]
}

// GetLatestConsensusHeight returns the latest consensus height  for provided IBC port and channelID.
func (c ChainContext) GetLatestConsensusHeight(
	ctx context.Context,
	portID, channelID string,
) (ibcclienttypes.Height, error) {
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
	if err := c.ClientContext.InterfaceRegistry().UnpackAny(
		clientRes.IdentifiedClientState.ClientState,
		&clientState,
	); err != nil {
		return ibcclienttypes.Height{}, err
	}

	clientHeight, ok := clientState.GetLatestHeight().(ibcclienttypes.Height)
	if !ok {
		return ibcclienttypes.Height{},
			sdkerrors.Wrapf(
				cosmoserrors.ErrInvalidHeight,
				"invalid height type. expected type: %T, got: %T",
				ibcclienttypes.Height{},
				clientHeight,
			)
	}

	return clientHeight, nil
}

// AwaitForIBCClientAndConnectionIDs returns the clientID and channel for the peer chain.
func (c ChainContext) AwaitForIBCClientAndConnectionIDs(
	ctx context.Context,
	t *testing.T,
	peerChainID string,
) (string, string) {
	t.Helper()

	t.Logf(
		"Waiting for IBC client and connection for the chain %s, on the chain: %s.",
		peerChainID,
		c.ChainSettings.ChainID,
	)

	var (
		clientID, connectionID string
		err                    error
	)

	require.NoError(t, c.AwaitState(ctx, func(ctx context.Context) error {
		clientID, connectionID, err = c.getIBCClientAndConnectionIDs(ctx, peerChainID)
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
		Pagination: &query.PageRequest{Limit: query.PaginationMaxLimit},
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

func (c ChainContext) getAllOpenChannels(ctx context.Context) (map[string]*ibcchanneltypes.IdentifiedChannel, error) {
	ibcChannelClient := ibcchanneltypes.NewQueryClient(c.ClientContext)

	var openChannels []*ibcchanneltypes.IdentifiedChannel

	channelsPagination := &query.PageRequest{Limit: query.DefaultLimit}

	for {
		ibcChannelsRes, err := ibcChannelClient.Channels(
			ctx,
			&ibcchanneltypes.QueryChannelsRequest{Pagination: channelsPagination},
		)
		if err != nil {
			return nil, err
		}

		openChannelsBatch := lo.Filter(ibcChannelsRes.Channels, func(ch *ibcchanneltypes.IdentifiedChannel, _ int) bool {
			return ch.State == ibcchanneltypes.OPEN
		})
		openChannels = append(openChannels, openChannelsBatch...)

		if bytes.Equal(ibcChannelsRes.Pagination.NextKey, []byte("")) {
			break
		}
		channelsPagination.Key = ibcChannelsRes.Pagination.NextKey
	}

	openChannelsMap := lo.SliceToMap(openChannels,
		func(ch *ibcchanneltypes.IdentifiedChannel) (string, *ibcchanneltypes.IdentifiedChannel) {
			return ch.ChannelId, ch
		})

	return openChannelsMap, nil
}

func (c ChainContext) getIBCCounterpartyChainName(ctx context.Context, channelID, portID string) (string, error) {
	ibcChannelClient := ibcchanneltypes.NewQueryClient(c.ClientContext)

	channelClientStateRes, err := ibcChannelClient.ChannelClientState(
		ctx,
		&ibcchanneltypes.QueryChannelClientStateRequest{
			PortId:    portID,
			ChannelId: channelID,
		})
	if err != nil {
		return "", err
	}

	var clientState ibctmlightclienttypes.ClientState
	err = c.ClientContext.Codec().Unmarshal(channelClientStateRes.IdentifiedClientState.ClientState.Value, &clientState)
	if err != nil {
		return "", err
	}

	return clientState.ChainId, nil
}

func parseNumericChannelID(channelID string) (uint64, error) {
	chIDParts := strings.Split(channelID, "-")

	if len(chIDParts) != 2 || chIDParts[0] != "channel" {
		return 0, errors.Errorf("invalid channel ID: %s", channelID)
	}

	chIDNum, err := strconv.ParseUint(chIDParts[1], 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid channel ID: %s", channelID)
	}

	return chIDNum, nil
}
