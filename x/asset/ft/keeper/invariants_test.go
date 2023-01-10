package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestNonNegativeBalancesInvariant(t *testing.T) {
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
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(1000),
		Features: []types.TokenFeature{
			types.TokenFeature_freeze,    //nolint:nosnakecase
			types.TokenFeature_whitelist, //nolint:nosnakecase
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
	err = bankKeeper.SendCoins(ctx, issuer, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
	requireT.NoError(err)

	// check that current state is valid
	_, isBroken := keeper.NonNegativeBalancesInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break frozen the state and check
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.Coins{sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewInt(-1),
	}})
	_, isBroken = keeper.NonNegativeBalancesInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)

	// make the state valid
	ftKeeper.SetFrozenBalances(ctx, recipient, sdk.NewCoins(sdk.NewInt64Coin(denom, 10)))
	_, isBroken = keeper.NonNegativeBalancesInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break whitelisted the state and check
	ftKeeper.SetWhitelistedBalances(ctx, recipient, sdk.Coins{sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewInt(-1),
	}})
	_, isBroken = keeper.NonNegativeBalancesInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)
}

func TestBankMatchesInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	ftKeeper := testApp.AssetFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings1 := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(1000),
		Features: []types.TokenFeature{
			types.TokenFeature_freeze,    //nolint:nosnakecase
			types.TokenFeature_whitelist, //nolint:nosnakecase
		},
	}

	denom, err := ftKeeper.Issue(ctx, settings1)
	requireT.NoError(err)

	// check that initial state is valid
	_, isBroken := keeper.BankMetadataMatchesInvariant(ftKeeper)(ctx)
	requireT.False(isBroken)

	// break the definition
	definition, err := ftKeeper.GetTokenDefinition(ctx, denom)
	requireT.NoError(err)

	definition.Denom = "invalid"
	ftKeeper.SetTokenDefinition(ctx, settings1.Issuer, settings1.Subunit, definition)

	// check that state is broken now
	_, isBroken = keeper.BankMetadataMatchesInvariant(ftKeeper)(ctx)
	requireT.True(isBroken)
}
