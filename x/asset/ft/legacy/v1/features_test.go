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
	v1 "github.com/CoreumFoundation/coreum/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
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
		Features: []types.Feature{
			types.Feature_minting,
			types.Feature_whitelisting,
			types.Feature_freezing,
			types.Feature_burning,
			types.Feature_ibc,
		},
	}
	denom, err := keeper.Issue(ctx, settings)
	requireT.NoError(err)

	requireT.NoError(v1.MigrateFeatures(ctx, keeper, types.Feature_name))

	defChanged1, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	assertT.Equal([]types.Feature{
		types.Feature_minting,
		types.Feature_whitelisting,
		types.Feature_freezing,
		types.Feature_burning,
	}, defChanged1.Features)

	requireT.NoError(v1.MigrateFeatures(ctx, keeper, map[int32]string{
		1: "burning",
		3: "whitelisting",
	}))

	defChanged2, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	assertT.Equal([]types.Feature{
		types.Feature_whitelisting,
		types.Feature_burning,
	}, defChanged2.Features)
}
