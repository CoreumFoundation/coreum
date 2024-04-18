package v3_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	v3 "github.com/CoreumFoundation/coreum/v4/x/asset/ft/migrations/v3"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestMigrateDefinitions(t *testing.T) {
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
		InitialAmount: sdkmath.NewInt(1000),
	}
	denom, err := keeper.Issue(ctx, settings)
	requireT.NoError(err)
	def, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)

	def.Admin = ""
	keeper.SetDefinition(ctx, issuer, settings.Subunit, def)

	requireT.NoError(v3.MigrateDefinitions(ctx, keeper))

	defChanged, err := keeper.GetDefinition(ctx, denom)
	requireT.NoError(err)
	assertT.Equal(defChanged.Issuer, defChanged.Admin)
}
