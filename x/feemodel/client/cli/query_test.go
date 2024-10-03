package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v5/testutil/cli"
	"github.com/CoreumFoundation/coreum/v5/testutil/network"
	"github.com/CoreumFoundation/coreum/v5/x/feemodel/client/cli"
	"github.com/CoreumFoundation/coreum/v5/x/feemodel/types"
)

func TestMinGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	var resp sdk.DecCoin
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.GetQueryCmd(), []string{"min-gas-price"}, &resp)

	assert.Equal(t, testNetwork.Config.BondDenom, resp.Denom)
	assert.True(t, resp.Amount.GT(sdkmath.LegacyZeroDec()))
}

func TestRecommendedGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	cmd := cli.GetQueryCmd()

	var resp types.QueryRecommendedGasPriceResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cmd, []string{"recommended-gas-price", "--after", "10"}, &resp)

	assert.Greater(t, resp.Low.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
	assert.Greater(t, resp.Med.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
	assert.Greater(t, resp.High.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
}
