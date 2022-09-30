package cli_test

import (
	"encoding/json"
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/feemodel/client/cli"
)

func TestMinGasPrice(t *testing.T) {
	networkCfg, err := config.NetworkByChainID(config.Devnet)
	require.NoError(t, err)
	app.ChosenNetwork = networkCfg

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	cmd := cli.GetQueryCmd()
	buf, err := clitestutil.ExecTestCLICmd(ctx, cmd, []string{"min-gas-price", "--output", "json"})
	require.NoError(t, err)

	var resp sdk.DecCoin
	require.NoError(t, json.Unmarshal(buf.Bytes(), &resp))

	assert.Equal(t, "stake", resp.Denom)
	assert.True(t, resp.Amount.GT(sdk.ZeroDec()))
}
