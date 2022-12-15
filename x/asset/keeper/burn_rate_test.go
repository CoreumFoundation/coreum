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
	"github.com/CoreumFoundation/coreum/x/asset/types"
	"github.com/CoreumFoundation/coreum/x/wbank/keeper"
)

func TestKeeperCalculateBurnCoin(t *testing.T) {
	testCases := []struct {
		rate       string
		sendAmount int64
		burnAmount int64
	}{
		{
			rate:       "0.5",
			sendAmount: 0,
			burnAmount: 0,
		},
		{
			rate:       "0",
			sendAmount: 1,
			burnAmount: 0,
		},
		{
			rate:       "0.01",
			sendAmount: 1,
			burnAmount: 1,
		},
		{
			rate:       "0.01",
			sendAmount: 101,
			burnAmount: 2,
		},
		{
			rate:       "0.01",
			sendAmount: 100,
			burnAmount: 1,
		},
		{
			rate:       "0.1",
			sendAmount: 100,
			burnAmount: 10,
		},
		{
			rate:       "1.0",
			sendAmount: 73,
			burnAmount: 73,
		},
		{
			rate:       "0.1234",
			sendAmount: 97,
			burnAmount: 12,
		},
		{
			rate:       "0.0003",
			sendAmount: 492301,
			burnAmount: 148,
		},
		{
			rate:       "0.0103",
			sendAmount: 492301,
			burnAmount: 5071,
		},
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			definition := types.FungibleTokenDefinition{
				BurnRate: sdk.MustNewDecFromStr(tc.rate),
			}
			burnCoin := definition.CalculateBurnRateAmount(sdk.NewCoin("test", sdk.NewInt(tc.sendAmount)))
			assertT.EqualValues(sdk.NewInt(tc.burnAmount).String(), burnCoin.String())
		})
	}
}

func TestKeeper_BurnRate_BankSend(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper
	ba := newBankAsserter(ctx, t, bankKeeper)

	// issue with more than 1 burn rate
	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	settings := types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(600),
		Features:      []types.FungibleTokenFeature{},
		BurnRate:      sdk.MustNewDecFromStr("1.01"),
	}

	_, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.Error(err)
	requireT.True(types.ErrInvalidInput.Is(err))

	// issue token
	settings = types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     6,
		Description:   "DEF Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(600),
		Features:      []types.FungibleTokenFeature{},
		BurnRate:      sdk.MustNewDecFromStr("0.25"),
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
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

type bankAssertion struct {
	t   require.TestingT
	bk  keeper.BaseKeeperWrapper
	ctx sdk.Context
}

func newBankAsserter(
	ctx sdk.Context,
	t require.TestingT,
	bk keeper.BaseKeeperWrapper,
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
