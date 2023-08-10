package app_test

import (
	"fmt"
	"reflect"
	"testing"
	_ "unsafe"

	sdktestdatatypes "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktxtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcinterchainaccountstypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

// To access private variable from github.com/cosmos/gogoproto we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/cosmos/gogoproto/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestLegacyAmino_ExpectedMessages(t *testing.T) {
	expectedNonAminoMsgURLs := map[string]struct{}{
		// auth
		sdk.MsgTypeURL(&authtypes.MsgUpdateParams{}): {},

		// bank
		sdk.MsgTypeURL(&banktypes.MsgUpdateParams{}):   {},
		sdk.MsgTypeURL(&banktypes.MsgSetSendEnabled{}): {},

		// gov
		sdk.MsgTypeURL(&govtypesv1.MsgExecLegacyContent{}): {},

		// mint
		sdk.MsgTypeURL(&minttypes.MsgUpdateParams{}): {},

		// slashing
		sdk.MsgTypeURL(&slashingtypes.MsgUpdateParams{}): {},

		// staking
		sdk.MsgTypeURL(&stakingtypes.MsgUpdateParams{}): {},

		// ibc/core/client
		sdk.MsgTypeURL(&ibcclienttypes.MsgCreateClient{}):       {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}):       {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpgradeClient{}):      {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgSubmitMisbehaviour{}): {},

		// ibc/core/connection
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenInit{}):    {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenTry{}):     {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenAck{}):     {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenConfirm{}): {},

		// ibc/core/channel
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenInit{}):     {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenTry{}):      {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenAck{}):      {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelOpenConfirm{}):  {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelCloseInit{}):    {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgChannelCloseConfirm{}): {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgRecvPacket{}):          {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeout{}):             {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgTimeoutOnClose{}):      {},
		sdk.MsgTypeURL(&ibcchanneltypes.MsgAcknowledgement{}):     {},

		// ibc/applications/interchain_accounts
		sdk.MsgTypeURL(&ibcinterchainaccountstypes.MsgRegisterInterchainAccount{}): {},
		sdk.MsgTypeURL(&ibcinterchainaccountstypes.MsgSendTx{}):                    {},

		// internal cosmos
		sdk.MsgTypeURL(&sdktestdatatypes.MsgCreateDog{}): {},
		sdk.MsgTypeURL(&sdktxtypes.Tx{}):                 {},
	}

	for protoType := range revProtoTypes {
		sdkMsg, isSDKMessage := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !isSDKMessage {
			continue
		}

		messageURL := sdk.MsgTypeURL(sdkMsg)
		_, isLegacyMessage := reflect.New(protoType.Elem()).Interface().(legacytx.LegacyMsg)

		if isLegacyMessage {
			continue
		}

		_, expectedNonAmino := expectedNonAminoMsgURLs[messageURL]
		require.True(t, expectedNonAmino, fmt.Sprintf("Unexpected non-amino message:%s", messageURL))
		delete(expectedNonAminoMsgURLs, messageURL)
	}

	require.Empty(t, expectedNonAminoMsgURLs)
}
