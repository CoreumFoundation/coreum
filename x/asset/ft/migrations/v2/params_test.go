package v2_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	v2 "github.com/CoreumFoundation/coreum/v4/x/asset/ft/migrations/v2"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.NewContextLegacy(false, tmproto.Header{})

	testParams := types.Params{
		IssueFee:                    sdk.NewCoin("test-coin", sdkmath.NewInt(10)),
		TokenUpgradeDecisionTimeout: time.Now().UTC(),
		TokenUpgradeGracePeriod:     time.Second,
	}
	keeper := testApp.AssetFTKeeper
	paramsKeeper := testApp.ParamsKeeper
	sp, ok := paramsKeeper.GetSubspace(types.ModuleName)
	requireT.True(ok)
	// set KeyTable if it has not already been set
	if !sp.HasKeyTable() {
		sp.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&types.Params{}))
	}

	sp.SetParamSet(ctx, &testParams)

	requireT.NoError(v2.MigrateParams(ctx, keeper, paramsKeeper))
	params := keeper.GetParams(ctx)
	assertT.EqualValues(params, testParams)
}
