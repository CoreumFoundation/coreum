package cli_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testutilcli "github.com/CoreumFoundation/coreum/testutil/cli"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/feemodel/client/cli"
)

func TestMinGasPrice(t *testing.T) {
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	var resp sdk.DecCoin
	require.NoError(t, testutilcli.ExecQueryCmd(ctx, cli.GetQueryCmd(), []string{"min-gas-price"}, &resp))

	assert.Equal(t, testNetwork.Config.BondDenom, resp.Denom)
	assert.True(t, resp.Amount.GT(sdk.ZeroDec()))
}
