package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Description:   "ABC Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.TokenFeature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to freeze non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdk.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrFTNotFound))

	// try to freeze unfreezable FT
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(unfreezableDenom, sdk.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrFeatureNotActive))

	// try to freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.Freeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(10)))
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// try to freeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.ErrorIs(sdkerrors.ErrInvalidCoins, err)

	// try to unfreeze 0 balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.ErrorIs(sdkerrors.ErrInvalidCoins, err)

	// try to freeze more than balance
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance := ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// try to unfreeze more than frozen balance
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(130)))
	requireT.ErrorIs(types.ErrNotEnoughBalance, err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// set frozen balance back to zero
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(0)).String(), frozenBalance.String())

	// freeze, query frozen
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(40)).String(), frozenBalance.String())

	// test query all frozen
	allBalances, pageRes, err := ftKeeper.GetAccountsFrozenBalances(ctx, &query.PageRequest{})
	assertT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(40))).String(), allBalances[0].Coins.String())

	// increase frozen and query
	err = ftKeeper.Freeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), frozenBalance)

	// try to send more than available
	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(80)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, issuer, coinsToSend)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: issuer.String(), Coins: coinsToSend}})
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// try to send unfrozen balance
	recipient2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(90)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(10)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(20)), balance)

	// try to unfreeze from non issuer address
	err = ftKeeper.Unfreeze(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// unfreeze, query frozen, and try to send
	err = ftKeeper.Unfreeze(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.NoError(err)
	frozenBalance = ftKeeper.GetFrozenBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), frozenBalance)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(40)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, recipient2, coinsToSend)
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(40)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(60)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient2.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, recipient2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)), balance)
}
