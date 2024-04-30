//go:build integrationtests

package ibc

import (
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
)

// TestPFMViaCoreum tests the packet forwarding middleware integration into Coreum by sending:
// Osmosis -> Coreum -> Gaia IBC transfer.
func TestPFMViaCoreum(t *testing.T) {
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis
	gaiaChain := chains.Gaia

	osmosisSender := osmosisChain.GenAccount()
	gaiaReceiver := gaiaChain.GenAccount()

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisSender,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		coreumChain.ChainSettings.ChainID,
	)
	coreumToOsmosiChannelID := coreumChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		osmosisChain.ChainSettings.ChainID,
	)

	sendToGaiaCoin := osmosisChain.NewCoin(sdkmath.NewInt(10_000_000))

	// Forward metadata example:
	// {
	//  "forward": {
	//    "receiver": "chain-c-bech32-address",
	//    "port": "transfer",
	//    "channel": "channel-123"
	//  }
	//}
	forwardMetadata := struct {
		Forward packetforwardtypes.ForwardMetadata `json:"forward"`
	}{
		Forward: packetforwardtypes.ForwardMetadata{
			Receiver: gaiaChain.MustConvertToBech32Address(gaiaReceiver),
			Port:     ibctransfertypes.PortID,
			Channel:  gaiaToCoreumChannelID,
		},
	}

	pfmMemo, err := json.Marshal(forwardMetadata)
	requireT.NoError(err)

	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisSender,
		sendToGaiaCoin,
		coreumChain.ChainContext,
		// It is recommended to use an invalid bech32 string (such as "pfm") for the receiver on intermediate chains.
		// More details here:
		//nolint:lll // https://github.com/cosmos/ibc-apps/tree/middleware/packet-forward-middleware/v7.1.3/middleware/packet-forward-middleware#intermediate-receivers
		"pfm",
		string(pfmMemo),
	)
	requireT.NoError(err)

	// Packet denom is the IBC denom sent from coreum to gaia in raw format (without bech32 encoding).
	// Example: "transfer/channel-1/stake"
	packetDenom := fmt.Sprintf("%s/%s/%s", ibctransfertypes.PortID, coreumToOsmosiChannelID, sendToGaiaCoin.Denom)
	// So a received packet on gaia looks like this:
	// port: "transfer"
	// channel: "channel-0"
	// denom: "transfer/channel-1/stake"
	receivedDenomOnGaia := ConvertToIBCDenom(gaiaToCoreumChannelID, packetDenom)

	expectedGaiaReceiverBalance := sdk.NewCoin(receivedDenomOnGaia, sendToGaiaCoin.Amount)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaReceiver, expectedGaiaReceiverBalance))
}
