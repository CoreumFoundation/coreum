package deterministicgas_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
)

// To access private variable from github.com/gogo/protobuf we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/gogo/protobuf/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

//nolint:funlen
func TestDeterministicGas_DeterministicMessages(t *testing.T) {
	// A list of valid message prefixes or messages which are unknown and not
	// determined as neither deterministic nor undeterministic.
	ignoredMsgTypes := []string{
		// Not-integrated modules:
		// IBC:

		// ibc/core/client
		"/ibc.core.client.v1.MsgCreateClient",
		"/ibc.core.client.v1.MsgCreateClientResponse",
		"/ibc.core.client.v1.MsgUpdateClient",
		"/ibc.core.client.v1.MsgUpdateClientResponse",
		"/ibc.core.client.v1.MsgUpgradeClient",
		"/ibc.core.client.v1.MsgUpgradeClientResponse",
		"/ibc.core.client.v1.MsgSubmitMisbehaviour",
		"/ibc.core.client.v1.MsgSubmitMisbehaviourResponse",

		// ibc/core/connection
		"/ibc.core.connection.v1.MsgConnectionOpenInit",
		"/ibc.core.connection.v1.MsgConnectionOpenInitResponse",
		"/ibc.core.connection.v1.MsgConnectionOpenTry",
		"/ibc.core.connection.v1.MsgConnectionOpenTryResponse",
		"/ibc.core.connection.v1.MsgConnectionOpenAck",
		"/ibc.core.connection.v1.MsgConnectionOpenAckResponse",
		"/ibc.core.connection.v1.MsgConnectionOpenConfirm",
		"/ibc.core.connection.v1.MsgConnectionOpenConfirmResponse",

		// ibc/core/channel
		"/ibc.core.channel.v1.MsgChannelOpenInit",
		"/ibc.core.channel.v1.MsgChannelOpenInitResponse",
		"/ibc.core.channel.v1.MsgChannelOpenTry",
		"/ibc.core.channel.v1.MsgChannelOpenTryResponse",
		"/ibc.core.channel.v1.MsgChannelOpenAck",
		"/ibc.core.channel.v1.MsgChannelOpenAckResponse",
		"/ibc.core.channel.v1.MsgChannelOpenConfirm",
		"/ibc.core.channel.v1.MsgChannelOpenConfirmResponse",
		"/ibc.core.channel.v1.MsgChannelCloseInit",
		"/ibc.core.channel.v1.MsgChannelCloseInitResponse",
		"/ibc.core.channel.v1.MsgChannelCloseConfirm",
		"/ibc.core.channel.v1.MsgChannelCloseConfirmResponse",
		"/ibc.core.channel.v1.MsgRecvPacket",
		"/ibc.core.channel.v1.MsgRecvPacketResponse",
		"/ibc.core.channel.v1.MsgTimeout",
		"/ibc.core.channel.v1.MsgTimeoutResponse",
		"/ibc.core.channel.v1.MsgTimeoutOnClose",
		"/ibc.core.channel.v1.MsgTimeoutOnCloseResponse",
		"/ibc.core.channel.v1.MsgAcknowledgement",
		"/ibc.core.channel.v1.MsgAcknowledgementResponse",

		// ibc.applications.transfer
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/ibc.applications.transfer.v1.MsgTransferResponse",

		// ibc.applications.fee
		"/ibc.applications.fee.v1.MsgRegisterPayee",
		"/ibc.applications.fee.v1.MsgRegisterPayeeResponse",
		"/ibc.applications.fee.v1.MsgRegisterCounterpartyPayee",
		"/ibc.applications.fee.v1.MsgRegisterCounterpartyPayeeResponse",
		"/ibc.applications.fee.v1.MsgPayPacketFee",
		"/ibc.applications.fee.v1.MsgPayPacketFeeResponse",
		"/ibc.applications.fee.v1.MsgPayPacketFeeAsync",
		"/ibc.applications.fee.v1.MsgPayPacketFeeAsyncResponse",

		// Internal cosmos protos:
		"/testdata.TestMsg",
		"/testdata.MsgCreateDog",
		"/cosmos.tx.v1beta1.Tx",
	}

	// WASM messages will be added here
	undetermMsgTypes := []string{
		// crisis
		"/cosmos.crisis.v1beta1.MsgVerifyInvariant",

		// evidence
		"/cosmos.evidence.v1beta1.MsgSubmitEvidence",

		// wasm
		"/cosmwasm.wasm.v1.MsgStoreCode",
		"/cosmwasm.wasm.v1.MsgInstantiateContract",
		"/cosmwasm.wasm.v1.MsgInstantiateContract2",
		"/cosmwasm.wasm.v1.MsgExecuteContract",
		"/cosmwasm.wasm.v1.MsgMigrateContract",
		"/cosmwasm.wasm.v1.MsgIBCCloseChannel",
		"/cosmwasm.wasm.v1.MsgIBCSend",
	}

	cfg := deterministicgas.DefaultConfig()

	var determMsgs []sdk.Msg
	var undetermMsgs []sdk.Msg
	for protoType := range revProtoTypes {
		sdkMsg, ok := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !ok {
			continue
		}

		// Skip unknow messages.
		if lo.ContainsBy(ignoredMsgTypes, func(msgType string) bool {
			return deterministicgas.MsgType(sdkMsg) == msgType
		}) {
			continue
		}

		// Add message to undeterministic.
		if lo.ContainsBy(undetermMsgTypes, func(msgType string) bool {
			return deterministicgas.MsgType(sdkMsg) == msgType
		}) {
			undetermMsgs = append(undetermMsgs, sdkMsg)
			continue
		}

		// Add message to deterministic.
		determMsgs = append(determMsgs, sdkMsg)
	}

	// To make sure we do not increase/decrease deterministic types accidentally
	// we assert length to be equal to exact number, so each change requires
	// explicit adjustment of tests.
	assert.Equal(t, 9, len(undetermMsgs))
	assert.Equal(t, 38, len(determMsgs))

	for _, sdkMsg := range determMsgs {
		sdkMsg := sdkMsg
		t.Run("deterministic: "+deterministicgas.MsgType(sdkMsg), func(t *testing.T) {
			gas, ok := cfg.GasRequiredByMessage(sdkMsg)
			assert.True(t, ok)
			assert.Positive(t, gas)
		})
	}

	for _, sdkMsg := range undetermMsgs {
		sdkMsg := sdkMsg
		t.Run("undeterministic: "+deterministicgas.MsgType(sdkMsg), func(t *testing.T) {
			gas, ok := cfg.GasRequiredByMessage(sdkMsg)
			assert.False(t, ok)
			assert.Zero(t, gas)
		})
	}
}

