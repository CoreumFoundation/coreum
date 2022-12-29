package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
	wbankkeeper "github.com/CoreumFoundation/coreum/x/wbank/keeper"
)

//nolint:funlen // there are too many tests cases
func TestKeeperCalculateBurnRateShare(t *testing.T) {
	testCases := []struct {
		rate      string
		inputSum  int64
		outputSum int64
		senders   map[string]int64
		shares    map[string]int64
	}{
		{
			rate:      "0.5",
			inputSum:  0,
			outputSum: 0,
			senders:   map[string]int64{},
			shares:    map[string]int64{},
		},
		{
			rate:      "0.5",
			inputSum:  10,
			outputSum: 0,
			senders: map[string]int64{
				"1": 5,
				"2": 5,
			},
			shares: map[string]int64{},
		},
		{
			rate:      "0.1",
			inputSum:  1000,
			outputSum: 2000,
			senders: map[string]int64{
				"1": 400,
				"2": 600,
			},
			shares: map[string]int64{
				"1": 40,
				"2": 60,
			},
		},
		{
			rate:      "0.1",
			inputSum:  1001,
			outputSum: 2000,
			senders: map[string]int64{
				"1": 399,
				"2": 602,
			},
			shares: map[string]int64{
				"1": 40,
				"2": 61,
			},
		},
		{
			rate:      "0.01",
			inputSum:  50000,
			outputSum: 20000,
			senders: map[string]int64{
				"1": 30000,
				"2": 20000,
			},
			shares: map[string]int64{
				"1": 120,
				"2": 80,
			},
		},
		{
			rate:      "0.01001",
			inputSum:  50000,
			outputSum: 20000,
			senders: map[string]int64{
				"1": 30000,
				"2": 20000,
			},
			shares: map[string]int64{
				"1": 121,
				"2": 81,
			},
		},
		{
			rate:      "0.1234",
			inputSum:  97,
			outputSum: 97,
			senders: map[string]int64{
				"1": 80,
				"2": 17,
			},
			shares: map[string]int64{
				"1": 10,
				"2": 3,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			senders := map[string]sdk.Int{}
			for addr, amt := range tc.senders {
				senders[addr] = sdk.NewInt(amt)
			}
			si := keeper.MultiSendIterationInfo{
				FT:                 types.FTDefinition{BurnRate: sdk.MustNewDecFromStr(tc.rate)},
				NonIssuerInputSum:  sdk.NewInt(tc.inputSum),
				NonIssuerOutputSum: sdk.NewInt(tc.outputSum),
				NonIssuerSenders:   senders,
			}
			shares := si.CalculateBurnRateShares()
			for acc, share := range shares {
				assertT.EqualValues(tc.shares[acc], share.Int64())
			}
		})
	}
}

func TestKeeper_BurnRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue with more than 1 burn rate
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(600),
		Features:      []types.TokenFeature{},
		BurnRate:      sdk.MustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.Error(err)
	requireT.True(types.ErrInvalidInput.Is(err))

	// issue token
	settings = types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(600),
		Features:      []types.TokenFeature{},
		BurnRate:      sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (burn must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (burn must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     100,
	})

	// send from recipient to issuer account (burn must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     475,
	})
}

