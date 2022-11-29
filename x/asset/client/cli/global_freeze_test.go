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
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestGlobalFreezeFungibleToken(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg
	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "l" + uuid.NewString()[:4]
	ctx := testNetwork.Validators[0].ClientCtx
	issuer := testNetwork.Validators[0].Address
	denom := types.BuildFungibleTokenDenom(symbol, issuer)

	// Issue token
	args := []string{symbol, testNetwork.Validators[0].Address.String(), "777", `"My Token"`,
		"--features", types.FungibleTokenFeature_freeze.String(), //nolint:nosnakecase
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	// Globally freeze the token
	args = append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGlobalFreezeFungibleToken(), args)
	requireT.NoError(err)

	var resp types.QueryFungibleTokenResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.True(resp.FungibleToken.GloballyFrozen)

	// Globally unfreeze the token
	args = append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGlobalUnfreezeFungibleToken(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.False(resp.FungibleToken.GloballyFrozen)
}
