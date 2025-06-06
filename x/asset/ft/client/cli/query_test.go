package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v6/testutil/cli"
	"github.com/CoreumFoundation/coreum/v6/testutil/network"
	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
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
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.2"),
	}

	ctx := testNetwork.Validators[0].ClientCtx

	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	var resp types.QueryTokensResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryTokens(),
		[]string{issuer.String(), "--limit", "1"},
		&resp,
	)

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = types.CurrentTokenVersion
	expectedToken.Admin = testNetwork.Validators[0].Address.String()
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
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.2"),
	}
	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	var resp types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryToken(), []string{denom}, &resp)

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
	expectedToken.Version = types.CurrentTokenVersion
	expectedToken.Admin = testNetwork.Validators[0].Address.String()
	requireT.Equal(expectedToken, resp.Token)

	// query balance
	var respBalance types.QueryBalanceResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryBalance(),
		[]string{expectedToken.Issuer, denom},
		&respBalance,
	)
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	var statusesRes types.QueryTokenUpgradeStatusesResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdTokenUpgradeStatuses(), []string{denom}, &statusesRes)
	// we can't check non-empty values
	requireT.Nil(statusesRes.Statuses.V1)
}

func TestQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	var resp types.QueryParamsResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryParams(), []string{}, &resp)
	expectedIssueFee := sdk.Coin{Denom: constant.DenomDev, Amount: sdkmath.NewInt(10_000_000)}
	requireT.Equal(expectedIssueFee, resp.Params.IssueFee)
}
