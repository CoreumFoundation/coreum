package v1_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v3/x/customparams/migrations/v1"
	"github.com/CoreumFoundation/coreum/v3/x/customparams/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})

	testParams := types.StakingParams{
		MinSelfDelegation: sdk.NewInt(1245),
	}
	keeper := testApp.CustomParamsKeeper
	paramsKeeper := testApp.ParamsKeeper
	sp, ok := paramsKeeper.GetSubspace(types.CustomParamsStaking)
	requireT.True(ok)
	// set KeyTable if it has not already been set
	if !sp.HasKeyTable() {
		sp = sp.WithKeyTable(types.StakingParamKeyTable())
	}
	sp.SetParamSet(ctx, &testParams)

	requireT.NoError(v1.MigrateParams(ctx, keeper, paramsKeeper))
	params := keeper.GetStakingParams(ctx)
	assertT.EqualValues(params, testParams)
}
