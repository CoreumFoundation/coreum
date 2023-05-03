//go:build integrationtests

package ibc

import (
	"context"
	"testing"
	"time"

	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

func TestMain(m *testing.M) {
	// it normally takes 70 seconds to stablish channels for the first time
	// so we 2 minute timeout is good enough
	AwaitChannels(2 * time.Minute)
	m.Run()
}

func AwaitChannels(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	chain := integrationtests.GetChain()

	clientCtx := chain.ChainContext.ClientContext

	ibcChannelClient := ibcchanneltypes.NewQueryClient(clientCtx)

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
			if ch.ChannelId == chain.GaiaContext.ChannelID {
				expectedOpenChannels++
			}
		}

		// currently we only expect gaia channel so the count should be 1
		// TODO: increase the expected value to 2 after we add osmosis
		if expectedOpenChannels == 1 {
			return nil
		}

		return retry.Retryable(errors.New("waiting for channels to open"))
	})
	if err != nil {
		panic(err)
	}
}
