//go:build integrationtests

package ibc

import (
	"context"
	"testing"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	ibctmlightclient "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// TODO: remove the await after we build this logic into crust.
func awaitChannels(ctx context.Context, srcChainID string, dstClientContext client.Context, t *testing.T) string {
	ibcChannelClient := ibcchanneltypes.NewQueryClient(dstClientContext)
	var channelID string

	expectedOpenChannels := 0

	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()

	err := retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
		defer requestCancel()

		channels, err := ibcChannelClient.Channels(requestCtx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range channels.Channels {
			if ch.State != ibcchanneltypes.OPEN {
				continue
			}

			if ch.ChannelId == channelID {
				continue
			}

			chainID, err := getChainID(ctx, dstClientContext, ibctransfertypes.PortID, ch.ChannelId)
			if err != nil {
				return err
			}

			if chainID == srcChainID {
				expectedOpenChannels++
				channelID = ch.ChannelId
			}
		}

		if expectedOpenChannels == 1 {
			return nil
		}

		return retry.Retryable(errors.New("waiting for channels to open"))
	})
	require.NoError(t, err)

	return channelID
}

func getChainID(
	ctx context.Context,
	clientCtx client.Context,
	portID string,
	channelID string,
) (string, error) {
	ibcChannelClient := ibcchanneltypes.NewQueryClient(clientCtx)

	requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
	defer requestCancel()

	res, err := ibcChannelClient.ChannelClientState(requestCtx, &ibcchanneltypes.QueryChannelClientStateRequest{
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

func queryLatestConsensusHeight(ctx context.Context,
	clientCtx client.Context, portID, channelID string,
) (ibcclienttypes.Height, error) {
	queryClient := ibcchanneltypes.NewQueryClient(clientCtx)
	req := &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
	defer requestCancel()

	clientRes, err := queryClient.ChannelClientState(requestCtx, req)
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
