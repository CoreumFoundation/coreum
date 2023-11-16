package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var params = Params{
	Model: ModelParams{
		InitialGasPrice:         sdk.NewDec(1500),
		MaxGasPriceMultiplier:   sdk.NewDec(1000),
		MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
		EscalationStartFraction: sdk.MustNewDecFromStr("0.8"),
		MaxBlockGas:             1000,
		ShortEmaBlockLength:     10,
		LongEmaBlockLength:      1000,
	},
}

func TestParamsValidation(t *testing.T) {
	requireT := require.New(t)

	requireT.NoError(params.ValidateBasic())

	testParams := params
	testParams.Model.InitialGasPrice = sdk.NewDec(0)
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxGasPriceMultiplier = sdk.ZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxGasPriceMultiplier = sdk.OneDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdk.ZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdk.OneDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdk.ZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.EscalationStartFraction = sdk.ZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.EscalationStartFraction = sdk.OneDec()
	requireT.Error(testParams.ValidateBasic())
}
