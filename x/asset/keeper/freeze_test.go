package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

//nolint:funlen // this is complex test scenario and breaking it down is not helpful
func TestKeeper_FreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	settings := types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(666),
		Features:      []types.FungibleTokenFeature{types.FungibleTokenFeature_freeze}, //nolint:nosnakecase
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Description:   "ABC Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(666),
		Features:      []types.FungibleTokenFeature{},
	}

	unfreezableDenom, err := assetKeeper.IssueFungibleToken(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = assetKeeper.GetFungibleToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	receiver := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, receiver, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to freeze non-existent denom
	nonExistentDenom := types.BuildFungibleTokenDenom("nonexist", issuer)
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(nonExistentDenom, sdk.NewInt(10)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, types.ErrFungibleTokenNotFound))

	// try to freeze unfreezable FT
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(unfreezableDenom, sdk.NewInt(10)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, types.ErrFeatureNotActive))

	// try to freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = assetKeeper.FreezeFungibleToken(ctx, randomAddr, receiver, sdk.NewCoin(denom, sdk.NewInt(10)))
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// try to freeze 0 balance
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.True(sdkerrors.ErrInvalidCoins.Is(err))

	// try to unfreeze 0 balance
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(0)))
	requireT.True(sdkerrors.ErrInvalidCoins.Is(err))

	// try to freeze more than balance
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance := assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// try to unfreeze more than frozen balance
	err = assetKeeper.UnfreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(130)))
	requireT.True(types.ErrNotEnoughBalance.Is(err))
	frozenBalance = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(110)), frozenBalance)

	// set frozen balance back to zero
	err = assetKeeper.UnfreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.NoError(err)
	frozenBalance = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	assertT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(0)).String(), frozenBalance.String())

	// freeze, query frozen
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(40)).String(), frozenBalance.String())

	// test query all frozen
	allBalances, pageRes, err := assetKeeper.GetAccountsFrozenBalances(ctx, &query.PageRequest{})
	assertT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(receiver.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(40))).String(), allBalances[0].Coins.String())

	// increase frozen and query
	err = assetKeeper.FreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(40)))
	requireT.NoError(err)
	frozenBalance = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), frozenBalance)

	// try to send more than available
	err = bankKeeper.SendCoins(ctx, receiver, issuer, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(80))))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// try to send unfrozen balance
	receiver2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, receiver, receiver2, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(20))))
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(80)), balance)
	balance = bankKeeper.GetBalance(ctx, receiver2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(20)), balance)

	// try to unfreeze from non issuer address
	err = assetKeeper.UnfreezeFungibleToken(ctx, randomAddr, receiver, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// unfreeze, query frozen, and try to send
	err = assetKeeper.UnfreezeFungibleToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.NoError(err)
	frozenBalance = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), frozenBalance)
	err = bankKeeper.SendCoins(ctx, receiver, receiver2, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(80))))
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, receiver2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)), balance)
}
