//go:build integrationtests

package ibc

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

type IBCChain struct {
	integrationtests.Chain
}

// convertToIBCDenom returns the IBC denom based on the channelID and denom.
func convertToIBCDenom(channelID, denom string) string {
	return ibctransfertypes.ParseDenomTrace(
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}

// NewIBCTestingContext returns the configured chains and new context for the integration tests.
func NewIBCTestingContext(t *testing.T) (context.Context, integrationtests.CoreumChain, IBCChains) {
	testCtx, coreumChain := integrationtests.NewCoreumTestingContext(t)
	return testCtx, coreumChain, IBCChains{
		Coreum:  IBCChain{coreumChain.Chain},
		Gaia:    chains.Gaia,
		Osmosis: chains.Osmosis,
	}
}

// ExecuteIBCTransfer executes IBC transfer transaction.
func (c IBCChain) ExecuteIBCTransfer(
	ctx context.Context,
	t *testing.T,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext IBCChain,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := c.ConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.ConvertToBech32Address(recipientAddress)
	t.Logf("Sending IBC transfer sender: %s, receiver: %s, amount: %s.", sender, receiver, coin.String())

	recipientChannelID := c.GetIBCChannelID(ctx, t, recipientChainContext.ChainSettings.ChainID)
	height, err := queryLatestConsensusHeight(
		ctx,
		c.ClientContext,
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

	txRes, err := integrationtests.BroadcastTxWithSigner(
		ctx,
		c.ChainContext,
		c.TxFactory().WithSimulateAndExecute(true),
		senderAddress,
		&ibcSend,
	)

	return txRes, err
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
