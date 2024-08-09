package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var params = Params{
	MintFee: sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000),
}

func TestParamsValidation(t *testing.T) {
	requireT := require.New(t)

	requireT.NoError(params.ValidateBasic())

	testParams := params
	testParams.MintFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	requireT.NoError(params.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Denom: sdk.DefaultBondDenom}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Amount: sdkmath.OneInt()}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.MintFee = sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdkmath.NewInt(-10_000_000)}
	requireT.Error(testParams.ValidateBasic())
}
