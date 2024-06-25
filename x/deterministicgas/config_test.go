package deterministicgas_test

import (
	"fmt"
	"reflect"
	"testing"
	_ "unsafe"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/deterministicgas"
	"github.com/CoreumFoundation/coreum/v4/x/deterministicgas/types"
)

// To access private variable from github.com/cosmos/gogoproto we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/cosmos/gogoproto/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestDeterministicGas_DeterministicMessages(t *testing.T) {
	// A list of valid message prefixes or messages which are unknown and not
	// determined as neither deterministic nor nondeterministic.
	ignoredMsgURLs := []deterministicgas.MsgURL{
		// Not-integrated modules:
		// IBC:

		// ibc.applications.fee
		"/ibc.applications.fee.v1.MsgRegisterPayee",
		"/ibc.applications.fee.v1.MsgRegisterCounterpartyPayee",
		"/ibc.applications.fee.v1.MsgPayPacketFee",
		"/ibc.applications.fee.v1.MsgPayPacketFeeAsync",

		// Internal cosmos protos:
		"/testpb.TestMsg",
		"/testpb.MsgCreateDog",
		"/cosmos.tx.v1beta1.Tx",
	}

	// This is required to compile all the messages used by the app, not only those included in deterministic gas config
	simapp.New()

	cfg := deterministicgas.DefaultConfig()

	deterministicMsgCount := 0
	nondeterministicMsgCount := 0
	extensionMsgCount := 0
	nonextensionMsgCount := 0
	for protoType := range revProtoTypes {
		sdkMsg, ok := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !ok {
			continue
		}

		// skip some messages which don't have the message handlers
		if lo.ContainsBy(ignoredMsgURLs, func(msgURL deterministicgas.MsgURL) bool {
			return deterministicgas.MsgToMsgURL(sdkMsg) == msgURL
		}) {
			continue
		}

		msgURL := deterministicgas.MsgToMsgURL(sdkMsg)
		gasFunc, ok := cfg.GasByMessageMap()[msgURL]
		assert.True(t, ok, fmt.Sprintf("sdk.Msg %s, not found in the gasByMsg map", msgURL))

		_, _, nonExtensionMsg, err := types.TypeAssertMessages(sdkMsg)
		require.NoError(t, err)
		if nonExtensionMsg {
			nonextensionMsgCount++
		} else {
			extensionMsgCount++
		}

		gas, ok := gasFunc(sdkMsg)
		if ok {
			assert.NotZero(t, gas)
			deterministicMsgCount++
			continue
		}
		assert.Zero(t, gas)
		nondeterministicMsgCount++
	}

	// To make sure we do not increase/decrease deterministic and extension types accidentally,
	// we assert length to be equal to exact number, so each change requires
	// explicit adjustment of tests.
	assert.Equal(t, 62, nondeterministicMsgCount)
	assert.Equal(t, 69, deterministicMsgCount)
	assert.Equal(t, 14, extensionMsgCount)
	assert.Equal(t, 117, nonextensionMsgCount)
}

func TestDeterministicGas_GasRequiredByMessage(t *testing.T) {
	const (
		denom   = "ducore"
		address = "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6"

		assetFTIssue                 = 70000
		bankSendPerCoinGas           = deterministicgas.BankSendPerCoinGas
		bankMultiSendPerOperationGas = deterministicgas.BankMultiSendPerOperationsGas
		authzMsgExecOverhead         = deterministicgas.AuthzExecOverhead
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
			expectedGas:             bankSendPerCoinGas,
			expectedIsDeterministic: true,
		},
		{
			name:                    "bank.MsgSend: 1 entry",
			msg:                     &banktypes.MsgSend{Amount: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
			expectedGas:             bankSendPerCoinGas,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgSend: 6 entries",
			msg: &banktypes.MsgSend{
				Amount: lo.RepeatBy(6, func(i int) sdk.Coin {
					return sdk.NewCoin(denom, sdkmath.NewInt(int64(i)))
				}),
			},
			expectedGas:             6 * bankSendPerCoinGas,
			expectedIsDeterministic: true,
		},
		{
			name:                    "bank.MsgMultiSend 0 input & 0 output",
			msg:                     &banktypes.MsgMultiSend{},
			expectedGas:             bankMultiSendPerOperationGas * 2,
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
			expectedGas:             bankMultiSendPerOperationGas * 2,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgMultiSend: 1 input & 2 outputs",
			msg: &banktypes.MsgMultiSend{
				Inputs: []banktypes.Input{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt()))},
				},
			},
			expectedGas:             3 * bankMultiSendPerOperationGas,
			expectedIsDeterministic: true,
		},
		{
			name: "bank.MsgMultiSend: 3 inputs & 2 outputs",
			msg: &banktypes.MsgMultiSend{
				Inputs: []banktypes.Input{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(3)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(3)))},
				},
			},
			expectedGas:             5 * bankMultiSendPerOperationGas,
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
			expectedGas:             authzMsgExecOverhead + bankSendPerCoinGas + 2*bankMultiSendPerOperationGas,
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
			expectedGas: authzMsgExecOverhead +
				authzMsgExecOverhead +
				bankSendPerCoinGas +
				2*bankMultiSendPerOperationGas +
				bankSendPerCoinGas,
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
