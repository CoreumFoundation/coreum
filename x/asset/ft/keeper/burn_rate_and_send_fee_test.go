package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		burnRate         string
		commissionRate   string
		inputSum         int64
		outputSum        int64
		senders          map[string]int64
		burnShares       map[string]int64
		commissionShares map[string]int64
	}{
		{
			burnRate:         "0.5",
			commissionRate:   "0.5",
			inputSum:         0,
			outputSum:        0,
			senders:          map[string]int64{},
			burnShares:       map[string]int64{},
			commissionShares: map[string]int64{},
		},
		{
			burnRate:       "0.5",
			commissionRate: "0.5",
			inputSum:       10,
			outputSum:      0,
			senders: map[string]int64{
				"1": 5,
				"2": 5,
			},
			burnShares:       map[string]int64{},
			commissionShares: map[string]int64{},
		},
		{
			burnRate:       "0.1",
			commissionRate: "0.1",
			inputSum:       1000,
			outputSum:      2000,
			senders: map[string]int64{
				"1": 400,
				"2": 600,
			},
			burnShares: map[string]int64{
				"1": 40,
				"2": 60,
			},
			commissionShares: map[string]int64{
				"1": 40,
				"2": 60,
			},
		},
		{
			burnRate:       "0.1",
			commissionRate: "0.1",
			inputSum:       1001,
			outputSum:      2000,
			senders: map[string]int64{
				"1": 399,
				"2": 602,
			},
			burnShares: map[string]int64{
				"1": 40,
				"2": 61,
			},
			commissionShares: map[string]int64{
				"1": 40,
				"2": 61,
			},
		},
		{
			burnRate:       "0.01",
			commissionRate: "0.01",
			inputSum:       50000,
			outputSum:      20000,
			senders: map[string]int64{
				"1": 30000,
				"2": 20000,
			},
			burnShares: map[string]int64{
				"1": 120,
				"2": 80,
			},
			commissionShares: map[string]int64{
				"1": 120,
				"2": 80,
			},
		},
		{
			burnRate:       "0.01001",
			commissionRate: "0.01001",
			inputSum:       50000,
			outputSum:      20000,
			senders: map[string]int64{
				"1": 30000,
				"2": 20000,
			},
			burnShares: map[string]int64{
				"1": 121,
				"2": 81,
			},
			commissionShares: map[string]int64{
				"1": 121,
				"2": 81,
			},
		},
		{
			burnRate:       "0.1234",
			commissionRate: "0.1234",
			inputSum:       97,
			outputSum:      97,
			senders: map[string]int64{
				"1": 80,
				"2": 17,
			},
			burnShares: map[string]int64{
				"1": 10,
				"2": 3,
			},
			commissionShares: map[string]int64{
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
				FT: types.FTDefinition{
					BurnRate:           sdk.MustNewDecFromStr(tc.burnRate),
					SendCommissionRate: sdk.MustNewDecFromStr(tc.commissionRate),
				},
				NonIssuerInputSum:  sdk.NewInt(tc.inputSum),
				NonIssuerOutputSum: sdk.NewInt(tc.outputSum),
				NonIssuerSenders:   senders,
			}
			burnShares, commissionShares := si.CalculateBurnRateShares()
			for acc, share := range burnShares {
				assertT.EqualValues(tc.burnShares[acc], share.Int64())
			}

			for acc, share := range commissionShares {
				assertT.EqualValues(tc.commissionShares[acc], share.Int64())
			}
		})
	}
}

//nolint:dupl // We don't care
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
	requireT.ErrorIs(types.ErrInvalidInput, err)

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

//nolint:dupl // We don't care
func TestKeeper_SendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue with more than 1 send commission rate
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.TokenFeature{},
		SendCommissionRate: sdk.MustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.Issue(ctx, settings)
	requireT.ErrorIs(types.ErrInvalidInput, err)

	// issue token
	settings = types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.TokenFeature{},
		SendCommissionRate: sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (send commission rate must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  375,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (send commission rate must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     500,
	})
}

func TestKeeper_BurnRateAndSendCommissionRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue token
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueSettings{
		Issuer:             issuer,
		Symbol:             "DEF",
		Subunit:            "def",
		Precision:          6,
		Description:        "DEF Desc",
		InitialAmount:      sdk.NewInt(600),
		Features:           []types.TokenFeature{},
		BurnRate:           sdk.MustNewDecFromStr("0.5"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (fees must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient: 500,
		&issuer:    100,
	})

	// send from recipient1 to recipient2 (fees must apply)
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient:  325,
		&recipient2: 100,
		&issuer:     125,
	})

	// send from recipient to issuer account (fees must not apply)
	err = bankKeeper.SendCoins(ctx, recipient, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(325)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&recipient2: 100,
		&issuer:     450,
	})
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
