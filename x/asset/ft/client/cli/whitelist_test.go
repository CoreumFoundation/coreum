package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

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
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Issue token
	args := []string{symbol, subunit, precision, "777", `"My Token"`,
		"--features", types.FungibleTokenFeature_whitelist.String(), //nolint:nosnakecase
	}
	args = append(args, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
	requireT.NoError(err)

	// test pagination
	for i := 0; i < 2; i++ {
		symbol := "l" + uuid.NewString()[:4]
		subunit := "sub" + symbol
		args := []string{symbol, subunit, precision, "777", `"My Token"`,
			"--features", types.FungibleTokenFeature_whitelist.String(), //nolint:nosnakecase
		}
		args = append(args, txValidator1Args(testNetwork)...)
		_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssueFungibleToken(), args)
		requireT.NoError(err)

		denom := types.BuildFungibleTokenDenom(subunit, issuer)
		token := "75" + denom
		args = append([]string{recipient.String(), token, "--output", "json"}, txValidator1Args(testNetwork)...)
		_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxSetWhitelistedLimitFungibleToken(), args)
		requireT.NoError(err)
	}

	var balancesResp types.QueryWhitelistedBalancesResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalances(), []string{recipient.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 2)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFungibleTokenWhitelistedBalances(), []string{recipient.String(), "--output", "json", "--limit", "1"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 1)
}
