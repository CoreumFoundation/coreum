package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v3/testutil/cli"
	"github.com/CoreumFoundation/coreum/v3/testutil/network"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
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
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryTokens(), []string{issuer.String(), "--limit", "1"}, &resp))

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
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryToken(), []string{denom}, &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = types.CurrentTokenVersion
	requireT.Equal(expectedToken, resp.Token)

	// query balance
	var respBalance types.QueryBalanceResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryBalance(), []string{expectedToken.Issuer, denom}, &respBalance))
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

	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	var statusesRes types.QueryTokenUpgradeStatusesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdTokenUpgradeStatuses(), []string{denom}, &statusesRes))
	// we can't check non-empty values
	requireT.Nil(statusesRes.Statuses.V1)
}

func TestQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	var resp types.QueryParamsResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryParams(), []string{}, &resp))
	expectedIssueFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdkmath.NewInt(10_000_000)}
	requireT.Equal(expectedIssueFee, resp.Params.IssueFee)
}
