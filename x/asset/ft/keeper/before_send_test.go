package keeper_test

import (
	"bytes"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	assetftkeeper "github.com/CoreumFoundation/coreum/v4/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	cwasmtypes "github.com/CoreumFoundation/coreum/v4/x/wasm/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v4/x/wibctransfer/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

func TestApplyRate(t *testing.T) {
	genAccount := func() string {
		return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	}
	var accounts []string
	var smartContracts []string
	for i := 0; i < 11; i++ {
		accounts = append(accounts, genAccount())
	}
	for i := byte(0); i < 2; i++ {
		smartContracts = append(smartContracts, sdk.AccAddress(bytes.Repeat([]byte{i}, wasmtypes.ContractAddrLen)).String())
	}

	issuer := genAccount()
	dummyAddress := genAccount()
	key := sdk.NewKVStoreKey(types.StoreKey)
	assetFTKeeper := assetftkeeper.NewKeeper(nil, key, nil, nil, nil, nil, nil, "")

	testCases := []struct {
		name         string
		rate         string
		sender       string
		recipient    string
		amount       sdkmath.Int
		ibcDirection wibctransfertypes.Purpose
		appliedRate  sdkmath.Int
	}{
		{
			name:        "issuer_receiver",
			rate:        "0.5",
			sender:      accounts[0],
			recipient:   issuer,
			amount:      sdkmath.NewInt(10),
			appliedRate: sdkmath.NewInt(5),
		},
		{
			name:        "issuer_sender",
			rate:        "0.5",
			sender:      issuer,
			recipient:   accounts[5],
			amount:      sdkmath.NewInt(5),
			appliedRate: sdkmath.NewInt(3),
		},
		{
			name:        "non_issuer",
			rate:        "0.1",
			sender:      accounts[0],
			recipient:   accounts[5],
			amount:      sdkmath.NewInt(1000),
			appliedRate: sdkmath.NewInt(100),
		},
		{
			name:        "with_rounding",
			rate:        "0.1",
			sender:      accounts[0],
			recipient:   accounts[10],
			amount:      sdkmath.NewInt(1001),
			appliedRate: sdkmath.NewInt(101),
		},
		{
			name:         "sender_ibc",
			rate:         "0.5",
			sender:       accounts[0],
			recipient:    dummyAddress,
			amount:       sdkmath.NewInt(10),
			ibcDirection: wibctransfertypes.PurposeOut,
			appliedRate:  sdkmath.NewInt(5),
		},
		{
			name:         "issuer_sender_ibc",
			rate:         "0.5",
			sender:       issuer,
			recipient:    dummyAddress,
			amount:       sdkmath.NewInt(10),
			ibcDirection: wibctransfertypes.PurposeOut,
			appliedRate:  sdkmath.NewInt(5),
		},
		{
			name:         "receiver_ibc",
			rate:         "0.5",
			sender:       dummyAddress,
			recipient:    accounts[0],
			amount:       sdkmath.NewInt(10),
			ibcDirection: wibctransfertypes.PurposeIn,
			appliedRate:  sdkmath.NewInt(0),
		},
		{
			name:         "ibc_escrow_sender_issuer_receiver",
			rate:         "0.5",
			sender:       dummyAddress,
			recipient:    issuer,
			amount:       sdkmath.NewInt(10),
			ibcDirection: wibctransfertypes.PurposeIn,
			appliedRate:  sdkmath.NewInt(0),
		},
		{
			name:        "smart_contract_sender",
			rate:        "0.5",
			sender:      smartContracts[0],
			recipient:   dummyAddress,
			amount:      sdkmath.NewInt(10),
			appliedRate: sdkmath.NewInt(0),
		},
		{
			name:        "issuer_to_smart_contract",
			rate:        "0.5",
			sender:      issuer,
			recipient:   smartContracts[0],
			amount:      sdkmath.NewInt(10),
			appliedRate: sdkmath.NewInt(5),
		},
		{
			name:        "smart_contract_receiver",
			rate:        "0.5",
			sender:      dummyAddress,
			recipient:   smartContracts[0],
			amount:      sdkmath.NewInt(10),
			appliedRate: sdkmath.NewInt(5),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)

			if tc.ibcDirection != "" {
				ctx = wibctransfertypes.WithPurpose(ctx, tc.ibcDirection)
			}

			if len(sdk.MustAccAddressFromBech32(tc.sender)) == wasmtypes.ContractAddrLen {
				ctx = cwasmtypes.WithSmartContractSender(ctx, tc.sender)
			}
			if len(sdk.MustAccAddressFromBech32(tc.recipient)) == wasmtypes.ContractAddrLen {
				ctx = cwasmtypes.WithSmartContractRecipient(ctx, tc.recipient)
			}

			appliedRate := assetFTKeeper.CalculateRate(
				ctx,
				sdk.MustNewDecFromStr(tc.rate),
				sdk.MustAccAddressFromBech32(tc.sender),
				sdk.MustAccAddressFromBech32(tc.recipient),
				sdk.NewCoin("test", tc.amount))
			assertT.EqualValues(tc.appliedRate.String(), appliedRate.String())
		})
	}
}
