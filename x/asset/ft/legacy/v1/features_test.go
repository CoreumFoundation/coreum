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
	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
)

func TestMigrateFeatures(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := testApp.NewContext(false, tmproto.Header{}).WithBlockTime(blockTime)

	keeper := testApp.AssetFTKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEF",
		Subunit:       "def",
		Precision:     1,
		Description:   "DEF Desc",
		InitialAmount: sdk.NewInt(1000),
	}
	denom, err := keeper.Issue(ctx, settings)
	requireT.NoError(err)
	def, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)

	def.Features = []types.Feature{
		types.Feature_minting,
		types.Feature_minting, // should be removed
		types.Feature_whitelisting,
		3000,                       // should be removed
		types.Feature_whitelisting, // should be removed
		types.Feature_freezing,
		types.Feature_freezing, // should be removed
		types.Feature_burning,
		1000,                  // should be removed
		types.Feature_burning, // should be removed
		types.Feature_ibc,     // should be removed
		types.Feature_ibc,     // should be removed
		3000,                  // should be removed
	}
	keeper.SetDefinition(ctx, issuer, settings.Subunit, def)

	requireT.NoError(v1.MigrateFeatures(ctx, keeper))

	defChanged, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	assertT.Equal([]types.Feature{
		types.Feature_minting,
		types.Feature_whitelisting,
		types.Feature_freezing,
		types.Feature_burning,
	}, defChanged.Features)
}
