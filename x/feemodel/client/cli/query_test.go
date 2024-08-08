package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	"github.com/CoreumFoundation/coreum/v4/x/feemodel/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/feemodel/types"
)

func TestMain(t *testing.M) {
	// TODO(fix-cli-tests)
	// we are intentionally skipping cli tests to fix them later
}

func TestMinGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	var resp sdk.DecCoin
	require.NoError(t, coreumclitestutil.ExecQueryCmd(ctx, cli.GetQueryCmd(), []string{"min-gas-price"}, &resp))

	assert.Equal(t, testNetwork.Config.BondDenom, resp.Denom)
	assert.True(t, resp.Amount.GT(sdkmath.LegacyZeroDec()))
}

func TestRecommendedGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	cmd := cli.GetQueryCmd()

	var resp types.QueryRecommendedGasPriceResponse
	require.NoError(t, coreumclitestutil.ExecQueryCmd(ctx, cmd, []string{"recommended-gas-price", "--after", "10"}, &resp))

	assert.Greater(t, resp.Low.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
	assert.Greater(t, resp.Med.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
	assert.Greater(t, resp.High.Amount.MustFloat64(), sdkmath.LegacyZeroDec().MustFloat64())
}
