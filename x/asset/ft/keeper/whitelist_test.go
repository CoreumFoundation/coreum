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
func TestKeeper_Whitelist(t *testing.T) {
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
		Features:      []types.TokenFeature{types.TokenFeature_whitelist}, //nolint:nosnakecase
	}

	denom, err := ftKeeper.Issue(ctx, settings)
	requireT.NoError(err)

	unwhitelistableSettings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "ABC",
		Subunit:       "abc",
		Description:   "ABC Desc",
		InitialAmount: sdk.NewInt(666),
		Features:      []types.TokenFeature{},
	}

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	unwhitelistableDenom, err := ftKeeper.Issue(ctx, unwhitelistableSettings)
	requireT.NoError(err)
	_, err = ftKeeper.GetToken(ctx, unwhitelistableDenom)
	requireT.NoError(err)

	// whitelisting fails on unwhitelistable token
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(1)))
	requireT.ErrorIs(types.ErrFeatureNotActive, err)

	// try to whitelist non-existent denom
	nonExistentDenom := types.BuildDenom("nonexist", issuer)
	err = ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(nonExistentDenom, sdk.NewInt(10)))
	assertT.True(sdkerrors.IsOf(err, types.ErrFTNotFound))

	// try to whitelist from non issuer address
	randomAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(10)))
	assertT.ErrorIs(sdkerrors.ErrUnauthorized, err)

	// set whitelisted balance to 0
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(0))))
	whitelistedBalance := ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(0)).String(), whitelistedBalance.String())

	coinsToSend := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.True(types.ErrWhitelistedLimitExceeded.Is(err))

	// set whitelisted balance to 100
	requireT.NoError(ftKeeper.SetWhitelistedBalance(ctx, issuer, recipient, sdk.NewCoin(denom, sdk.NewInt(100))))
	whitelistedBalance = ftKeeper.GetWhitelistedBalance(ctx, recipient, denom)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(100)).String(), whitelistedBalance.String())

	// test query all whitelisted balances
	allBalances, pageRes, err := ftKeeper.GetAccountsWhitelistedBalances(ctx, &query.PageRequest{})
	assertT.NoError(err)
	assertT.Len(allBalances, 1)
	assertT.EqualValues(1, pageRes.GetTotal())
	assertT.EqualValues(recipient.String(), allBalances[0].Address)
	requireT.Equal(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))).String(), allBalances[0].Coins.String())

	coinsToSend = sdk.NewCoins(
		sdk.NewCoin(denom, sdk.NewInt(50)),
		sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(50)),
	)
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.NoError(err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.NoError(err)

	// try to send more
	coinsToSend = sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1)))
	// send
	err = bankKeeper.SendCoins(ctx, issuer, recipient, coinsToSend)
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)
	// multi-send
	err = bankKeeper.InputOutputCoins(ctx,
		[]banktypes.Input{{Address: issuer.String(), Coins: coinsToSend}},
		[]banktypes.Output{{Address: recipient.String(), Coins: coinsToSend}})
	requireT.ErrorIs(types.ErrWhitelistedLimitExceeded, err)

	// try to whitelist from non issuer address
	err = ftKeeper.SetWhitelistedBalance(ctx, randomAddr, recipient, sdk.NewCoin(denom, sdk.NewInt(80)))
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))
}
