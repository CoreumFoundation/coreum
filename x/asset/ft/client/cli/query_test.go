package cli_test

import (
	"testing"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestQueryTokens(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	issuer := testNetwork.Validators[0].Address

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_whitelisting,
		},
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}

	ctx := testNetwork.Validators[0].ClientCtx

	initialAmount := sdk.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryTokens(), []string{issuer.String(), "--output", "json", "--limit", "1"})
	requireT.NoError(err)

	var resp types.QueryTokensResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = resp.Tokens[0].Version // test should work with all versions
	requireT.Equal(expectedToken, resp.Tokens[0])
}

func TestQueryToken(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_whitelisting,
		},
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}
	ctx := testNetwork.Validators[0].ClientCtx

	initialAmount := sdk.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = resp.Token.Version // test should work with all versions
	requireT.Equal(expectedToken, resp.Token)

	// query balance
	var respBalance types.QueryBalanceResponse
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryBalance(), []string{expectedToken.Issuer, denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &respBalance))
	requireT.Equal(initialAmount.String(), respBalance.Balance.String())
}

func TestQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryParams(), []string{"--output", "json"})
	requireT.NoError(err)

	var resp types.QueryParamsResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	expectedIssueFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdk.NewInt(10_000_000)}
	requireT.Equal(expectedIssueFee, resp.Params.IssueFee)
}
