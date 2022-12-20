package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestGloballyFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg
	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "l" + uuid.NewString()[:4]
	subunit := "sub" + symbol
	precision := "8"
	ctx := testNetwork.Validators[0].ClientCtx
	issuer := testNetwork.Validators[0].Address
	denom := types.BuildDenom(subunit, issuer)

	// Issue token
	args := []string{symbol, subunit, precision, "777", `"My Token"`,
		"--features", types.TokenFeature_freeze.String(), //nolint:nosnakecase
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssue(), args)
	requireT.NoError(err)

	// Globally freeze the token
	args = append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGloballyFreeze(), args)
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryTokenInfo(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.True(resp.Token.GloballyFrozen)

	// Globally unfreeze the token
	args = append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGloballyUnfreeze(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryTokenInfo(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.False(resp.Token.GloballyFrozen)
}
