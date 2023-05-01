package deterministicgas_test

import (
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

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
)

// To access private variable from github.com/gogo/protobuf we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/cosmos/gogoproto/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

//nolint:funlen
func TestDeterministicGas_DeterministicMessages(t *testing.T) {
	// A list of valid message prefixes or messages which are unknown and not
	// determined as neither deterministic nor nondeterministic.
	ignoredMsgTypes := []string{
		// Not-integrated modules:
		// IBC:

		// ibc/core/client
		"/ibc.core.client.v1.MsgCreateClient",
		"/ibc.core.client.v1.MsgUpdateClient",
		"/ibc.core.client.v1.MsgUpgradeClient",
		"/ibc.core.client.v1.MsgSubmitMisbehaviour",

		// ibc/core/connection
		"/ibc.core.connection.v1.MsgConnectionOpenInit",
		"/ibc.core.connection.v1.MsgConnectionOpenTry",
		"/ibc.core.connection.v1.MsgConnectionOpenAck",
		"/ibc.core.connection.v1.MsgConnectionOpenConfirm",

		// ibc/core/channel
		"/ibc.core.channel.v1.MsgChannelOpenInit",
		"/ibc.core.channel.v1.MsgChannelOpenTry",
		"/ibc.core.channel.v1.MsgChannelOpenAck",
		"/ibc.core.channel.v1.MsgChannelOpenConfirm",
		"/ibc.core.channel.v1.MsgChannelCloseInit",
		"/ibc.core.channel.v1.MsgChannelCloseConfirm",
		"/ibc.core.channel.v1.MsgRecvPacket",
		"/ibc.core.channel.v1.MsgTimeout",
		"/ibc.core.channel.v1.MsgTimeoutOnClose",
		"/ibc.core.channel.v1.MsgAcknowledgement",

		// ibc.applications.transfer
		"/ibc.applications.transfer.v1.MsgTransfer",

		// ibc.applications.interchain_accounts
		"/ibc.applications.interchain_accounts.controller.v1.MsgSendTx",
		"/ibc.applications.interchain_accounts.controller.v1.MsgRegisterInterchainAccount",

		// ibc.applications.fee
		"/ibc.applications.fee.v1.MsgRegisterPayee",
		"/ibc.applications.fee.v1.MsgRegisterCounterpartyPayee",
		"/ibc.applications.fee.v1.MsgPayPacketFee",
		"/ibc.applications.fee.v1.MsgPayPacketFeeAsync",

		// Internal cosmos protos:
		"/testdata.TestMsg",
		"/testdata.MsgCreateDog",
		"/cosmos.tx.v1beta1.Tx",
	}

	// WASM messages will be added here
	nondeterministicMsgTypes := []string{
		// auth
		"/cosmos.auth.v1beta1.MsgUpdateParams",

		// bank
		"/cosmos.bank.v1beta1.MsgSetSendEnabled",
		"/cosmos.bank.v1beta1.MsgUpdateParams",

		// distribution
		"/cosmos.distribution.v1beta1.MsgUpdateParams",
		"/cosmos.distribution.v1beta1.MsgCommunityPoolSpend",

		// consensus
		"/cosmos.consensus.v1.MsgUpdateParams",

		// crisis
		"/cosmos.crisis.v1beta1.MsgUpdateParams",

		// crisis
		"/cosmos.crisis.v1beta1.MsgVerifyInvariant",

		// evidence
		"/cosmos.evidence.v1beta1.MsgSubmitEvidence",

		// gov
		"/cosmos.gov.v1beta1.MsgSubmitProposal",

		"/cosmos.gov.v1.MsgSubmitProposal",
		"/cosmos.gov.v1.MsgExecLegacyContent",
		"/cosmos.gov.v1.MsgUpdateParams",

		"/cosmos.gov.v1.MsgSubmitProposal",
		"/cosmos.gov.v1.MsgExecLegacyContent",
		"/cosmos.gov.v1.MsgUpdateParams",

		// mint
		"/cosmos.mint.v1beta1.MsgUpdateParams",

		// staking
		"/cosmos.staking.v1beta1.MsgUpdateParams",
		"/cosmos.staking.v1beta1.MsgCancelUnbondingDelegation",

		// slashing
		"/cosmos.slashing.v1beta1.MsgUpdateParams",

		// upgrade
		"/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",

		// vesting
		"/cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount",
		"/cosmos.vesting.v1beta1.MsgCreatePermanentLockedAccount",

		// wasm
		"/cosmwasm.wasm.v1.MsgStoreCode",
		"/cosmwasm.wasm.v1.MsgInstantiateContract",
		"/cosmwasm.wasm.v1.MsgInstantiateContract2",
		"/cosmwasm.wasm.v1.MsgExecuteContract",
		"/cosmwasm.wasm.v1.MsgMigrateContract",
		"/cosmwasm.wasm.v1.MsgIBCCloseChannel",
		"/cosmwasm.wasm.v1.MsgIBCSend",
		"/cosmwasm.wasm.v1.MsgUpdateInstantiateConfig",
		"/cosmwasm.wasm.v1.MsgUpdateParams",
		"/cosmwasm.wasm.v1.MsgUnpinCodes",
		"/cosmwasm.wasm.v1.MsgPinCodes",
		"/cosmwasm.wasm.v1.MsgSudoContract",
		"/cosmwasm.wasm.v1.MsgStoreAndInstantiateContract",
	}

	// This is required to compile all the messages used by the app, not only those included in deterministic gas config
	simapp.New()

	cfg := deterministicgas.DefaultConfig()

	var deterministicMsgs []sdk.Msg
	var nondeterministicMsgs []sdk.Msg
	for protoType := range revProtoTypes {
		sdkMsg, ok := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !ok {
			continue
		}

		// Skip unknown messages.
		if lo.ContainsBy(ignoredMsgTypes, func(msgType string) bool {
			return deterministicgas.MsgType(sdkMsg) == msgType
		}) {
			continue
		}

		// Add message to nondeterministic.
		if lo.ContainsBy(nondeterministicMsgTypes, func(msgType string) bool {
			return deterministicgas.MsgType(sdkMsg) == msgType
		}) {
			nondeterministicMsgs = append(nondeterministicMsgs, sdkMsg)
			continue
		}

		// Add message to deterministic.
		deterministicMsgs = append(deterministicMsgs, sdkMsg)
	}

	// To make sure we do not increase/decrease deterministic types accidentally
	// we assert length to be equal to exact number, so each change requires
	// explicit adjustment of tests.
	assert.Equal(t, 34, len(nondeterministicMsgs))
	assert.Equal(t, 42, len(deterministicMsgs))

	for _, sdkMsg := range deterministicMsgs {
		sdkMsg := sdkMsg
		t.Run("deterministic: "+deterministicgas.MsgType(sdkMsg), func(t *testing.T) {
			gas, ok := cfg.GasRequiredByMessage(sdkMsg)
			assert.True(t, ok)
			assert.Positive(t, gas)
		})
	}

	for _, sdkMsg := range nondeterministicMsgs {
		sdkMsg := sdkMsg
		t.Run("nondeterministic: "+deterministicgas.MsgType(sdkMsg), func(t *testing.T) {
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
					return sdk.NewCoin(denom, sdkmath.NewInt(int64(i)))
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
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
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
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(2)))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(3)))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(3)))},
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