//nolint:funlen
func TestDeterministicGas_GasRequiredByMessage(t *testing.T) {
	const (
		denom   = "ducore"
		address = "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6"

		assetFTIssue             = 70000
		bankSendPerEntryGas      = 24000
		bankMultiSendPerEntryGas = 11000
		authzMsgExecOverhead     = 2000
	)

	cfg := deterministicgas.DefaultConfig()

	tests := []struct {
		name                    string
		msg                     sdk.Msg
		expectedGas             uint64
		expectedIsDeterministic bool
	}{
		{
			name:                    "wasm.MsgExecuteContract",
			msg:                     &wasmtypes.MsgExecuteContract{},
			expectedGas:             0,
			expectedIsDeterministic: false,
		},
		{
			name:                    "assetft.MsgIssue",
			msg:                     &assetfttypes.MsgIssue{},
			expectedGas:             assetFTIssue,
			expectedIsDeterministic: true,
		},
		{
			name:                    "bank.MsgSend: 0 entries",
			msg:                     &banktypes.MsgSend{},
			expectedGas:             bankSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name:                    "bank.MsgSend: 1 entry",
			msg:                     &banktypes.MsgSend{Amount: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
			expectedGas:             bankSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgSend: 6 entries",
			msg: &banktypes.MsgSend{
				Amount: lo.RepeatBy(6, func(i int) sdk.Coin {
					return sdk.NewCoin(denom, sdk.NewInt(int64(i)))
				}),
			},
			expectedGas:             6 * bankSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name:                    "bank.MsgMultiSend 0 input & 0 output",
			msg:                     &banktypes.MsgMultiSend{},
			expectedGas:             bankMultiSendPerEntryGas * 2,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgMultiSend: 1 input & 1 output",
			msg: &banktypes.MsgMultiSend{
				Inputs: []banktypes.Input{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
				},
			},
			expectedGas:             bankMultiSendPerEntryGas * 2,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgMultiSend: 1 input & 2 outputs",
			msg: &banktypes.MsgMultiSend{
				Inputs: []banktypes.Input{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2)))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
				},
			},
			expectedGas:             3 * bankMultiSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgMultiSend: 3 inputs & 2 outputs",
			msg: &banktypes.MsgMultiSend{
				Inputs: []banktypes.Input{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(2)))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(3)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(3)))},
				},
			},
			expectedGas:             5 * bankMultiSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name:                    "authz.MsgExec: 0 messages",
			msg:                     &authz.MsgExec{},
			expectedGas:             authzMsgExecOverhead,
			expectedIsDeterministic: true,
		},
		{
			name: "authz.MsgExec: 1 bank.MsgSend & 1 bank.MsgMultiSend",
			msg: lo.ToPtr(
				authz.NewMsgExec(
					sdk.AccAddress(address),
					[]sdk.Msg{&banktypes.MsgSend{}, &banktypes.MsgMultiSend{}},
				),
			),
			expectedGas:             authzMsgExecOverhead + bankSendPerEntryGas + 2*bankMultiSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name: "authz.MsgExec: 1 authz.MsgExec (1 bank.MsgSend & 1 bank.MsgMultiSend) & bank.MsgSend",
			msg: lo.ToPtr(
				authz.NewMsgExec(
					sdk.AccAddress(address),
					[]sdk.Msg{
						lo.ToPtr(authz.NewMsgExec(sdk.AccAddress(address), []sdk.Msg{&banktypes.MsgSend{}, &banktypes.MsgMultiSend{}})),
						&banktypes.MsgSend{},
					},
				),
			),
			expectedGas:             authzMsgExecOverhead + authzMsgExecOverhead + bankSendPerEntryGas + 2*bankMultiSendPerEntryGas + bankSendPerEntryGas,
			expectedIsDeterministic: true,
		},
		{
			name: "authz.MsgExec: 1 bank.MsgSend & 1 wasm.MsgExecuteContract",
			msg: lo.ToPtr(
				authz.NewMsgExec(
					sdk.AccAddress(address),
					[]sdk.Msg{&wasmtypes.MsgExecuteContract{}, &banktypes.MsgSend{}},
				),
			),
			expectedGas:             0,
			expectedIsDeterministic: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gas, isDeterministic := cfg.GasRequiredByMessage(tc.msg)
			assert.Equal(t, tc.expectedIsDeterministic, isDeterministic)
			assert.Equal(t, tc.expectedGas, gas)
		})
	}
}
