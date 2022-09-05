package feemodel_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/x/feemodel"
	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

var genesis = types.GenesisState{
	Params: types.Params{
		InitialGasPrice:         sdk.NewInt(15),
		MaxGasPrice:             sdk.NewInt(150),
		MaxDiscount:             sdk.MustNewDecFromStr("0.1"),
		EscalationStartBlockGas: 7,
		MaxBlockGas:             10,
		ShortEmaBlockLength:     1,
		LongEmaBlockLength:      3,
	},
	MinGasPrice: sdk.NewCoin("coin", sdk.NewInt(155)),
}

type keeperMock struct {
}

func (k keeperMock) TrackedGas(ctx sdk.Context) int64 {
	panic("not implemented")
}

func (k keeperMock) SetParams(ctx sdk.Context, params types.Params) {
	panic("not implemented")
}

func (k keeperMock) GetParams(ctx sdk.Context) types.Params {
	return genesis.Params
}

func (k keeperMock) GetShortEMAGas(ctx sdk.Context) int64 {
	panic("not implemented")
}

func (k keeperMock) SetShortEMAGas(ctx sdk.Context, emaGas int64) {
	panic("not implemented")
}

func (k keeperMock) GetLongEMAGas(ctx sdk.Context) int64 {
	panic("not implemented")
}

func (k keeperMock) SetLongEMAGas(ctx sdk.Context, emaGas int64) {
	panic("not implemented")
}

func (k keeperMock) GetMinGasPrice(ctx sdk.Context) sdk.Coin {
	return genesis.MinGasPrice
}

func (k keeperMock) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin) {
	panic("not implemented")
}

func TestExport(t *testing.T) {
	cdc := app.NewEncodingConfig().Marshaler

	module := feemodel.NewAppModule(keeperMock{})
	encodedGenesis := module.ExportGenesis(sdk.Context{}, cdc)

	var decodedGenesis types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(encodedGenesis, &decodedGenesis))
	assert.EqualValues(t, genesis, decodedGenesis)
}
