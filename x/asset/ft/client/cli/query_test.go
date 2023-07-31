package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testutilcli "github.com/CoreumFoundation/coreum/testutil/cli"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v2/testutil/network"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
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

	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	var resp types.QueryTokensResponse
	requireT.NoError(testutilcli.ExecQueryCmd(ctx, cli.CmdQueryTokens(), []string{issuer.String(), "--limit", "1"}, &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = types.CurrentTokenVersion
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

	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	var resp types.QueryTokenResponse
	requireT.NoError(testutilcli.ExecQueryCmd(ctx, cli.CmdQueryToken(), []string{denom}, &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = types.CurrentTokenVersion
	requireT.Equal(expectedToken, resp.Token)

	// query balance
	var respBalance types.QueryBalanceResponse
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryBalance(), []string{expectedToken.Issuer, denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &respBalance))
	requireT.Equal(initialAmount.String(), respBalance.Balance.String())
}

func TestCmdTokenUpgradeStatuses(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features:    []types.Feature{},
	}
	ctx := testNetwork.Validators[0].ClientCtx

	initialAmount := sdk.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTokenUpgradeStatuses(), []string{denom, "--output", "json"})
	var statusesRes types.QueryTokenUpgradeStatusesResponse
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &statusesRes))
	// we can't check non-empty values
	requireT.Nil(statusesRes.Statuses.V1)
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
