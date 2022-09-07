package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var params = Params{
	InitialGasPrice:         sdk.NewInt(1500),
	MaxGasPrice:             sdk.NewInt(1500000),
	MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
	EscalationStartBlockGas: 700,
	MaxBlockGas:             1000,
	ShortEmaBlockLength:     10,
	LongEmaBlockLength:      1000,
}

func TestParamsValidation(t *testing.T) {
	assert.NoError(t, params.Validate())

	testParams := params
	testParams.InitialGasPrice = sdk.NewInt(0)
	assert.Error(t, testParams.Validate())

	testParams = params
	testParams.MaxGasPrice = testParams.InitialGasPrice
	assert.Error(t, testParams.Validate())

	testParams = params
	testParams.MaxDiscount = sdk.ZeroDec()
	assert.Error(t, testParams.Validate())

	testParams = params
	testParams.MaxDiscount = sdk.OneDec()
	assert.Error(t, testParams.Validate())

	testParams = params
	testParams.EscalationStartBlockGas = 0
	assert.Error(t, testParams.Validate())

	testParams = params
	testParams.MaxBlockGas = testParams.EscalationStartBlockGas
	assert.Error(t, testParams.Validate())
}
