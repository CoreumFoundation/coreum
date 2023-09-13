package v1_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v3/x/feemodel/migrations/v1"
	"github.com/CoreumFoundation/coreum/v3/x/feemodel/types"
)

func TestMigrateParams(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)

	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})

	testParams := types.Params{
		Model: types.ModelParams{
			InitialGasPrice:         sdk.NewDec(15),
			MaxGasPriceMultiplier:   sdk.NewDec(1000),
			MaxDiscount:             sdk.MustNewDecFromStr("0.1"),
			EscalationStartFraction: sdk.MustNewDecFromStr("0.8"),
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
