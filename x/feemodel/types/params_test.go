package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

var params = Params{
	Model: ModelParams{
		InitialGasPrice:         sdkmath.LegacyNewDec(1500),
		MaxGasPriceMultiplier:   sdkmath.LegacyNewDec(1000),
		MaxDiscount:             sdkmath.LegacyMustNewDecFromStr("0.5"),
		EscalationStartFraction: sdkmath.LegacyMustNewDecFromStr("0.8"),
		MaxBlockGas:             1000,
		ShortEmaBlockLength:     10,
		LongEmaBlockLength:      1000,
	},
}

func TestParamsValidation(t *testing.T) {
	requireT := require.New(t)

	requireT.NoError(params.ValidateBasic())

	testParams := params
	testParams.Model.InitialGasPrice = sdkmath.LegacyNewDec(0)
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxGasPriceMultiplier = sdkmath.LegacyZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxGasPriceMultiplier = sdkmath.LegacyOneDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdkmath.LegacyZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdkmath.LegacyOneDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.MaxDiscount = sdkmath.LegacyZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.EscalationStartFraction = sdkmath.LegacyZeroDec()
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.Model.EscalationStartFraction = sdkmath.LegacyOneDec()
	requireT.Error(testParams.ValidateBasic())
}
