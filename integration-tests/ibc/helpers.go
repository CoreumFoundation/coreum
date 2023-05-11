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
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

type channelsInfo struct {
	gaiaChannelID string
}

// TODO: remove the await after we build this logic into crust.
func awaitChannels(t *testing.T) channelsInfo {
	ctx, chain := integrationtests.NewTestingContext(t)
	clientCtx := chain.ChainContext.ClientContext

	ibcChannelClient := ibcchanneltypes.NewQueryClient(clientCtx)
	var gaiaChannelID string

	expectedOpenChannels := 0
	err := retry.Do(ctx, time.Second, func() error {
		channels, err := ibcChannelClient.Channels(ctx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range channels.Channels {
			if ch.State != ibcchanneltypes.OPEN {
				continue
			}

			if ch.ChannelId == gaiaChannelID {
				continue
			}

			chainID, err := getChainID(ctx, clientCtx, ibctransfertypes.PortID, ch.ChannelId)
			if err != nil {
				return err
			}
			if chainID == chain.GaiaContext.ClientContext.ChainID() {
				expectedOpenChannels++
				gaiaChannelID = ch.ChannelId
			}
		}

		if expectedOpenChannels == 1 {
			return nil
		}

		return retry.Retryable(errors.New("waiting for channels to open"))
	})
	require.NoError(t, err)

	return channelsInfo{
		gaiaChannelID: gaiaChannelID,
	}
}

func getChainID(
	ctx context.Context,
	clientCtx client.Context,
	portID string,
	channelID string,
) (string, error) {
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

func queryLatestConsensusHeight(
	clientCtx client.Context, portID, channelID string,
) (ibcclienttypes.Height, error) {
	queryClient := ibcchanneltypes.NewQueryClient(clientCtx)
	req := &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	clientRes, err := queryClient.ChannelClientState(context.Background(), req)
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
