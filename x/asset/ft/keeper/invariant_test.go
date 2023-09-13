package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

func TestFrozenBalancesInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []types.Feature{
			types.Feature_freezing,
		},
	}

	denom1, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	requireT.NoError(err)

	// check that current state is valid
	_, isBroken := keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break frozen state and check
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.Coins{sdk.Coin{
		Denom:  denom1,
		Amount: sdkmath.NewInt(-1),
	}})
	_, isBroken = keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)

	// make the state valid
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	_, isBroken = keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// make the state valid
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	_, isBroken = keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	settings2 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF2",
		Subunit:       "def2",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		// the freezing and whitelisting disabled
		Features: []types.Feature{
			types.Feature_minting,
		},
	}

	denom2, err := ftKeeper.Issue(ctx, settings2)
	requireT.NoError(err)

	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom2, 10)))
	_, isBroken = keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)
	// make the state valid (we use the slice here since the `sdk.NewCoins` sanitizes the empty coins)
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.Coins{sdk.NewInt64Coin(denom2, 0)})
	_, isBroken = keeper.FreezingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)
}

func TestWhitelistedBalancesInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper
	bankKeeper := testApp.BankKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []types.Feature{
			types.Feature_whitelisting,
		},
	}

	denom1, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)))
	requireT.NoError(err)

	// check that current state is valid
	_, isBroken := keeper.WhitelistingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break whitelisted state and check
	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.Coins{sdk.Coin{
		Denom:  denom1,
		Amount: sdkmath.NewInt(-1),
	}})
	_, isBroken = keeper.WhitelistingInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)

	// make the state valid
	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.Coins{sdk.Coin{
		Denom:  denom1,
		Amount: sdkmath.NewInt(0),
	}})
	_, isBroken = keeper.WhitelistingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	settings2 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF2",
		Subunit:       "def2",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		// the freezing and whitelisting disabled
		Features: []types.Feature{
			types.Feature_minting,
		},
	}

	denom2, err := ftKeeper.Issue(ctx, settings2)
	requireT.NoError(err)

	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom2, 10)))
	_, isBroken = keeper.WhitelistingInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)
	// make the state valid (we use the slice here since the `sdk.NewCoins` sanitizes the empty coins)
	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.Coins{sdk.NewInt64Coin(denom2, 0)})
	_, isBroken = keeper.WhitelistingInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)
}

func TestBankMetadataExistInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdkmath.NewInt(1000),
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_whitelisting,
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	// check that initial state is valid
	_, isBroken := keeper.BankMetadataExistInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break the definition
	definition, err := ftKeeper.GetDefinition(ctx, denom)
	requireT.NoError(err)

	definition.Denom = "invalid"
	ftKeeper.SetDefinition(ctx, settings1.Issuer, settings1.Subunit, definition)

	// check that state is broken now
	_, isBroken = keeper.BankMetadataExistInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)
}
