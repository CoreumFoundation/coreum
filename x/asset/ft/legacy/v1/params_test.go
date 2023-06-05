package v1_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := testApp.NewContext(false, tmproto.Header{}).WithBlockTime(blockTime)

	keeper := testApp.AssetFTKeeper
	paramsOld := keeper.GetParams(ctx)
	paramsOld.TokenUpgradeDecisionTimeout = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	paramsOld.TokenUpgradeGracePeriod = time.Second
	keeper.SetParams(ctx, paramsOld)
	paramsOld2 := keeper.GetParams(ctx)
	requireT.Equal(paramsOld, paramsOld2)

	requireT.NoError(v1.MigrateParams(ctx, keeper))

	paramsNew := keeper.GetParams(ctx)

	assertT.Equal(blockTime.Add(types.DefaultTokenUpgradeDecisionPeriod), paramsNew.TokenUpgradeDecisionTimeout)
	assertT.Equal(types.DefaultTokenUpgradeGracePeriod, paramsNew.TokenUpgradeGracePeriod)
}
