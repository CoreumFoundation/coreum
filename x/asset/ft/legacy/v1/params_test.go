package v1_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestMigrateParams(t *testing.T) {
	assertT := assert.New(t)

	testApp := simapp.New()
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := testApp.NewContext(false, tmproto.Header{}).WithBlockTime(blockTime)

	keeper := testApp.AssetFTKeeper
	paramsOld := keeper.GetParamsV1(ctx)

	v1.MigrateParams(ctx, keeper)

	paramsNew := keeper.GetParams(ctx)

	assertT.Equal(paramsOld.IssueFee, paramsNew.IssueFee)
	assertT.Equal(blockTime.Add(v1.InitialTokenUpgradeDecisionPeriod), paramsNew.TokenUpgradeDecisionTimeout)
	assertT.Equal(types.DefaultTokenUpgradeGracePeriod, paramsNew.TokenUpgradeGracePeriod)
}
