package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestWhitelistFungibleToken(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	// the denom must start from the letter
	symbol := "l" + uuid.NewString()[:4]
	subunit := "sub" + symbol
	precision := "6"
	ctx := testNetwork.Validators[0].ClientCtx
	issuer := testNetwork.Validators[0].Address
	denom := types.BuildFungibleTokenDenom(subunit, issuer)

	// Issue token
	args := []string{symbol, subunit, precision, issuer.String(), "777", `"My Token"`,
		"--features", types.FungibleTokenFeature_whitelist.String(), //nolint:nosnakecase
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	// verify that initial balance is whitelisted
	var resp types.QueryWhitelistedBalanceResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalance(), []string{issuer.String(), denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.Equal("777"+denom, resp.Balance.String())

	// test pagination
	for i := 0; i < 2; i++ {
		symbol := "l" + uuid.NewString()[:4]
		subunit := "sub" + symbol
		args := []string{symbol, subunit, precision, issuer.String(), "777", `"My Token"`,
			"--features", types.FungibleTokenFeature_whitelist.String(), //nolint:nosnakecase
		}
		args = append(args, txValidator1Args(testNetwork)...)
		_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
		requireT.NoError(err)
	}

	var balancesResp types.QueryWhitelistedBalancesResponse
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalances(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 3)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalances(), []string{issuer.String(), "--output", "json", "--limit", "1"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 1)

	// Change whitelisted limit
	token := "75" + denom
	args = append([]string{issuer.String(), token, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxSetWhitelistedLimitFungibleToken(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalance(), []string{issuer.String(), denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(token, resp.Balance.String())
}
