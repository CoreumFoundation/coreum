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

// Ready returns the if the config is fully ready.
func (c ChannelsConfig) Ready() bool {
	if c.CoreumToGaiaChannelID == "" || c.GaiaToCoreumChannelID == "" {
		return false
	}

	return true
}

// AwaitForIBCConfig await for the IBC channels to be opened and returns them.
// TODO(milad): remove the await after we build this logic into crust.
func AwaitForIBCConfig(t *testing.T) ChannelsConfig {
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumIBCChannelClient := ibcchanneltypes.NewQueryClient(chains.Coreum.ClientContext)
	gaiaIBCChannelClient := ibcchanneltypes.NewQueryClient(chains.Gaia.ClientContext)

	var ibcConfig ChannelsConfig
	err := retry.Do(ctx, time.Second, func() error {
		coreumChannelsRes, err := coreumIBCChannelClient.Channels(ctx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range coreumChannelsRes.Channels {
			if ch.State != ibcchanneltypes.OPEN {
				continue
			}

			chainID, err := getChainID(ctx, chains.Coreum.ClientContext, ibctransfertypes.PortID, ch.ChannelId)
			if err != nil {
				return err
			}

			if chainID == chains.Gaia.ChainSettings.ChainID {
				ibcConfig.CoreumToGaiaChannelID = ch.ChannelId
			}
		}

		gaiaChannelsRes, err := gaiaIBCChannelClient.Channels(ctx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range gaiaChannelsRes.Channels {
			if ch.State != ibcchanneltypes.OPEN {
				continue
			}

			chainID, err := getChainID(ctx, chains.Gaia.ClientContext, ibctransfertypes.PortID, ch.ChannelId)
			if err != nil {
				return err
			}

			if chainID == chains.Coreum.ChainSettings.ChainID {
				ibcConfig.GaiaToCoreumChannelID = ch.ChannelId
			}
		}

		if ibcConfig.Ready() {
			return nil
		}

		return retry.Retryable(errors.New("waiting for channels to open"))
	})
	require.NoError(t, err)

	return ibcConfig
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
