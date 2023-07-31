package cli_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutilcli "github.com/CoreumFoundation/coreum/testutil/cli"
	"github.com/CoreumFoundation/coreum/v2/testutil/network"
	"github.com/CoreumFoundation/coreum/v2/x/feemodel/client/cli"
	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

func TestMinGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	var resp sdk.DecCoin
	require.NoError(t, testutilcli.ExecQueryCmd(ctx, cli.GetQueryCmd(), []string{"min-gas-price"}, &resp))

	assert.Equal(t, testNetwork.Config.BondDenom, resp.Denom)
	assert.True(t, resp.Amount.GT(sdk.ZeroDec()))
}

func TestRecommendedGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	cmd := cli.GetQueryCmd()
	buf, err := clitestutil.ExecTestCLICmd(ctx, cmd, []string{"recommended-gas-price", "--after", "10", "--output", "json"})
	require.NoError(t, err)

	var resp types.QueryRecommendedGasPriceResponse
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	assert.Greater(t, resp.Low.Amount.MustFloat64(), sdk.ZeroDec().MustFloat64())
	assert.Greater(t, resp.Med.Amount.MustFloat64(), sdk.ZeroDec().MustFloat64())
	assert.Greater(t, resp.High.Amount.MustFloat64(), sdk.ZeroDec().MustFloat64())
}
