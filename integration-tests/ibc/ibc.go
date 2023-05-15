//go:build integrationtests

package ibc

import (
	"context"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	ibctmlightclient "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// ChannelsConfig defines the config required for the IBC tests.
type ChannelsConfig struct {
	CoreumToGaiaChannelID string
	GaiaToCoreumChannelID string
}

// Ready returns the true the config is fully ready.
func (c ChannelsConfig) Ready() bool {
	return c.CoreumToGaiaChannelID != "" && c.GaiaToCoreumChannelID != ""
}

// AwaitForIBCConfig await for the IBC channels to be opened and returns them.
// TODO(milad): remove the await after we build this logic into crust.
func AwaitForIBCConfig(t *testing.T) ChannelsConfig {
	ctx, chains := integrationtests.NewChainsTestingContext(t)

	var ibcConfig ChannelsConfig
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	err := retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
		defer requestCancel()

		coreumIBCChannels, err := getOpenChannelIDs(requestCtx, chains.Coreum.ClientContext)
		if err != nil {
			return err
		}
		ibcConfig.CoreumToGaiaChannelID = coreumIBCChannels[chains.Gaia.ChainSettings.ChainID]

		gaiaIBCChannels, err := getOpenChannelIDs(requestCtx, chains.Gaia.ClientContext)
		if err != nil {
			return err
		}
		ibcConfig.GaiaToCoreumChannelID = gaiaIBCChannels[chains.Coreum.ChainSettings.ChainID]

		if ibcConfig.Ready() {
			return nil
		}

		return retry.Retryable(errors.New("waiting for channels to open"))
	})
	require.NoError(t, err)

	return ibcConfig
}

func getOpenChannelIDs(ctx context.Context, clientCtx client.Context) (map[string]string, error) {
	ibcClient := ibcchanneltypes.NewQueryClient(clientCtx)
	channelsRes, err := ibcClient.Channels(ctx, &ibcchanneltypes.QueryChannelsRequest{})
	if err != nil {
		return nil, err
	}
	chainToChannel := make(map[string]string)
	for _, ch := range channelsRes.Channels {
		if ch.State != ibcchanneltypes.OPEN {
			continue
		}

		chainID, err := getChainID(ctx, clientCtx, ibctransfertypes.PortID, ch.ChannelId)
		if err != nil {
			return nil, err
		}

		chainToChannel[chainID] = ch.ChannelId
	}

	return chainToChannel, err
}

// ConvertToIBCDenom returns the IBC denom based on the channelID and denom.
func ConvertToIBCDenom(channelID, denom string) string {
	return ibctransfertypes.ParseDenomTrace(
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}

// ExecuteIBCTransfer executes IBC transfer transaction.
func ExecuteIBCTransfer(
	ctx context.Context,
	senderChain integrationtests.Chain,
	senderAddress sdk.AccAddress,
	channelID string,
	sendCoin sdk.Coin,
	recipientChain integrationtests.Chain,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	height, err := queryLatestConsensusHeight(
		ctx,
		senderChain.ChainContext.ClientContext,
		ibctransfertypes.PortID,
		channelID,
	)
	if err != nil {
		return nil, err
	}

	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: channelID,
		Token:         sendCoin,
		Sender:        senderChain.ChainContext.ConvertToBech32Address(senderAddress),
		Receiver:      recipientChain.ConvertToBech32Address(recipientAddress),
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: height.RevisionNumber,
			RevisionHeight: height.RevisionHeight + 1000,
		},
	}

	return integrationtests.BroadcastTxWithSigner(
		ctx,
		senderChain.ChainContext,
		senderChain.TxFactory().WithSimulateAndExecute(true),
		senderAddress,
		&ibcSend,
	)
}

// QueryNonZeroIBCBalance queries for the balance with retry and timeout.
func QueryNonZeroIBCBalance(
	ctx context.Context,
	chain integrationtests.Chain,
	address sdk.AccAddress,
	denom string,
) (sdk.Coin, error) {
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	var balance sdk.Coin
	err := retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(retryCtx, 5*time.Second)
		defer requestCancel()

		balancesRes, err := bankClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{
			Address: chain.ConvertToBech32Address(address),
		})
		if err != nil {
			return err
		}

		if balancesRes.Balances.AmountOf(denom).IsZero() {
			return retry.Retryable(errors.Errorf("balances of %s is still empty, all balances:%s", denom, balancesRes.String()))
		}

		balance = sdk.NewCoin(denom, balancesRes.Balances.AmountOf(denom))

		return nil
	})
	if err != nil {
		return sdk.Coin{}, err
	}

	return balance, nil
}

func getChainID(ctx context.Context, clientCtx client.Context, portID, channelID string) (string, error) {
	ibcChannelClient := ibcchanneltypes.NewQueryClient(clientCtx)
	res, err := ibcChannelClient.ChannelClientState(ctx, &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	})
	if err != nil {
		return "", err
	}

	var clientState ibctmlightclient.ClientState
	err = clientCtx.Codec().Unmarshal(res.IdentifiedClientState.ClientState.Value, &clientState)
	if err != nil {
		return "", err
	}

	return clientState.ChainId, nil
}

func queryLatestConsensusHeight(ctx context.Context, clientCtx client.Context, portID, channelID string) (ibcclienttypes.Height, error) {
	queryClient := ibcchanneltypes.NewQueryClient(clientCtx)
	req := &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	clientRes, err := queryClient.ChannelClientState(ctx, req)
	if err != nil {
		return ibcclienttypes.Height{}, err
	}

	var clientState exported.ClientState
	if err := clientCtx.InterfaceRegistry().UnpackAny(clientRes.IdentifiedClientState.ClientState, &clientState); err != nil {
		return ibcclienttypes.Height{}, err
	}

	clientHeight, ok := clientState.GetLatestHeight().(ibcclienttypes.Height)
	if !ok {
		return ibcclienttypes.Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid height type. expected type: %T, got: %T",
			ibcclienttypes.Height{}, clientHeight)
	}

	return clientHeight, nil
}
