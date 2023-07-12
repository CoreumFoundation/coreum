package keeper_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/x/feemodel/keeper"
	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

func newParamSubspaceMock() *paramSubspaceMock {
	return &paramSubspaceMock{
		params: map[string][]byte{},
	}
}

type paramSubspaceMock struct {
	params map[string][]byte
}

func (psm *paramSubspaceMock) GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet) {
	for _, pair := range ps.ParamSetPairs() {
		must.OK(json.Unmarshal(psm.params[string(pair.Key)], pair.Value))
	}
}

func (psm *paramSubspaceMock) SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet) {
	for _, pair := range ps.ParamSetPairs() {
		psm.params[string(pair.Key)] = must.Bytes(json.Marshal(pair.Value))
	}
}

func setup() (sdk.Context, keeper.Keeper) {
	key := sdk.NewKVStoreKey(types.StoreKey)
	tKey := sdk.NewTransientStoreKey(types.TransientStoreKey)

	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tKey, sdk.StoreTypeTransient, db)
	must.OK(cms.LoadLatestVersion())
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, log.NewNopLogger())

	return ctx, keeper.NewKeeper(newParamSubspaceMock(), key, tKey)
}

func TestTrackGas(t *testing.T) {
	ctx, keeper := setup()

	assert.EqualValues(t, 0, keeper.TrackedGas(ctx))

	keeper.TrackGas(ctx, 10)
	assert.EqualValues(t, 10, keeper.TrackedGas(ctx))

	keeper.TrackGas(ctx, 5)
	assert.EqualValues(t, 15, keeper.TrackedGas(ctx))
}

func TestShortEMAGas(t *testing.T) {
	ctx, keeper := setup()

	assert.EqualValues(t, 0, keeper.GetShortEMAGas(ctx))

	keeper.SetShortEMAGas(ctx, 10)
	assert.EqualValues(t, 10, keeper.GetShortEMAGas(ctx))
}

func TestLongEMAGas(t *testing.T) {
	ctx, keeper := setup()

	assert.EqualValues(t, 0, keeper.GetLongEMAGas(ctx))

	keeper.SetLongEMAGas(ctx, 10)
	assert.EqualValues(t, 10, keeper.GetLongEMAGas(ctx))
}

func TestMinGasPrice(t *testing.T) {
	ctx, keeper := setup()

	keeper.SetMinGasPrice(ctx, sdk.NewDecCoin("coin", sdk.NewInt(10)))
	minGasPrice := keeper.GetMinGasPrice(ctx)
	assert.Equal(t, "10.000000000000000000", minGasPrice.Amount.String())
	assert.Equal(t, "coin", minGasPrice.Denom)

	keeper.SetMinGasPrice(ctx, sdk.NewDecCoin("coin", sdk.NewInt(20)))
	minGasPrice = keeper.GetMinGasPrice(ctx)
	assert.EqualValues(t, "20.000000000000000000", minGasPrice.Amount.String())
	assert.Equal(t, "coin", minGasPrice.Denom)
}

func TestParams(t *testing.T) {
	ctx, keeper := setup()

	defParams := types.DefaultParams()
	keeper.SetParams(ctx, defParams)
	params := keeper.GetParams(ctx)

	assert.Equal(t, defParams.Model.InitialGasPrice.String(), params.Model.InitialGasPrice.String())
	assert.Equal(t, defParams.Model.MaxGasPriceMultiplier.String(), params.Model.MaxGasPriceMultiplier.String())
	assert.Equal(t, defParams.Model.MaxDiscount.String(), params.Model.MaxDiscount.String())
	assert.Equal(t, defParams.Model.EscalationStartFraction.String(), params.Model.EscalationStartFraction.String())
	assert.Equal(t, defParams.Model.MaxBlockGas, params.Model.MaxBlockGas)
	assert.Equal(t, defParams.Model.ShortEmaBlockLength, params.Model.ShortEmaBlockLength)
	assert.Equal(t, defParams.Model.LongEmaBlockLength, params.Model.LongEmaBlockLength)
}

