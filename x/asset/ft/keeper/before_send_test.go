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

	"github.com/CoreumFoundation/coreum/v3/pkg/config"
	"github.com/CoreumFoundation/coreum/v3/pkg/config/constant"
	assetftkeeper "github.com/CoreumFoundation/coreum/v3/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v3/x/wibctransfer/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

type wasmKeeperMock struct {
	contracts map[string]struct{}
}

func (k wasmKeeperMock) HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool {
	_, exists := k.contracts[contractAddress.String()]
	return exists
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

	wasmKeeper := wasmKeeperMock{
		contracts: map[string]struct{}{},
	}
	for i := byte(0); i < 2; i++ {
		addr := sdk.AccAddress(bytes.Repeat([]byte{i}, wasmtypes.ContractAddrLen)).String()
		smartContracts = append(smartContracts, sdk.AccAddress(bytes.Repeat([]byte{i}, wasmtypes.ContractAddrLen)).String())
		wasmKeeper.contracts[addr] = struct{}{}
	}

	issuer := genAccount()
	dummyAddress := genAccount()
	key := sdk.NewKVStoreKey(types.StoreKey)
	assetFTKeeper := assetftkeeper.NewKeeper(nil, key, nil, nil, wasmKeeper, "")

	testCases := []struct {
		name         string
		rate         string
		sender       string
		receivers    map[string]sdkmath.Int
		ibcDirection wibctransfertypes.Direction
		appliedRate  sdkmath.Int
	}{
		{
			name:   "issuer_receiver",
			rate:   "0.5",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			appliedRate: sdkmath.ZeroInt(),
		},
		{
			name:   "issuer_sender_two_receivers",
			rate:   "0.5",
			sender: issuer,
			receivers: map[string]sdkmath.Int{
				accounts[5]: sdkmath.NewInt(5),
				accounts[6]: sdkmath.NewInt(5),
			},
			appliedRate: sdkmath.ZeroInt(),
		},
		{
			name:   "one_receiver",
			rate:   "0.1",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				accounts[10]: sdkmath.NewInt(1000),
			},
			appliedRate: sdkmath.NewInt(100),
		},
		{
			name:   "one_receiver_with_rounding",
			rate:   "0.1",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				accounts[10]: sdkmath.NewInt(1001),
			},
			appliedRate: sdkmath.NewInt(101),
		},
		{
			name:   "two_receivers_with_rounding",
			rate:   "0.1",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				accounts[10]: sdkmath.NewInt(501),
				accounts[9]:  sdkmath.NewInt(501),
			},
			appliedRate: sdkmath.NewInt(101),
		},
		{
			name:   "issuer_in_receivers",
			rate:   "0.01",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(30000),
				genAccount(): sdkmath.NewInt(20000),
				genAccount(): sdkmath.NewInt(20000),
			},
			appliedRate: sdkmath.NewInt(400),
		},
		{
			name:   "two_receiver_with_issuer_rounding",
			rate:   "0.01001",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(30000),
				genAccount(): sdkmath.NewInt(20000),
			},
			appliedRate: sdkmath.NewInt(201),
		},
		{
			name:   "four_receivers_with_issuer",
			rate:   "0.01",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(2101),
				genAccount(): sdkmath.NewInt(300),
				genAccount(): sdkmath.NewInt(1100),
				genAccount(): sdkmath.NewInt(3300),
			},
			appliedRate: sdkmath.NewInt(47),
		},
		{
			name:   "sender_ibc",
			rate:   "0.5",
			sender: accounts[0],
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeOut,
			appliedRate:  sdkmath.NewInt(5),
		},
		{
			name:   "issuer_sender_ibc",
			rate:   "0.5",
			sender: issuer,
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeOut,
			appliedRate:  sdkmath.NewInt(0),
		},
		{
			name:   "one_receiver_ibc",
			rate:   "0.5",
			sender: dummyAddress,
			receivers: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeIn,
			appliedRate:  sdkmath.NewInt(0),
		},
		{
			name:   "ibc_escrow_sender_issuer_receiver",
			rate:   "0.5",
			sender: dummyAddress,
			receivers: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeIn,
			appliedRate:  sdkmath.NewInt(0),
		},
		{
			name:   "smart_contract_sender",
			rate:   "0.5",
			sender: smartContracts[0],
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			appliedRate: sdkmath.NewInt(0),
		},
		{
			name:   "issuer_to_smart_contract",
			rate:   "0.5",
			sender: issuer,
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(10),
			},
			appliedRate: sdkmath.NewInt(0),
		},
		{
			name:   "smart_contract_receiver",
			rate:   "0.5",
			sender: dummyAddress,
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(10),
			},
			appliedRate: sdkmath.NewInt(5),
		},
		{
			name:   "sender_to_smart_contract_and_issuer",
			rate:   "0.5",
			sender: dummyAddress,
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(5),
				issuer:            sdkmath.NewInt(5),
			},
			appliedRate: sdkmath.NewInt(3),
		},
		{
			name:   "sender_to_smart_contracts",
			rate:   "0.5",
			sender: dummyAddress,
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(5),
				smartContracts[1]: sdkmath.NewInt(5),
			},
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

			appliedRate := assetFTKeeper.ApplyRate(
				ctx,
				sdk.MustNewDecFromStr(tc.rate),
				sdk.MustAccAddressFromBech32(issuer),
				sdk.MustAccAddressFromBech32(tc.sender),
				tc.receivers)
			assertT.EqualValues(tc.appliedRate.String(), appliedRate.String())
		})
	}
}
