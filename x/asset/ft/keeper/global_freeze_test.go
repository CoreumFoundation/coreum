package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

//nolint:funlen // this is complex tests scenario and breaking it down is not beneficial
func TestKeeper_GlobalFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	freezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Desc",
		InitialAmount: sdk.NewInt(777),
		Features:      []types.TokenFeature{types.TokenFeature_freeze}, //nolint:nosnakecase
	}

	freezableDenom, err := ftKeeper.Issue(ctx, freezableSettings)
	requireT.NoError(err)

	unfreezableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "NOFREEZE",
		Subunit:       "nofreeze",
		Precision:     6,
		Description:   "NOFREEZE Desc",
		InitialAmount: sdk.NewInt(777),
		Features:      []types.TokenFeature{},
	}

	unfreezableDenom, err := ftKeeper.Issue(ctx, unfreezableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unfreezableDenom)
	requireT.NoError(err)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(
		sdk.NewCoin(freezableDenom, sdk.NewInt(100)),
		sdk.NewCoin(unfreezableDenom, sdk.NewInt(100)),
	))
	requireT.NoError(err)

	// try to global-freeze non-existent
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.GloballyFreeze(ctx, issuer, nonExistentDenom)
	assertT.True(sdkerrors.IsOf(err, types.ErrTokenNotFound))

	// try to global-freeze unfreezable FT
	err = ftKeeper.GloballyFreeze(ctx, issuer, unfreezableDenom)
	assertT.ErrorIs(types.ErrFeatureDisabled, err)

	// try to global-freeze from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.GloballyFreeze(ctx, randomAddr, freezableDenom)
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// freeze twice to check global-freeze idempotence
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	frozenFt, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.True(frozenFt.GloballyFrozen)

	// try to global-unfreeze from non issuer address
	err = ftKeeper.GloballyUnfreeze(ctx, randomAddr, freezableDenom)
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// unfreeze twice to check global-unfreeze idempotence
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	unfrozenFt, err := ftKeeper.GetToken(ctx, freezableDenom)
	requireT.NoError(err)
	assertT.False(unfrozenFt.GloballyFrozen)

	// freeze, try to send & verify balance
	err = ftKeeper.GloballyFreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	coinsToSend := sdk.NewCoins(sdk.NewCoin(freezableDenom, sdk.NewInt(10)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	assertT.ErrorIs(types.ErrGloballyFrozen, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	assertT.ErrorIs(types.ErrGloballyFrozen, err)

	// unfreeze, try to send & verify balance
	err = ftKeeper.GloballyUnfreeze(ctx, issuer, freezableDenom)
	requireT.NoError(err)
	coinsToSend = sdk.NewCoins(sdk.NewCoin(freezableDenom, sdk.NewInt(6)))
	// send
	err = bankKeeper.SendCoins(ctx, recipient, randomAddr, coinsToSend)
	requireT.NoError(err)
	balance := bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdk.NewInt(6)), balance)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: recipient.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: randomAddr.String(), Coins: coinsToSend}})
	requireT.NoError(err)
	balance = bankKeeper.GetBalance(ctx, randomAddr, freezableDenom)
	requireT.Equal(sdk.NewCoin(freezableDenom, sdk.NewInt(12)), balance)
}
