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
