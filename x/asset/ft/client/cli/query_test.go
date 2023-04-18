package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	testutilcli "github.com/CoreumFoundation/coreum/testutil/cli"
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

	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	var resp types.QueryTokensResponse
	requireT.NoError(testutilcli.ExecQueryCmd(ctx, cli.CmdQueryTokens(), []string{issuer.String(), "--limit", "1"}, &resp))

	expectedToken := token
	expectedToken.Denom = denom
	expectedToken.Issuer = testNetwork.Validators[0].Address.String()
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
	requireT.Equal(expectedToken, resp.Token)
}
