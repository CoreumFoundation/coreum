package app_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	"cosmossdk.io/api/amino"
	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/nft"
	sdktestdatatypes "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktxtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	ibclightclienttypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	ibcinterchainaccountscontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	ibcinterchainaccountshosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcv2clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/v2/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibcv2channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	protobuf "github.com/golang/protobuf/proto" //nolint:staticcheck // We need this dependency to convert protos to be able to read their options
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// To access private variable from github.com/cosmos/gogoproto we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/cosmos/gogoproto/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestLegacyAmino_ExpectedMessages(t *testing.T) {
	expectedNonAminoMsgURLs := map[string]struct{}{
		// bank
		sdk.MsgTypeURL(&banktypes.Input{}): {},

		// nft
		sdk.MsgTypeURL(&nft.MsgSend{}): {},

		// gov
		sdk.MsgTypeURL(&govtypesv1.MsgCancelProposal{}): {},

		// ibc/core/client
		sdk.MsgTypeURL(&ibcclienttypes.MsgCreateClient{}):        {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateClient{}):        {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpgradeClient{}):       {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgDeleteClientCreator{}): {},
		//nolint:staticcheck // it is here so the list is complete.
		sdk.MsgTypeURL(&ibcclienttypes.MsgSubmitMisbehaviour{}):     {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgIBCSoftwareUpgrade{}):     {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgRecoverClient{}):          {},
		sdk.MsgTypeURL(&ibcclienttypes.MsgUpdateParams{}):           {},
		sdk.MsgTypeURL(&ibcv2clienttypes.MsgRegisterCounterparty{}): {},
		sdk.MsgTypeURL(&ibcv2clienttypes.MsgUpdateClientConfig{}):   {},

		// ibc/lightclients
		sdk.MsgTypeURL(&ibclightclienttypes.MsgStoreCode{}):       {},
		sdk.MsgTypeURL(&ibclightclienttypes.MsgRemoveChecksum{}):  {},
		sdk.MsgTypeURL(&ibclightclienttypes.MsgMigrateContract{}): {},

		// ibc/apps/transfer
		sdk.MsgTypeURL(&ibctransfertypes.MsgUpdateParams{}): {},

		// ibc/core/connection
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenInit{}):    {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenTry{}):     {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenAck{}):     {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgConnectionOpenConfirm{}): {},
		sdk.MsgTypeURL(&ibcconnectiontypes.MsgUpdateParams{}):          {},

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
		sdk.MsgTypeURL(&ibcv2channeltypes.MsgSendPacket{}):        {},
		sdk.MsgTypeURL(&ibcv2channeltypes.MsgRecvPacket{}):        {},
		sdk.MsgTypeURL(&ibcv2channeltypes.MsgTimeout{}):           {},
		sdk.MsgTypeURL(&ibcv2channeltypes.MsgAcknowledgement{}):   {},

		// ibc/applications/interchain_accounts/controller
		sdk.MsgTypeURL(&ibcinterchainaccountscontrollertypes.MsgRegisterInterchainAccount{}): {},
		sdk.MsgTypeURL(&ibcinterchainaccountscontrollertypes.MsgSendTx{}):                    {},
		sdk.MsgTypeURL(&ibcinterchainaccountscontrollertypes.MsgUpdateParams{}):              {},

		// ibc/applications/interchain_accounts/host
		sdk.MsgTypeURL(&ibcinterchainaccountshosttypes.MsgModuleQuerySafe{}): {},
		sdk.MsgTypeURL(&ibcinterchainaccountshosttypes.MsgUpdateParams{}):    {},

		// internal cosmos
		sdk.MsgTypeURL(&sdktestdatatypes.MsgCreateDog{}): {},
		sdk.MsgTypeURL(&sdktxtypes.Tx{}):                 {},

		// feegrant
		sdk.MsgTypeURL(&feegrant.MsgPruneAllowances{}): {},
	}

	for protoType := range revProtoTypes {
		protoInterface := reflect.New(protoType.Elem()).Interface()

		sdkMsg, isSDKMessage := protoInterface.(sdk.Msg)
		if !isSDKMessage {
			continue
		}

		messageURL := sdk.MsgTypeURL(sdkMsg)
		_, isLegacyMessage := protoInterface.(legacytx.LegacyMsg)
		if isLegacyMessage {
			continue
		}

		options := protobuf.MessageV2(protoInterface).ProtoReflect().Descriptor().Options()

		signersFields := proto.GetExtension(options, msgv1.E_Signer).([]string)
		if len(signersFields) == 0 && messageURL != "/cosmos.tx.v1beta1.Tx" {
			continue
		}

		aminoNameField := proto.GetExtension(options, amino.E_Name).(string)
		if len(aminoNameField) > 0 {
			continue
		}

		_, expectedNonAmino := expectedNonAminoMsgURLs[messageURL]
		require.True(t, expectedNonAmino, "Unexpected non-amino message:%s", messageURL)
		delete(expectedNonAminoMsgURLs, messageURL)
	}

	require.Empty(t, expectedNonAminoMsgURLs)
}
