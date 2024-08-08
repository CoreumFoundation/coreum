package v1_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v4/x/feemodel/migrations/v1"
	"github.com/CoreumFoundation/coreum/v4/x/feemodel/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.NewContextLegacy(false, tmproto.Header{})

	testParams := types.Params{
		Model: types.ModelParams{
			InitialGasPrice:         sdkmath.LegacyNewDec(15),
			MaxGasPriceMultiplier:   sdkmath.LegacyNewDec(1000),
			MaxDiscount:             sdkmath.LegacyMustNewDecFromStr("0.1"),
			EscalationStartFraction: sdkmath.LegacyMustNewDecFromStr("0.8"),
			MaxBlockGas:             10,
			ShortEmaBlockLength:     1,
			LongEmaBlockLength:      3,
		},
	}
	keeper := testApp.FeeModelKeeper
	paramsKeeper := testApp.ParamsKeeper
	sp, ok := paramsKeeper.GetSubspace(types.ModuleName)
	requireT.True(ok)
	if !sp.HasKeyTable() {
		sp.WithKeyTable(paramstypes.NewKeyTable().RegisterParamSet(&types.Params{}))
	}
	sp.SetParamSet(ctx, &testParams)

	requireT.NoError(v1.MigrateParams(ctx, keeper, paramsKeeper))
	params := keeper.GetParams(ctx)
	assertT.EqualValues(params, testParams)
}
