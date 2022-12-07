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
		rate     float32
		input    int64
		expected int64
	}{
		{
			rate:     0,
			input:    1,
			expected: 0,
		},
		{
			rate:     0.01,
			input:    1,
			expected: 1,
		},
		{
			rate:     0.01,
			input:    101,
			expected: 2,
		},
		{
			rate:     0.01,
			input:    100,
			expected: 1,
		},
		{
			rate:     0.1,
			input:    100,
			expected: 10,
		},
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			definition := types.FungibleTokenDefinition{
				BurnRate: tc.rate,
			}
			burnCoin := definition.CalculateBurnCoin(sdk.NewCoin("test", sdk.NewInt(tc.input)))
			assertT.EqualValues(sdk.NewInt(tc.expected).String(), burnCoin.Amount.String())
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
		BurnRate:      1.01,
	}

	_, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.Error(err)
	requireT.True(types.ErrInvalidBurnRate.Is(err))

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
		BurnRate:      0.25,
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)

	receiver := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// send from issuer to recipient (burn must not apply)
	err = bankKeeper.SendCoins(ctx, issuer, receiver, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(500)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&receiver: 500,
		&issuer:   100,
	})

	// send from issuer to recipient to another account (burn must apply)
	receiver2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, receiver, receiver2, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&receiver:  375,
		&receiver2: 100,
		&issuer:    100,
	})

	// send from recipient to issuer account (burn must not apply)
	err = bankKeeper.SendCoins(ctx, receiver, issuer, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(375)),
	))
	requireT.NoError(err)

	ba.assertCoinDistribution(denom, map[*sdk.AccAddress]int64{
		&receiver2: 100,
		&issuer:    475,
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
