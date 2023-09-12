package v1_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/ft/legacy/v1"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := testApp.NewContext(false, tmproto.Header{}).WithBlockTime(blockTime)

	paramsKeeper := testApp.ParamsKeeper

	sp, ok := paramsKeeper.GetSubspace(types.ModuleName)
	requireT.True(ok)
	// set KeyTable if it has not already been set
	if !sp.HasKeyTable() {
		sp.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&types.Params{}))
	}
	sp.Set(ctx, types.KeyIssueFee, sdk.NewCoin("stake", sdk.ZeroInt()))

	requireT.NoError(v1.MigrateParams(ctx, paramsKeeper))

	var params types.Params
	sp.GetParamSet(ctx, &params)

	assertT.Equal("0stake", params.IssueFee.String())
	assertT.Equal(blockTime.Add(v1.InitialTokenUpgradeDecisionPeriod), params.TokenUpgradeDecisionTimeout)
	assertT.Equal(types.DefaultTokenUpgradeGracePeriod, params.TokenUpgradeGracePeriod)
}