//nolint:funlen // splitting tests will make the code less readable
func TestEstimateGasPriceInFuture(t *testing.T) {
	ctx, keeper := setup()
	defParams := types.Params{
		Model: types.ModelParams{
			InitialGasPrice:         sdk.MustNewDecFromStr("0.0625"),
			MaxGasPriceMultiplier:   sdk.MustNewDecFromStr("1000.0"),
			MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
			EscalationStartFraction: sdk.MustNewDecFromStr("0.8"),
			MaxBlockGas:             50000000, // 400 * BankSend message
			ShortEmaBlockLength:     50,
			LongEmaBlockLength:      1000,
		},
	}
	keeper.SetParams(ctx, defParams)

	testCases := []struct {
		name        string
		shortEMA    int64
		longEMA     int64
		afterBlocks uint32
		assertions  func(t *testing.T, low, high sdk.DecCoin)
	}{
		{
			name:        "short and long ema are 0. (10 blocks after)",
			shortEMA:    0,
			longEMA:     0,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.03215
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount(), "low amount is max discount")
				assertT.EqualValues(high.Amount, model.CalculateGasPriceWithMaxDiscount(), "high amount is max discount")
			},
		},
		{
			name:        "short and long ema are 0. (50 blocks after)",
			shortEMA:    0,
			longEMA:     0,
			afterBlocks: 50,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.03215
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.EqualValues(high.Amount, model.CalculateGasPriceWithMaxDiscount())
			},
		},
		{
			name:        "short and long ema are equal. (1 tx 50 blocks after)",
			shortEMA:    100_000,
			longEMA:     100_000,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.032203830017345169
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.Greater(high.Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
				assertT.Less(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64())
			},
		},
		{
			name:        "short and long ema are equal. (1 tx in block 50 blocks after)",
			shortEMA:    100_000,
			longEMA:     100_000,
			afterBlocks: 50,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.043154704024898720
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.Greater(high.Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64(), "high amount is greater than max discount")
				assertT.Less(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64(), "high amount is less than initial gas price")
			},
		},
		{
			name:        "short and long ema are equal. (10 tx in block on average)",
			shortEMA:    1_000_000,
			longEMA:     1_000_000,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.032203835927155826
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.Greater(high.Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
				assertT.Less(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64())
			},
		},
		{
			name:        "short ema is smaller than long ema.",
			shortEMA:    4_000_000,
			longEMA:     5_000_000,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.034857583137282673
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.Greater(high.Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
				assertT.Less(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64())
			},
		},
		{
			name:        "short ema is between long ema and escalation region. (10 block after)",
			shortEMA:    5_000_000,
			longEMA:     1_000_000,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.03215
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount())
				assertT.EqualValues(high.Amount, model.CalculateGasPriceWithMaxDiscount())
			},
		},
		{
			name:        "short ema is between long ema and escalation region. (50 block after)",
			shortEMA:    5_000_000,
			longEMA:     1_000_000,
			afterBlocks: 50,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.03215
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount(), "low amount is equal to max discount")
				assertT.EqualValues(high.Amount, model.CalculateGasPriceWithMaxDiscount(), "high amount is equal to max discount")
			},
		},
		{
			name:        "short ema is right before escalation region. (10 block after)",
			shortEMA:    39_000_000,
			longEMA:     10_000_000,
			afterBlocks: 10,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 0.671267795027995000
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64(), "low amount is equal to max discount")
				assertT.Greater(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64()*10, "high amount is much higher than the initial price. (in escalation)")
			},
		},
		{
			name:        "short ema is right before escalation region. (50 block after)",
			shortEMA:    39_000_000,
			longEMA:     10_000_000,
			afterBlocks: 50,
			assertions: func(t *testing.T, low, high sdk.DecCoin) {
				// observed min: 0.03215
				// observed max: 22.475936159292449688
				assertT := assert.New(t)
				model := types.NewModel(defParams.Model)
				assertT.EqualValues(low.Amount, model.CalculateGasPriceWithMaxDiscount(), "low amount is equal to max discount")
				assertT.Greater(high.Amount.MustFloat64(), model.Params().InitialGasPrice.MustFloat64()*300, "high amount is much higher than the initial price. (in escalation)")
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			keeper.SetMinGasPrice(ctx, sdk.NewDecCoinFromDec("coin", sdk.MustNewDecFromStr("0.0625")))
			keeper.SetShortEMAGas(ctx, tc.shortEMA)
			keeper.SetLongEMAGas(ctx, tc.longEMA)
			low, high, err := keeper.CalculateEdgeGasPriceAfterBlocks(ctx, tc.afterBlocks)
			require.NoError(t, err)
			tc.assertions(t, low, high)
		})
	}
}
