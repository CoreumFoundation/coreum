//go:build integrationtests

package ibc

import (
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

const (
	// It is recommended to use an invalid bech32 string (such as "pfm") for the receiver on intermediate chains.
	// More details here:
	//nolint:lll // https://github.com/cosmos/ibc-apps/tree/middleware/packet-forward-middleware/v7.1.3/middleware/packet-forward-middleware#intermediate-receivers
	pfmRecipient = "pfm"
)

// Forward metadata example:
//
//	{
//	 "forward": {
//	   "receiver": "chain-c-bech32-address",
//	   "port": "transfer",
//	   "channel": "channel-123" // this is the chain C on chain B channel.
//	 }
//	}
type pfmForwardMetadata struct {
	Forward packetforwardtypes.ForwardMetadata `json:"forward"`
}

// TestPFMViaCoreumForOsmosisToken tests the packet
// forwarding middleware integration into Coreum by sending Osmosis native token:
// Osmosis -> Coreum -> Gaia IBC transfer.
func TestPFMViaCoreumForOsmosisToken(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis
	gaiaChain := chains.Gaia

	osmosisSender := osmosisChain.GenAccount()
	coreumSender := coreumChain.GenAccount()

	gaiaReceiver := gaiaChain.GenAccount()

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisSender,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)
	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumSender,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		gaiaChain.ChainContext,
	)
	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		coreumChain.ChainContext,
	)
	coreumToOsmosiChannelID := coreumChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		osmosisChain.ChainContext,
	)

	forwardMetadata := pfmForwardMetadata{
		Forward: packetforwardtypes.ForwardMetadata{
			Receiver: gaiaChain.MustConvertToBech32Address(gaiaReceiver),
			Port:     ibctransfertypes.PortID,
			Channel:  coreumToGaiaChannelID,
		},
	}

	pfmMemo, err := json.Marshal(forwardMetadata)
	requireT.NoError(err)

	sendToGaiaCoin := osmosisChain.NewCoin(sdkmath.NewInt(10_000_000))
	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain.TxFactoryAuto(),
		osmosisSender,
		sendToGaiaCoin,
		coreumChain.ChainContext,
		pfmRecipient,
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

// TestPFMViaCoreumForCoreumToken tests the packet forwarding middleware integration into Coreum
// by sending Coreum native token to Osmosis and then sending it to gaia via Coreum:
// tx1: Coreum -> Osmosis, tx2: Osmosis -> Coreum -> Gaia.
func TestPFMViaCoreumForCoreumToken(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis
	gaiaChain := chains.Gaia

	osmosisSender := osmosisChain.GenAccount()
	coreumSender := coreumChain.GenAccount()

	gaiaReceiver := gaiaChain.GenAccount()

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisSender,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)
	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumSender,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		gaiaChain.ChainContext,
	)
	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		coreumChain.ChainContext,
	)
	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		coreumChain.ChainContext,
	)

	// ********** Send funds to Osmosis **********

	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(10_000_000))
	_, err := coreumChain.ExecuteIBCTransfer(
		ctx,
		t,
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{})),
		coreumSender,
		sendToOsmosisCoin,
		osmosisChain.ChainContext,
		osmosisSender,
	)
	requireT.NoError(err)

	expectedOsmosisRecipientBalance := sdk.NewCoin(
		ConvertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		sendToOsmosisCoin.Amount,
	)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, t, osmosisSender, expectedOsmosisRecipientBalance))

	// ********** Send funds to Gaia via Coreum using PFM **********

	forwardMetadata := pfmForwardMetadata{
		Forward: packetforwardtypes.ForwardMetadata{
			Receiver: gaiaChain.MustConvertToBech32Address(gaiaReceiver),
			Port:     ibctransfertypes.PortID,
			Channel:  coreumToGaiaChannelID,
		},
	}

	pfmMemo, err := json.Marshal(forwardMetadata)
	requireT.NoError(err)

	sendToGaiaCoin := expectedOsmosisRecipientBalance
	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain.TxFactoryAuto(),
		osmosisSender,
		sendToGaiaCoin,
		coreumChain.ChainContext,
		pfmRecipient,
		string(pfmMemo),
	)
	requireT.NoError(err)

	// Note that denom is resolved in the same way as if was sent from Coreum to Gaia directly.
	expectedGaiaReceiverBalance := sdk.NewCoin(
		ConvertToIBCDenom(gaiaToCoreumChannelID, coreumChain.ChainSettings.Denom),
		sendToGaiaCoin.Amount,
	)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaReceiver, expectedGaiaReceiverBalance))
}
