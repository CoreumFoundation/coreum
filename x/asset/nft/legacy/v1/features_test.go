package v1_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/x/asset/nft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
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
		Features: []types.ClassFeature{
			types.ClassFeature_disable_sending,
			types.ClassFeature_whitelisting,
			types.ClassFeature_freezing,
			types.ClassFeature_burning,
		},
	}

	classID, err := keeper.IssueClass(ctx, settings)
	requireT.NoError(err)

	requireT.NoError(v1.MigrateClassFeatures(ctx, keeper, types.ClassFeature_name))

	defUnchanged, err := keeper.GetClassDefinition(ctx, classID)
	requireT.NoError(err)
	assertT.Equal(settings.Features, defUnchanged.Features)

	requireT.NoError(v1.MigrateClassFeatures(ctx, keeper, map[int32]string{
		1: "freezing",
		2: "whitelisting",
	}))

	defChanged, err := keeper.GetClassDefinition(ctx, classID)
	requireT.NoError(err)
	assertT.Equal([]types.ClassFeature{
		types.ClassFeature_whitelisting,
		types.ClassFeature_freezing,
	}, defChanged.Features)
}
