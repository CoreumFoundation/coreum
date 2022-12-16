package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestMintBurn(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "abc"
	subunit := "subunit"
	precision := "8"
	ctx := testNetwork.Validators[0].ClientCtx
	issuer := testNetwork.Validators[0].Address
	denom := types.BuildDenom(subunit, issuer)

	// Issue token
	args := []string{symbol, subunit, precision, "777", `"My Token"`,
		"--features", types.TokenFeature_mint.String(), //nolint:nosnakecase
		"--features", types.TokenFeature_burn.String(), //nolint:nosnakecase
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssue(), args)
	requireT.NoError(err)

	// mint new tokens
	token := "100" + denom
	args = append([]string{token, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args)
	requireT.NoError(err)

	var balanceRsp banktypes.QueryAllBalancesResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balanceRsp))
	requireT.Equal("877", balanceRsp.Balances.AmountOf(denom).String())

	var supplyRsp sdk.Coin
	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetCmdQueryTotalSupply(), []string{issuer.String(), "--denom", denom, "--output", "json"})
	requireT.NoError(err)
	bs := buf.Bytes()
	requireT.NoError(ctx.Codec.UnmarshalJSON(bs, &supplyRsp))
	requireT.Equal("877"+denom, supplyRsp.String())

	// burn tokens
	token = "200" + denom
	args = append([]string{token, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balanceRsp))
	requireT.Equal("677", balanceRsp.Balances.AmountOf(denom).String())

	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetCmdQueryTotalSupply(), []string{issuer.String(), "--denom", denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &supplyRsp))
	requireT.Equal("677"+denom, supplyRsp.String())
}
