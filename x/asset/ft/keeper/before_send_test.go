package keeper_test

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	assetftkeeper "github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	wibctransfertypes "github.com/CoreumFoundation/coreum/x/wibctransfer/types"
)

func TestMain(m *testing.M) {
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	n.SetSDKConfig()
	m.Run()
}

func TestCalculateRateShares(t *testing.T) {
	genAccount := func() string {
		return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	}
	var accounts []string
	for i := 0; i < 11; i++ {
		accounts = append(accounts, genAccount())
	}
	issuer := genAccount()
	dummyAddress := genAccount()
	assetFTKeeper := assetftkeeper.NewKeeper(nil, nil, nil, nil, nil)
	pow10 := func(ex int64) sdk.Int {
		return sdk.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(ex), nil))
	}
	testCases := []struct {
		name         string
		rate         string
		senders      map[string]sdk.Int
		receivers    map[string]sdk.Int
		ibcDirection wibctransfertypes.Direction
		shares       map[string]sdk.Int
	}{
		{
			name:    "empty_senders",
			rate:    "0.5",
			senders: map[string]sdk.Int{},
			shares:  map[string]sdk.Int{},
		},
		{
			name: "two_senders_issuer_receiver",
			rate: "0.5",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(5),
				accounts[1]: sdk.NewInt(5),
			},
			receivers: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			shares: map[string]sdk.Int{},
		},
		{
			name: "issuer_sender_two_receivers",
			rate: "0.5",
			senders: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				accounts[5]: sdk.NewInt(5),
				accounts[6]: sdk.NewInt(5),
			},
			shares: map[string]sdk.Int{},
		},
		{
			name: "two_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(400),
				accounts[1]: sdk.NewInt(600),
			},
			receivers: map[string]sdk.Int{
				accounts[10]: sdk.NewInt(1000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(40),
				accounts[1]: sdk.NewInt(60),
			},
		},
		{
			name: "two_senders_one_receiver_with_rounding",
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(399),
				accounts[1]: sdk.NewInt(602),
			},
			receivers: map[string]sdk.Int{
				accounts[10]: sdk.NewInt(1001),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(40),
				accounts[1]: sdk.NewInt(61),
			},
		},
		{
			name: "issuer_sender_and_two_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(90),
				accounts[0]: sdk.NewInt(29),
				accounts[1]: sdk.NewInt(32),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(90 + 29 + 32),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(3),
				accounts[1]: sdk.NewInt(4),
			},
		},
		{
			name: "two_senders_issuer_receiver_one_receiver",
			rate: "0.01",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(30000),
				accounts[1]: sdk.NewInt(20000),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(30000),
				genAccount(): sdk.NewInt(20000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(120),
				accounts[1]: sdk.NewInt(80),
			},
		},
		{
			name: "two_senders_issuer_receiver_one_receiver_rounding",
			rate: "0.01001",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(30000),
				accounts[1]: sdk.NewInt(20000),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(30000),
				genAccount(): sdk.NewInt(20000),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(121),
				accounts[1]: sdk.NewInt(81),
			},
		},
		{
			name: "two_senders_one_receiver_rounding",
			rate: "0.1234",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(80),
				accounts[1]: sdk.NewInt(17),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(97),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(10),
				accounts[1]: sdk.NewInt(3),
			},
		},
		{
			name: "three_senders_one_receiver",
			rate: "0.1",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1),
				accounts[1]: sdk.NewInt(2),
				accounts[2]: sdk.NewInt(9),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(12),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1),
				accounts[1]: sdk.NewInt(1),
				accounts[2]: sdk.NewInt(1),
			},
		},
		{
			name: "issuer_sender_three_senders_issuer_receiver_three_receivers",
			rate: "0.01",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(2100),
				accounts[0]: sdk.NewInt(1100),
				accounts[1]: sdk.NewInt(1700),
				accounts[2]: sdk.NewInt(1900),
			},
			receivers: map[string]sdk.Int{
				issuer:       sdk.NewInt(2100),
				genAccount(): sdk.NewInt(300),
				genAccount(): sdk.NewInt(1100),
				genAccount(): sdk.NewInt(3300),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(11),
				accounts[1]: sdk.NewInt(17),
				accounts[2]: sdk.NewInt(19),
			},
		},
		{
			name: "three_senders_three_receivers",
			rate: "0.01",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(100).Mul(pow10(24)),
				accounts[1]: sdk.NewInt(300).Mul(pow10(25)),
				accounts[2]: sdk.NewInt(500).Mul(pow10(26)),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(100).Mul(pow10(24)),
				genAccount(): sdk.NewInt(300).Mul(pow10(25)),
				genAccount(): sdk.NewInt(500).Mul(pow10(26)),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(100).Mul(pow10(22)),
				accounts[1]: sdk.NewInt(300).Mul(pow10(23)),
				accounts[2]: sdk.NewInt(500).Mul(pow10(24)),
			},
		},
		{
			name: "issuer_sender_three_senders_four_receivers",
			rate: "0.99",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(2100),
				accounts[0]: sdk.NewInt(1100),
				accounts[1]: sdk.NewInt(1700),
				accounts[2]: sdk.NewInt(2728),
			},
			receivers: map[string]sdk.Int{
				genAccount(): sdk.NewInt(2100),
				genAccount(): sdk.NewInt(1000),
				genAccount(): sdk.NewInt(1800),
				genAccount(): sdk.NewInt(2728),
			},
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(1089),
				accounts[1]: sdk.NewInt(1683),
				accounts[2]: sdk.NewInt(2701),
			},
		},
		{
			name: "one_sender_ibc",
			rate: "0.5",
			senders: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				dummyAddress: sdk.NewInt(10),
			},
			ibcDirection: wibctransfertypes.DirectionOut,
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(5),
			},
		},
		{
			name: "issuer_sender_ibc",
			rate: "0.5",
			senders: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				dummyAddress: sdk.NewInt(10),
			},
			ibcDirection: wibctransfertypes.DirectionOut,
			shares:       map[string]sdk.Int{},
		},
		{
			name: "issuer_sender_two_senders_ibc",
			rate: "0.5",
			senders: map[string]sdk.Int{
				issuer:      sdk.NewInt(10),
				accounts[0]: sdk.NewInt(10),
				accounts[1]: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				dummyAddress: sdk.NewInt(20),
			},
			ibcDirection: wibctransfertypes.DirectionOut,
			shares: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(5),
				accounts[1]: sdk.NewInt(5),
			},
		},
		{
			name: "one_receiver_ibc",
			rate: "0.5",
			senders: map[string]sdk.Int{
				dummyAddress: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				accounts[0]: sdk.NewInt(10),
			},
			ibcDirection: wibctransfertypes.DirectionIn,
			shares:       map[string]sdk.Int{},
		},
		{
			name: "ibc_escrow_sender_issuer_receiver",
			rate: "0.5",
			senders: map[string]sdk.Int{
				dummyAddress: sdk.NewInt(10),
			},
			receivers: map[string]sdk.Int{
				issuer: sdk.NewInt(10),
			},
			ibcDirection: wibctransfertypes.DirectionIn,
			shares:       map[string]sdk.Int{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			assertT := assert.New(t)
			ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)

			if tc.ibcDirection != "" {
				ctx = wibctransfertypes.WithDirection(ctx, tc.ibcDirection)
			}

			shares := assetFTKeeper.CalculateRateShares(ctx, sdk.MustNewDecFromStr(tc.rate), issuer, tc.senders, tc.receivers)
			for account, share := range shares {
				assertT.EqualValues(tc.shares[account].String(), share.String())
			}
		})
	}
}
