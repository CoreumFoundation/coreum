package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/keeper"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v3/x/nft"
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
	err = nftKeeper.SetClassDefinition(ctx, types.ClassDefinition{
		ID:       types.BuildClassID("sampleid1", issuer),
		Issuer:   sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()).String(),
		Features: []types.ClassFeature{},
	})
	requireT.NoError(err)

	// invariant is broken
	_, isBroken = keeper.OriginalClassExistsInvariant(nftKeeper)(ctx)
	requireT.True(isBroken)
}

func TestFrozenNFTExistsInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue a class
	settings1 := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
		Features:    []types.ClassFeature{types.ClassFeature_freezing},
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, settings1)
	requireT.NoError(err)

	mintSettings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "nft-id",
	}
	err = assetNFTKeeper.Mint(ctx, mintSettings)
	requireT.NoError(err)

	err = assetNFTKeeper.Freeze(ctx, issuer, classID, mintSettings.ID)
	requireT.NoError(err)

	// invariant is valid
	_, isBroken := keeper.FreezingInvariant(assetNFTKeeper)(ctx)
	requireT.False(isBroken)

	// non-existing nft (invariant is broken)
	requireT.NoError(assetNFTKeeper.SetFrozen(ctx, classID, "next-nft", true))
	_, isBroken = keeper.FreezingInvariant(assetNFTKeeper)(ctx)
	requireT.True(isBroken)
}

func TestBurntNFTNotExistsInvariant(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})
	assetNFTKeeper := testApp.AssetNFTKeeper
	nftKeeper := testApp.NFTKeeper

	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// Issue a class
	settings1 := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
	}

	classID, err := assetNFTKeeper.IssueClass(ctx, settings1)
	requireT.NoError(err)

	mintSettings := types.MintSettings{
		Sender:  issuer,
		ClassID: classID,
		ID:      "nft-id",
	}
	err = assetNFTKeeper.Mint(ctx, mintSettings)
	requireT.NoError(err)

	err = assetNFTKeeper.Burn(ctx, mintSettings.Sender, mintSettings.ClassID, mintSettings.ID)
	requireT.NoError(err)

	// invariant is valid
	_, isBroken := keeper.BurntNFTInvariant(assetNFTKeeper)(ctx)
	requireT.False(isBroken)

	// burnt nft exists in nft module (invariant is broken)
	requireT.NoError(assetNFTKeeper.SetBurnt(ctx, classID, "next-nft"))
	requireT.NoError(nftKeeper.Mint(ctx, nft.NFT{
		ClassId: classID,
		Id:      "next-nft",
	}, mintSettings.Sender))

	_, isBroken = keeper.BurntNFTInvariant(assetNFTKeeper)(ctx)
	requireT.True(isBroken)
}
