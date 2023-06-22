package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var params = Params{
	IssueFee:                    sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000),
	TokenUpgradeGracePeriod:     time.Second,
	TokenUpgradeDecisionTimeout: time.Date(2023, 3, 2, 1, 11, 12, 13, time.UTC),
}

func TestParamsValidation(t *testing.T) {
	assert.NoError(t, params.ValidateBasic())

	testParams := params
	testParams.IssueFee = sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)
	assert.NoError(t, params.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Denom: sdk.DefaultBondDenom}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Amount: sdk.OneInt()}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.IssueFee = sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(-10_000_000)}
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.TokenUpgradeGracePeriod = 0
	assert.Error(t, testParams.ValidateBasic())

	testParams = params
	testParams.TokenUpgradeGracePeriod = -1
	assert.Error(t, testParams.ValidateBasic())
}
