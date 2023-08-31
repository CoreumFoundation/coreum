package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var params = Params{
	MintFee: sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000),
}

func TestParamsValidation(t *testing.T) {
	assert.NoError(t, params.ValidateBasic())

	testParams := params
	testParams.MintFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	assert.NoError(t, params.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Denom: sdk.DefaultBondDenom}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Amount: sdk.OneInt()}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdkmath.NewInt(-10_000_000)}
	assert.Error(t, testParams.ValidateBasic())
}
