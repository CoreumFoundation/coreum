package keeper_test

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestKeeper_IssueFungibleToken(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "BTC",
		Description:   "BTC Desc",
		Recipient:     addr,
		InitialAmount: sdk.NewInt(777),
		Options:       []types.FungibleTokenOption{types.FungibleTokenOption_Freezable}, //nolint:nosnakecase
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildFungibleTokenDenom(settings.Symbol, settings.Issuer), denom)

	gotToken, err := assetKeeper.GetFungibleToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.FungibleToken{
		Denom:       denom,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Description: settings.Description,
		Options:     []types.FungibleTokenOption{types.FungibleTokenOption_Freezable}, //nolint:nosnakecase
	}, gotToken)

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denom)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        denom,
		Symbol:      settings.Symbol,
		Description: settings.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		},
		Base:    denom,
		Display: denom,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, addr, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// issue one more time check the double issue validation
	_, err = assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.True(errors.Is(types.ErrInvalidFungibleToken, err))
}

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
		Symbol:        "ETH",
		Description:   "ETH Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(666),
		Options:       []types.FungibleTokenOption{types.FungibleTokenOption_Freezable}, //nolint:nosnakecase
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueFungibleTokenSettings{
		Issuer:        issuer,
		Symbol:        "BTC",
		Description:   "BTC Desc",
		Recipient:     issuer,
		InitialAmount: sdk.NewInt(666),
		Options:       []types.FungibleTokenOption{},
	}

	unFreezableDenom, err := assetKeeper.IssueFungibleToken(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = assetKeeper.GetFungibleToken(ctx, unFreezableDenom)
	requireT.NoError(err)

	receiver := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, receiver, sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(100)),
		sdk.NewCoin(unFreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to freeze unFreezable token
	err = assetKeeper.FreezeToken(ctx, issuer, receiver, sdk.NewCoin(unFreezableDenom, sdk.NewInt(10)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, types.ErrOptionNotActive))

	// try to freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = assetKeeper.FreezeToken(ctx, randomAddr, receiver, sdk.NewCoin(unFreezableDenom, sdk.NewInt(10)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// try to freeze more than balance
	err = assetKeeper.FreezeToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(110)))
	requireT.Error(err)

	// freeze, query frozen
	freezeAmount := sdk.NewCoin(denom, sdk.NewInt(80))
	err = assetKeeper.FreezeToken(ctx, issuer, receiver, freezeAmount)
	requireT.NoError(err)
	frozen := assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	requireT.Equal(freezeAmount, frozen)

	// try to send more than frozen
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
	err = assetKeeper.UnfreezeToken(ctx, randomAddr, receiver, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// try to unfreeze more than frozen
	err = assetKeeper.UnfreezeToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(90)))
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// unfreeze, query frozen, and try to send
	err = assetKeeper.UnfreezeToken(ctx, issuer, receiver, sdk.NewCoin(denom, sdk.NewInt(80)))
	requireT.NoError(err)
	frozen = assetKeeper.GetFrozenBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), frozen)
	err = bankKeeper.SendCoins(ctx, receiver, receiver2, sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(80))))
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, receiver, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)), balance)
	balance = bankKeeper.GetBalance(ctx, receiver2, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)), balance)
}
