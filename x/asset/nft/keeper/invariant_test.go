package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/nft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestOriginalClassExistsInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue a class
	settings1 := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
	}

	_, err := nftKeeper.IssueClass(ctx, settings1)
	requireT.NoError(err)

	// invariant is valid
	_, isBroken := keeper.OriginalClassExistsInvariant(nftKeeper)(ctx)
	requireT.False(isBroken)

	// set class definition directly (break consistency)
	nftKeeper.SetClassDefinition(ctx, types.ClassDefinition{
		ID:       "sample-id1",
		Issuer:   sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
		Features: []types.ClassFeature{},
	})

	// invariant is broken
	_, isBroken = keeper.OriginalClassExistsInvariant(nftKeeper)(ctx)
	requireT.True(isBroken)
}

func TestWhitelistedNFTExistsInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	recipient := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue a class
	settings1 := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
		Features:    []types.ClassFeature{types.ClassFeature_whitelisting},
	}

	classID, err := nftKeeper.IssueClass(ctx, settings1)
	requireT.NoError(err)

	mintSettings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "nft-id",
	}
	err = nftKeeper.Mint(ctx, mintSettings)
	requireT.NoError(err)

	err = nftKeeper.AddToWhitelist(ctx, classID, mintSettings.ID, issuer, recipient)
	requireT.NoError(err)

	// invariant is valid
	_, isBroken := keeper.WhitelistingInvariant(nftKeeper)(ctx)
	requireT.False(isBroken)

	// store nft directly (break consistency)
	requireT.NoError(nftKeeper.SetWhitelisting(ctx, classID, "next-nft", recipient, true))

	// invariant is broken
	_, isBroken = keeper.WhitelistingInvariant(nftKeeper)(ctx)
	requireT.True(isBroken)
}

func TestFrozenNFTExistsInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue a class
	settings1 := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
		Features:    []types.ClassFeature{types.ClassFeature_freezing},
	}

	classID, err := nftKeeper.IssueClass(ctx, settings1)
	requireT.NoError(err)

	mintSettings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "nft-id",
	}
	err = nftKeeper.Mint(ctx, mintSettings)
	requireT.NoError(err)

	err = nftKeeper.Freeze(ctx, issuer, classID, mintSettings.ID)
	requireT.NoError(err)

	// invariant is valid
	_, isBroken := keeper.FreezingInvariant(nftKeeper)(ctx)
	requireT.False(isBroken)

	// store nft directly (break consistency)
	requireT.NoError(nftKeeper.SetFrozen(ctx, classID, "next-nft", true))

	// invariant is broken
	_, isBroken = keeper.FreezingInvariant(nftKeeper)(ctx)
	requireT.True(isBroken)
}
