package deterministicgas_test

import (
	"reflect"
	"testing"
	"time"
	_ "unsafe"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	protobuf "github.com/golang/protobuf/proto" //nolint:staticcheck // We need this dependency to convert protos to be able to read their options
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v6/x/deterministicgas"
	"github.com/CoreumFoundation/coreum/v6/x/deterministicgas/types"
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
		// Internal cosmos protos:
		"/testpb.TestMsg",
		"/testpb.MsgCreateDog",
		"/cosmos.tx.v1beta1.Tx",
		"/cosmos.bank.v1beta1.Input",
	}

	// This is required to compile all the messages used by the app, not only those included in deterministic gas config
	simapp.New()

	cfg := deterministicgas.DefaultConfig()

	deterministicMsgCount := 0
	nondeterministicMsgCount := 0
	extensionMsgCount := 0
	nonExtensionMsgCount := 0
	for protoType := range revProtoTypes {
		sdkMsg, ok := reflect.New(protoType.Elem()).Interface().(sdk.Msg)
		if !ok {
			continue
		}

		options := protobuf.MessageV2(reflect.New(protoType.Elem()).Interface()).ProtoReflect().Descriptor().Options()

		signersFields := proto.GetExtension(options, msgv1.E_Signer).([]string)
		if len(signersFields) == 0 {
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
		assert.True(t, ok, "sdk.Msg %s, not found in the gasByMsg map", msgURL)

		_, _, nonExtensionMsg, err := types.TypeAssertMessages(sdkMsg)
		require.NoError(t, err)
		if nonExtensionMsg {
			nonExtensionMsgCount++
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
	assert.Equal(t, 85, nondeterministicMsgCount)
	assert.Equal(t, 68, deterministicMsgCount)
	assert.Equal(t, 12, extensionMsgCount)
	assert.Equal(t, 141, nonExtensionMsgCount)
}

func TestDeterministicGas_GasRequiredByMessage(t *testing.T) {
	const (
		denom   = "ducore"
		address = "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6"

		assetFTIssue                 = 70000
		bankSendPerCoinGas           = deterministicgas.BankSendPerCoinGas
		bankMultiSendPerOperationGas = deterministicgas.BankMultiSendPerOperationsGas
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
			msg:                     &banktypes.MsgSend{Amount: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.OneInt()))},
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
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.OneInt()))},
				},
				Outputs: []banktypes.Output{
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.OneInt()))},
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
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.OneInt()))},
					{Coins: sdk.NewCoins(sdk.NewCoin(denom, sdkmath.OneInt()))},
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
		t.Run(tc.name, func(t *testing.T) {
			gas, isDeterministic := cfg.GasRequiredByMessage(tc.msg)
			assert.Equal(t, tc.expectedIsDeterministic, isDeterministic)
			assert.Equal(t, tc.expectedGas, gas)
		})
	}
}

func TestDeterministicGas_AuthzGrant(t *testing.T) {
	address := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	testCases := []struct {
		name            string
		authzItemsCount int
		expectedGas     uint64
	}{
		{
			name:            "1_item",
			authzItemsCount: 1,
			expectedGas:     28000,
		},
		{
			name:            "50_items",
			authzItemsCount: 50,
			expectedGas:     187000,
		},
		{
			name:            "100_items",
			authzItemsCount: 100,
			expectedGas:     350000,
		},
	}
	genAuthFuncs := []struct {
		name string
		fn   func(itemsCount int) authz.Authorization
	}{
		{
			name: "send_auth",
			fn: func(itemsCount int) authz.Authorization {
				authorization := &assetnfttypes.SendAuthorization{}
				for range itemsCount {
					authorization.Nfts = append(authorization.Nfts, assetnfttypes.NFTIdentifier{
						ClassId: "class-id-" + address.String(),
						Id:      "id-" + address.String(),
					})
				}
				return authorization
			},
		},
		{
			name: "mint_auth",
			fn: func(itemsCount int) authz.Authorization {
				authorization := &assetfttypes.MintAuthorization{}
				for range itemsCount {
					authorization.MintLimit = append(
						authorization.MintLimit,
						sdk.NewCoin("random-denom-"+address.String(), sdkmath.NewInt(1_000_000_000_000)),
					)
				}
				return authorization
			},
		},
		{
			name: "burn_auth",
			fn: func(itemsCount int) authz.Authorization {
				authorization := &assetfttypes.BurnAuthorization{}
				for range itemsCount {
					authorization.BurnLimit = append(
						authorization.BurnLimit,
						sdk.NewCoin("random-denom-"+address.String(), sdkmath.NewInt(1_000_000_000_000)),
					)
				}
				return authorization
			},
		},
	}

	cfg := deterministicgas.DefaultConfig()
	for _, gen := range genAuthFuncs {
		for _, tc := range testCases {
			gen := gen
			t.Run(tc.name+"_"+gen.name, func(t *testing.T) {
				requireT := require.New(t)
				authorization := gen.fn(tc.authzItemsCount)
				grantMsg, err := authz.NewMsgGrant(
					address,
					address,
					authorization,
					lo.ToPtr(time.Now().Add(time.Minute)),
				)
				requireT.NoError(err)

				deterministicGas, ok := cfg.GasRequiredByMessage(grantMsg)
				requireT.True(ok)
				requireT.InEpsilon(tc.expectedGas, deterministicGas, 0.3)
			})
		}
	}
}
