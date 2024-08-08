package types

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var params = Params{
	IssueFee:                    sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000),
	TokenUpgradeGracePeriod:     time.Second,
	TokenUpgradeDecisionTimeout: time.Date(2023, 3, 2, 1, 11, 12, 13, time.UTC),
}

func TestParamsValidation(t *testing.T) {
	requireT := require.New(t)

	requireT.NoError(params.ValidateBasic())

	testParams := params
	testParams.IssueFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	requireT.NoError(params.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Denom: sdk.DefaultBondDenom}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Amount: sdkmath.OneInt()}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdkmath.NewInt(-10_000_000)}
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.TokenUpgradeGracePeriod = 0
	requireT.Error(testParams.ValidateBasic())

	testParams = params
	testParams.TokenUpgradeGracePeriod = -1
	requireT.Error(testParams.ValidateBasic())
}
