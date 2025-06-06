package feemodel_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/x/feemodel"
	"github.com/CoreumFoundation/coreum/v6/x/feemodel/types"
)

func newKeeperMock(genesisState types.GenesisState) *keeperMock {
	return &keeperMock{
		state: genesisState,
	}
}

type keeperMock struct {
	state types.GenesisState
}

func (k *keeperMock) TrackedGas(ctx sdk.Context) int64 {
	return 1
}

func (k *keeperMock) SetParams(ctx sdk.Context, params types.Params) error {
	k.state.Params = params
	return nil
}

func (k *keeperMock) GetParams(ctx sdk.Context) (types.Params, error) {
	return k.state.Params, nil
}

func (k *keeperMock) GetShortEMAGas(ctx sdk.Context) int64 {
	return 0
}

func (k *keeperMock) SetShortEMAGas(ctx sdk.Context, emaGas int64) error {
	return nil
}

func (k *keeperMock) GetLongEMAGas(ctx sdk.Context) int64 {
	return 0
}

func (k *keeperMock) SetLongEMAGas(ctx sdk.Context, emaGas int64) error {
	return nil
}

func (k *keeperMock) GetMinGasPrice(ctx sdk.Context) sdk.DecCoin {
	return k.state.MinGasPrice
}

func (k *keeperMock) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.DecCoin) error {
	k.state.MinGasPrice = minGasPrice
	return nil
}

func (k *keeperMock) CalculateEdgeGasPriceAfterBlocks(ctx sdk.Context, after uint32) (sdk.DecCoin, sdk.DecCoin, error) {
	return sdk.NewDecCoin("", sdkmath.ZeroInt()), sdk.NewDecCoin("", sdkmath.ZeroInt()), nil
}

func (k *keeperMock) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	return nil
}

func setup() (feemodel.AppModule, feemodel.Keeper, types.GenesisState, codec.Codec) {
	genesisState := types.GenesisState{
		Params: types.Params{
			Model: types.ModelParams{
				InitialGasPrice:         sdkmath.LegacyNewDec(15),
				MaxGasPriceMultiplier:   sdkmath.LegacyNewDec(1000),
				MaxDiscount:             sdkmath.LegacyMustNewDecFromStr("0.1"),
				EscalationStartFraction: sdkmath.LegacyMustNewDecFromStr("0.8"),
				MaxBlockGas:             10,
				ShortEmaBlockLength:     1,
				LongEmaBlockLength:      3,
			},
		},
		MinGasPrice: sdk.NewDecCoin("coin", sdkmath.NewInt(155)),
	}
	cdc := config.NewEncodingConfig(feemodel.AppModuleBasic{}).Codec
	keeper := newKeeperMock(genesisState)
	module := feemodel.NewAppModule(keeper, nil)

	return module, keeper, genesisState, cdc
}

func TestInitGenesis(t *testing.T) {
	module, keeper, state, cdc := setup()

	genesisState := state
	genesisState.Params.Model.InitialGasPrice.Add(sdkmath.LegacyOneDec())
	genesisState.Params.Model.MaxGasPriceMultiplier.Add(sdkmath.LegacyOneDec())
	genesisState.Params.Model.MaxDiscount.Add(sdkmath.LegacyMustNewDecFromStr("0.2"))
	genesisState.Params.Model.EscalationStartFraction.Sub(sdkmath.LegacyMustNewDecFromStr("0.1"))
	genesisState.Params.Model.MaxBlockGas++
	genesisState.Params.Model.ShortEmaBlockLength++
	genesisState.Params.Model.LongEmaBlockLength++
	genesisState.MinGasPrice.Denom = "coin2"
	genesisState.MinGasPrice.Amount.Add(sdkmath.LegacyOneDec())

	module.InitGenesis(sdk.Context{}, cdc, cdc.MustMarshalJSON(&genesisState))

	params, err := keeper.GetParams(sdk.Context{})
	require.NoError(t, err)
	minGasPrice := keeper.GetMinGasPrice(sdk.Context{})
	assert.Equal(t, genesisState.Params.Model.InitialGasPrice.String(), params.Model.InitialGasPrice.String())
	assert.Equal(t, genesisState.Params.Model.MaxGasPriceMultiplier.String(), params.Model.MaxGasPriceMultiplier.String())
	assert.Equal(t, genesisState.Params.Model.MaxDiscount.String(), params.Model.MaxDiscount.String())
	assert.Equal(
		t,
		genesisState.Params.Model.EscalationStartFraction.String(),
		params.Model.EscalationStartFraction.String(),
	)
	assert.Equal(t, genesisState.Params.Model.MaxBlockGas, params.Model.MaxBlockGas)
	assert.Equal(t, genesisState.Params.Model.ShortEmaBlockLength, params.Model.ShortEmaBlockLength)
	assert.Equal(t, genesisState.Params.Model.LongEmaBlockLength, params.Model.LongEmaBlockLength)
	assert.Equal(t, genesisState.MinGasPrice.Denom, minGasPrice.Denom)
	assert.True(t, genesisState.MinGasPrice.Amount.Equal(minGasPrice.Amount))
}

func TestExport(t *testing.T) {
	module, _, state, cdc := setup()

	var decodedGenesis types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(module.ExportGenesis(sdk.Context{}, cdc), &decodedGenesis))

	assert.Equal(t, state, decodedGenesis)
}

func TestEndBlock(t *testing.T) {
	module, keeper, state, _ := setup()

	require.NoError(t, module.EndBlock(sdk.Context{}))

	model := types.NewModel(state.Params.Model)
	minGasPrice := keeper.GetMinGasPrice(sdk.Context{})
	assert.True(t, minGasPrice.Amount.Equal(model.CalculateGasPriceWithMaxDiscount()))
	assert.Equal(t, minGasPrice.Denom, state.MinGasPrice.Denom)
}