//nolint:funlen // This is a complex test scenario and breaking it down will make it harder to read
func TestKeeper_BurnRate_BankMultiSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue 3 tokens
	var recipients []sdk.AccAddress
	var issuers []sdk.AccAddress
	var denoms []string
	for i := 0; i < 2; i++ {
		issuers = append(issuers, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		settings := types.IssueSettings{
			Issuer:        issuers[i],
			Symbol:        fmt.Sprintf("DEF%d", i),
			Subunit:       fmt.Sprintf("def%d", i),
			Precision:     6,
			Description:   "DEF Desc",
			InitialAmount: sdk.NewInt(1000),
			Features:      []types.TokenFeature{},
			BurnRate:      sdk.MustNewDecFromStr(fmt.Sprintf("0.%d", i+1)),
		}

		denom, err := assetKeeper.Issue(ctx, settings)
		requireT.NoError(err)
		denoms = append(denoms, denom)

		// create 2 recipient for every issuer to allow for complex test cases
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
		recipients = append(recipients, sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()))
	}

	var testCases = []struct {
		name         string
		inputs       []banktypes.Input
		outputs      []banktypes.Output
		distribution map[string]map[*sdk.AccAddress]int64
	}{
		{
			name: "send from issuer to other accounts",
			inputs: []banktypes.Input{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(200)))},
				{Address: issuers[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(200)))},
			},
			outputs: []banktypes.Output{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(100)),
					sdk.NewCoin(denoms[1], sdk.NewInt(100)),
				)},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(100)),
					sdk.NewCoin(denoms[1], sdk.NewInt(100)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    800,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
				denoms[1]: {
					&issuers[1]:    800,
					&recipients[0]: 100,
					&recipients[1]: 100,
				},
			},
		},
		{
			name: "include issuer in senders",
			inputs: []banktypes.Input{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(90)))},
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(29)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(32)))},
			},
			outputs: []banktypes.Output{
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(89)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(62)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]:    710,
					&recipients[0]: 68, // 100 - 29 - 3 (burn = roundup(29 * 10%))
					&recipients[1]: 64, // 100 - 32 - 4 (burn = roundup(32 * 10%))
					&recipients[2]: 89,
					&recipients[3]: 62,
				},
			},
		},
		{
			name: "include issuer in receivers",
			inputs: []banktypes.Input{
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(60)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(40)))},
			},
			outputs: []banktypes.Output{
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(40)),
				)},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(25)),
				)},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(35)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[1]: {
					&issuers[1]:    840,
					&recipients[0]: 32, // 100 - 60 - 8 (burn = roundup(60 * (60/100) * 20%))
					&recipients[1]: 55, // 100 - 40 - 5 (burn = roundup(40 * (60/100) * 20%))
					&recipients[2]: 25,
					&recipients[3]: 35,
				},
			},
		},
		{
			name: "send all coins back to issuers",
			inputs: []banktypes.Input{
				// coin[0]
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(68)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(64)))},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(89)))},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(62)))},
				// coin[1]
				{Address: recipients[0].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(32)))},
				{Address: recipients[1].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(55)))},
				{Address: recipients[2].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(25)))},
				{Address: recipients[3].String(), Coins: sdk.NewCoins(sdk.NewCoin(denoms[1], sdk.NewInt(35)))},
			},
			outputs: []banktypes.Output{
				{Address: issuers[0].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[0], sdk.NewInt(283)),
				)},
				{Address: issuers[1].String(), Coins: sdk.NewCoins(
					sdk.NewCoin(denoms[1], sdk.NewInt(147)),
				)},
			},
			distribution: map[string]map[*sdk.AccAddress]int64{
				denoms[0]: {
					&issuers[0]: 993,
				},
				denoms[1]: {
					&issuers[1]: 987,
				},
			},
		},
	}

	for counter, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s case #%d", tc.name, counter), func(t *testing.T) {
			err := bankKeeper.InputOutputCoins(ctx, tc.inputs, tc.outputs)
			requireT.NoError(err)

			for denom, dist := range tc.distribution {
				ba.assertCoinDistribution(denom, dist)
			}
		})
	}
}

type bankAssertion struct {
	t   require.TestingT
	bk  wbankkeeper.BaseKeeperWrapper
	ctx sdk.Context
}

func newBankAsserter(
	ctx sdk.Context,
	t require.TestingT,
	bk wbankkeeper.BaseKeeperWrapper,
) bankAssertion {
	return bankAssertion{
		t:   t,
		bk:  bk,
		ctx: ctx,
	}
}

func (ba bankAssertion) assertCoinDistribution(denom string, dist map[*sdk.AccAddress]int64) {
	requireT := require.New(ba.t)
	total := int64(0)
	for acc, expectedBalance := range dist {
		total += expectedBalance
		getBalance := ba.bk.GetBalance(ba.ctx, *acc, denom)
		requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(expectedBalance)).String(), getBalance.String())
	}

	totalSupply := ba.bk.GetSupply(ba.ctx, denom)
	requireT.Equal(totalSupply.String(), sdk.NewCoin(denom, sdk.NewInt(total)).String())
}
