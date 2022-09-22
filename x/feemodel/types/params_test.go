package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var params = Params{
	Model: ModelParams{
		InitialGasPrice:         sdk.NewDec(1500),
		MaxGasPrice:             sdk.NewDec(1500000),
		MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
		EscalationStartBlockGas: 700,
		MaxBlockGas:             1000,
		ShortEmaBlockLength:     10,
		LongEmaBlockLength:      1000,
	},
}

func TestParamsValidation(t *testing.T) {
	assert.NoError(t, params.ValidateBasic())

	testParams := params
	testParams.Model.InitialGasPrice = sdk.NewDec(0)
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxGasPrice = testParams.Model.InitialGasPrice
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdk.ZeroDec()
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdk.OneDec()
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.Model.EscalationStartBlockGas = 0
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxBlockGas = testParams.Model.EscalationStartBlockGas
	assert.Error(t, testParams.ValidateBasic())
}
