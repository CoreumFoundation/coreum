package keeper_test

import (
	"bytes"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func TestCalculateRateShares(t *testing.T) {
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
	pow10 := func(ex int64) sdkmath.Int {
		return sdkmath.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(ex), nil))
	}
	testCases := []struct {
		name         string
		rate         string
		senders      map[string]sdkmath.Int
		receivers    map[string]sdkmath.Int
		ibcDirection wibctransfertypes.Direction
		shares       map[string]sdkmath.Int
	}{
		{
			name:    "empty_senders",
			rate:    "0.5",
			senders: map[string]sdkmath.Int{},
			shares:  map[string]sdkmath.Int{},
		},
		{
			name: "two_senders_issuer_receiver",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(5),
				accounts[1]: sdkmath.NewInt(5),
			},
			receivers: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "issuer_sender_two_receivers",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				accounts[5]: sdkmath.NewInt(5),
				accounts[6]: sdkmath.NewInt(5),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "two_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(400),
				accounts[1]: sdkmath.NewInt(600),
			},
			receivers: map[string]sdkmath.Int{
				accounts[10]: sdkmath.NewInt(1000),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(40),
				accounts[1]: sdkmath.NewInt(60),
			},
		},
		{
			name: "two_senders_one_receiver_with_rounding",
			rate: "0.1",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(399),
				accounts[1]: sdkmath.NewInt(602),
			},
			receivers: map[string]sdkmath.Int{
				accounts[10]: sdkmath.NewInt(1001),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(40),
				accounts[1]: sdkmath.NewInt(61),
			},
		},
		{
			name: "issuer_sender_and_two_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdkmath.Int{
				issuer:      sdkmath.NewInt(90),
				accounts[0]: sdkmath.NewInt(29),
				accounts[1]: sdkmath.NewInt(32),
			},
			receivers: map[string]sdkmath.Int{
				genAccount(): sdkmath.NewInt(90 + 29 + 32),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(3),
				accounts[1]: sdkmath.NewInt(4),
			},
		},
		{
			name: "two_senders_issuer_receiver_one_receiver",
			rate: "0.01",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(30000),
				accounts[1]: sdkmath.NewInt(20000),
			},
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(30000),
				genAccount(): sdkmath.NewInt(20000),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(120),
				accounts[1]: sdkmath.NewInt(80),
			},
		},
		{
			name: "two_senders_issuer_receiver_one_receiver_rounding",
			rate: "0.01001",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(30000),
				accounts[1]: sdkmath.NewInt(20000),
			},
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(30000),
				genAccount(): sdkmath.NewInt(20000),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(121),
				accounts[1]: sdkmath.NewInt(81),
			},
		},
		{
			name: "two_senders_one_receiver_rounding",
			rate: "0.1234",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(80),
				accounts[1]: sdkmath.NewInt(17),
			},
			receivers: map[string]sdkmath.Int{
				genAccount(): sdkmath.NewInt(97),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(10),
				accounts[1]: sdkmath.NewInt(3),
			},
		},
		{
			name: "three_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(1),
				accounts[1]: sdkmath.NewInt(2),
				accounts[2]: sdkmath.NewInt(9),
			},
			receivers: map[string]sdkmath.Int{
				genAccount(): sdkmath.NewInt(12),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(1),
				accounts[1]: sdkmath.NewInt(1),
				accounts[2]: sdkmath.NewInt(1),
			},
		},
		{
			name: "issuer_sender_three_senders_issuer_receiver_three_receivers",
			rate: "0.01",
			senders: map[string]sdkmath.Int{
				issuer:      sdkmath.NewInt(2100),
				accounts[0]: sdkmath.NewInt(1100),
				accounts[1]: sdkmath.NewInt(1700),
				accounts[2]: sdkmath.NewInt(1900),
			},
			receivers: map[string]sdkmath.Int{
				issuer:       sdkmath.NewInt(2100),
				genAccount(): sdkmath.NewInt(300),
				genAccount(): sdkmath.NewInt(1100),
				genAccount(): sdkmath.NewInt(3300),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(11),
				accounts[1]: sdkmath.NewInt(17),
				accounts[2]: sdkmath.NewInt(19),
			},
		},
		{
			name: "three_senders_three_receivers",
			rate: "0.01",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(100).Mul(pow10(24)),
				accounts[1]: sdkmath.NewInt(300).Mul(pow10(25)),
				accounts[2]: sdkmath.NewInt(500).Mul(pow10(26)),
			},
			receivers: map[string]sdkmath.Int{
				genAccount(): sdkmath.NewInt(100).Mul(pow10(24)),
				genAccount(): sdkmath.NewInt(300).Mul(pow10(25)),
				genAccount(): sdkmath.NewInt(500).Mul(pow10(26)),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(100).Mul(pow10(22)),
				accounts[1]: sdkmath.NewInt(300).Mul(pow10(23)),
				accounts[2]: sdkmath.NewInt(500).Mul(pow10(24)),
			},
		},
		{
			name: "issuer_sender_three_senders_four_receivers",
			rate: "0.99",
			senders: map[string]sdkmath.Int{
				issuer:      sdkmath.NewInt(2100),
				accounts[0]: sdkmath.NewInt(1100),
				accounts[1]: sdkmath.NewInt(1700),
				accounts[2]: sdkmath.NewInt(2728),
			},
			receivers: map[string]sdkmath.Int{
				genAccount(): sdkmath.NewInt(2100),
				genAccount(): sdkmath.NewInt(1000),
				genAccount(): sdkmath.NewInt(1800),
				genAccount(): sdkmath.NewInt(2728),
			},
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(1089),
				accounts[1]: sdkmath.NewInt(1683),
				accounts[2]: sdkmath.NewInt(2701),
			},
		},
		{
			name: "one_sender_ibc",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeOut,
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(5),
			},
		},
		{
			name: "issuer_sender_ibc",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeOut,
			shares:       map[string]sdkmath.Int{},
		},
		{
			name: "issuer_sender_two_senders_ibc",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer:      sdkmath.NewInt(10),
				accounts[0]: sdkmath.NewInt(10),
				accounts[1]: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(20),
			},
			ibcDirection: wibctransfertypes.PurposeOut,
			shares: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(5),
				accounts[1]: sdkmath.NewInt(5),
			},
		},
		{
			name: "one_receiver_ibc",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				accounts[0]: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeIn,
			shares:       map[string]sdkmath.Int{},
		},
		{
			name: "ibc_escrow_sender_issuer_receiver",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			ibcDirection: wibctransfertypes.PurposeIn,
			shares:       map[string]sdkmath.Int{},
		},
		{
			name: "smart_contract_to_recipient",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "smart_contract_and_issuer_to_recipient",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer:            sdkmath.NewInt(5),
				smartContracts[0]: sdkmath.NewInt(5),
			},
			receivers: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "smart_contract_and_issuer_to_smart_contract",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer:            sdkmath.NewInt(5),
				smartContracts[0]: sdkmath.NewInt(5),
			},
			receivers: map[string]sdkmath.Int{
				smartContracts[1]: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "issuer_to_smart_contract",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				issuer: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{},
		},
		{
			name: "sender_to_smart_contract",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(10),
			},
			shares: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(5),
			},
		},
		{
			name: "sender_to_smart_contract_and_issuer",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(5),
				issuer:            sdkmath.NewInt(5),
			},
			shares: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(3),
			},
		},
		{
			name: "sender_to_smart_contracts",
			rate: "0.5",
			senders: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(10),
			},
			receivers: map[string]sdkmath.Int{
				smartContracts[0]: sdkmath.NewInt(5),
				smartContracts[1]: sdkmath.NewInt(5),
			},
			shares: map[string]sdkmath.Int{
				dummyAddress: sdkmath.NewInt(5),
			},
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

			shares, err := assetFTKeeper.CalculateRateShares(ctx, sdk.MustNewDecFromStr(tc.rate), issuer, tc.senders, tc.receivers)
			require.NoError(t, err)
			for account, share := range shares {
				assertT.EqualValues(tc.shares[account].String(), share.String())
			}
		})
	}
}
