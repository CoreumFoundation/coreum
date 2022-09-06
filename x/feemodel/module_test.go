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

func newKeeperMock(genesis types.GenesisState) keeperMock {
	return keeperMock{
		genesis: genesis,
	}
}

type keeperMock struct {
	genesis types.GenesisState
}

func (k keeperMock) TrackedGas(ctx sdk.Context) int64 {
	panic("not implemented")
}

func (k keeperMock) SetParams(ctx sdk.Context, params types.Params) {
	panic("not implemented")
}

func (k keeperMock) GetParams(ctx sdk.Context) types.Params {
	return k.genesis.Params
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
	return k.genesis.MinGasPrice
}

func (k keeperMock) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin) {
	panic("not implemented")
}

func TestExport(t *testing.T) {
	genesis := types.GenesisState{
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

	cdc := app.NewEncodingConfig().Marshaler

	module := feemodel.NewAppModule(newKeeperMock(genesis))
	encodedGenesis := module.ExportGenesis(sdk.Context{}, cdc)

	var decodedGenesis types.GenesisState
	require.NoError(t, cdc.UnmarshalJSON(encodedGenesis, &decodedGenesis))
	assert.EqualValues(t, genesis, decodedGenesis)
}
