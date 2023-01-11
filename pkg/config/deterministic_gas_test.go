package config_test

import (
	"reflect"
	"strings"
	"testing"
	_ "unsafe"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/pkg/config"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// To access private variable from github.com/gogo/protobuf we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/gogo/protobuf/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestDeterministicGasRequirements_DeterministicMessages(t *testing.T) {
	// A list of valid message prefixes or messages which are unknown and not
	// determined as neither deterministic nor undeterministic.
	unknownMsgPrefixes := []string{
		// Not-integrated modules:
		// IBC:
		"/ibc.core.",
		"/ibc.applications.",
		// To be integrated standard modules:
		"/cosmos.evidence.",
		"/cosmos.crisis.",
		"/cosmos.vesting.",

		// Internal cosmos protos:
		"/testdata.",
		"/cosmos.tx.v1beta1.Tx",
	}

	// WASM messages will be added here
	undetermMsgPrefixes := []string{
		// CosmWasm
		"/cosmwasm.wasm.",
	}

	dgr := config.DefaultDeterministicGasRequirements()

	var determMsgs []sdk.Msg
	var undetermMsgs []sdk.Msg
	for protoType := range revProtoTypes {
		sdkMsg, ok := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !ok {
			continue
		}

		// Skip unknow messages.
		if lo.ContainsBy(unknownMsgPrefixes, func(prefix string) bool {
			return strings.HasPrefix(config.MsgName(sdkMsg), prefix)
		}) {
			continue
		}

		// Add message to undeterministic.
		if lo.ContainsBy(undetermMsgPrefixes, func(prefix string) bool {
			return strings.HasPrefix(config.MsgName(sdkMsg), prefix)
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
	assert.Equal(t, 33, len(determMsgs))
	assert.Equal(t, 9, len(undetermMsgs))

	for _, sdkMsg := range determMsgs {
		sdkMsg := sdkMsg
		t.Run("deterministic: "+config.MsgName(sdkMsg), func(t *testing.T) {
			gas, ok := dgr.GasRequiredByMessage(sdkMsg)
			assert.True(t, ok)
			assert.Positive(t, gas)
		})
	}

	for _, sdkMsg := range undetermMsgs {
		sdkMsg := sdkMsg
		t.Run("undeterministic: "+config.MsgName(sdkMsg), func(t *testing.T) {
			gas, ok := dgr.GasRequiredByMessage(sdkMsg)
			assert.False(t, ok)
			assert.Zero(t, gas)
		})
	}
}

//nolint:funlen
func TestDeterministicGasRequirements_GasRequiredByMessage(t *testing.T) {
	const (
		denom   = "ducore"
		address = "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6"

		assetFTIssue             = 80000
		bankSendPerEntryGas      = 22000
		bankMultiSendPerEntryGas = 27000
		authzMsgExecOverhead     = 2000
	)

	dgr := config.DefaultDeterministicGasRequirements()

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
			expectedGas:             bankMultiSendPerEntryGas,
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
			expectedGas:             bankMultiSendPerEntryGas,
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
			expectedGas:             2 * bankMultiSendPerEntryGas,
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
			expectedGas:             3 * bankMultiSendPerEntryGas,
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
			expectedGas:             authzMsgExecOverhead + bankSendPerEntryGas + bankMultiSendPerEntryGas,
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
			expectedGas:             authzMsgExecOverhead + authzMsgExecOverhead + bankSendPerEntryGas + bankMultiSendPerEntryGas + bankSendPerEntryGas,
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
			gas, isDeterministic := dgr.GasRequiredByMessage(tc.msg)
			assert.Equal(t, tc.expectedIsDeterministic, isDeterministic)
			assert.Equal(t, tc.expectedGas, gas)
		})
	}
}
