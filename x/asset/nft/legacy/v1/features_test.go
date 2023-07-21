package v1_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/nft/legacy/v1"
	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
)

func TestMigrateFeatures(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := testApp.NewContext(false, tmproto.Header{}).WithBlockTime(blockTime)

	keeper := testApp.AssetNFTKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	settings := types.IssueClassSettings{
		Issuer:      issuer,
		Symbol:      "DEF",
		Description: "DEF Desc",
	}

	classID, err := keeper.IssueClass(ctx, settings)
	requireT.NoError(err)

	def, err := keeper.GetClassDefinition(ctx, classID)
	requireT.NoError(err)

	def.Features = []types.ClassFeature{
		types.ClassFeature_disable_sending,
		4000,                               // should be removed
		types.ClassFeature_disable_sending, // should be removed
		4000,                               // should be removed
		types.ClassFeature_whitelisting,
		types.ClassFeature_whitelisting, // should be removed
		types.ClassFeature_freezing,
		1000,                        // should be removed
		types.ClassFeature_freezing, // should be removed
		types.ClassFeature_burning,
		types.ClassFeature_burning, // should be removed
	}
	requireT.NoError(keeper.SetClassDefinition(ctx, def))

	requireT.NoError(v1.MigrateClassFeatures(ctx, keeper))

	defChanged, err := keeper.GetClassDefinition(ctx, classID)
	requireT.NoError(err)
	assertT.Equal([]types.ClassFeature{
		types.ClassFeature_disable_sending,
		types.ClassFeature_whitelisting,
		types.ClassFeature_freezing,
		types.ClassFeature_burning,
	}, defChanged.Features)
}
