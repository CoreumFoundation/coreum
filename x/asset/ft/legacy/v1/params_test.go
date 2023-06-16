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
	paramsKeeper := testApp.ParamsKeeper

	requireT.NoError(v1.MigrateParams(ctx, paramsKeeper))

	params := keeper.GetParams(ctx)

	assertT.Equal("0stake", params.IssueFee.String())
	assertT.Equal(blockTime.Add(v1.InitialTokenUpgradeDecisionPeriod), params.TokenUpgradeDecisionTimeout)
	assertT.Equal(types.DefaultTokenUpgradeGracePeriod, params.TokenUpgradeGracePeriod)
}
