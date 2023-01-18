package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/testutil/event"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestIssue(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_burning, //nolint:nosnakecase
		},
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdk.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	requireT.Equal(types.BuildDenom(token.Subunit, testNetwork.Validators[0].Address), denom)
}

func TestMintBurn(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_minting, //nolint:nosnakecase
			types.Feature_burning, //nolint:nosnakecase
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdk.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	issuer := testNetwork.Validators[0].Address

	// mint new tokens
	coinToMint := sdk.NewInt64Coin(denom, 100)
	args := append([]string{coinToMint.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxMint(), args)
	requireT.NoError(err)

	var balanceRsp banktypes.QueryAllBalancesResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balanceRsp))
	requireT.Equal(sdk.NewInt(877).String(), balanceRsp.Balances.AmountOf(denom).String())

	var supplyRsp sdk.Coin
	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetCmdQueryTotalSupply(), []string{issuer.String(), "--denom", denom, "--output", "json"})
	requireT.NoError(err)
	bs := buf.Bytes()
	requireT.NoError(ctx.Codec.UnmarshalJSON(bs, &supplyRsp))
	requireT.Equal(sdk.NewInt64Coin(denom, 877).String(), supplyRsp.String())

	// burn tokens
	coinToMint = sdk.NewInt64Coin(denom, 200)
	args = append([]string{coinToMint.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balanceRsp))
	requireT.Equal(sdk.NewInt(677).String(), balanceRsp.Balances.AmountOf(denom).String())

	buf, err = clitestutil.ExecTestCLICmd(ctx, bankcli.GetCmdQueryTotalSupply(), []string{issuer.String(), "--denom", denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &supplyRsp))
	requireT.Equal(sdk.NewInt64Coin(denom, 677).String(), supplyRsp.String())
}

func TestFreezeAndQueryFrozen(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_freezing, //nolint:nosnakecase
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdk.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	issuer := testNetwork.Validators[0].Address

	// freeze part of the token
	coinToFreeze := sdk.NewInt64Coin(denom, 100)
	args := append([]string{issuer.String(), coinToFreeze.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxFreeze(), args)
	requireT.NoError(err)

	var resp types.QueryFrozenBalanceResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozenBalance(), []string{issuer.String(), denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.Equal(coinToFreeze.String(), resp.Balance.String())

	// issue and freeze more to test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		newDenom := issue(requireT, ctx, token, initialAmount, testNetwork)
		coinToFreeze = sdk.NewInt64Coin(newDenom, 100)
		args = append([]string{issuer.String(), coinToFreeze.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
		_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxFreeze(), args)
		requireT.NoError(err)
	}

	var balancesResp types.QueryFrozenBalancesResponse
	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozenBalances(), []string{issuer.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 3)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozenBalances(), []string{issuer.String(), "--output", "json", "--limit", "1"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 1)

	// unfreeze part of the frozen token
	unfreezeTokens := sdk.NewInt64Coin(denom, 75)
	args = append([]string{issuer.String(), unfreezeTokens.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryFrozenBalance(), []string{issuer.String(), denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))

	requireT.Equal(sdk.NewInt64Coin(denom, 25).String(), resp.Balance.String())
}

func TestGloballyFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_freezing, //nolint:nosnakecase
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdk.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	// globally freeze the token
	args := append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGloballyFreeze(), args)
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.True(resp.Token.GloballyFrozen)

	// globally unfreeze the token
	args = append([]string{denom, "--output", "json"}, txValidator1Args(testNetwork)...)
	_, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdTxGloballyUnfreeze(), args)
	requireT.NoError(err)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryToken(), []string{denom, "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &resp))
	requireT.False(resp.Token.GloballyFrozen)
}

func TestWhitelistAndQueryWhitelisted(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_whitelisting, //nolint:nosnakecase
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdk.NewInt(777)
	_ = issue(requireT, ctx, token, initialAmount, testNetwork)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		denom := issue(requireT, ctx, token, initialAmount, testNetwork)

		coinToWhitelist := sdk.NewInt64Coin(denom, 100)
		args := append([]string{recipient.String(), coinToWhitelist.String(), "--output", "json"}, txValidator1Args(testNetwork)...)
		_, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxSetWhitelistedLimit(), args)
		requireT.NoError(err)

		var balancesResp types.QueryWhitelistedBalanceResponse
		buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryWhitelistedBalance(), []string{recipient.String(), denom, "--output", "json"})
		requireT.NoError(err)
		requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
		requireT.Equal(coinToWhitelist.String(), balancesResp.Balance.String())
	}

	var balancesResp types.QueryWhitelistedBalancesResponse
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryWhitelistedBalances(), []string{recipient.String(), "--output", "json"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 2)

	buf, err = clitestutil.ExecTestCLICmd(ctx, cli.CmdQueryWhitelistedBalances(), []string{recipient.String(), "--output", "json", "--limit", "1"})
	requireT.NoError(err)
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &balancesResp))
	requireT.Len(balancesResp.Balances, 1)
}

func issue(requireT *require.Assertions, ctx client.Context, token types.Token, initialAmount sdk.Int, testNetwork *network.Network) string {
	features := make([]string, 0, len(token.Features))
	for _, feature := range token.Features {
		features = append(features, feature.String())
	}
	// args
	args := []string{
		token.Symbol, token.Subunit, fmt.Sprint(token.Precision), initialAmount.String(), token.Description, // args
	}
	// flags
	if len(features) > 0 {
		args = append(args, fmt.Sprintf("--%s=%s", cli.FeaturesFlag, strings.Join(features, ",")))
	}
	if !token.BurnRate.IsNil() {
		args = append(args, fmt.Sprintf("--%s=%s", cli.BurnRateFlag, token.BurnRate.String()))
	}
	if !token.SendCommissionRate.IsNil() {
		args = append(args, fmt.Sprintf("--%s=%s", cli.SendCommissionRateFlag, token.SendCommissionRate.String()))
	}

	args = append(args, txValidator1Args(testNetwork)...)
	buf, err := clitestutil.ExecTestCLICmd(ctx, cli.CmdTxIssue(), args)
	requireT.NoError(err)

	var res sdk.TxResponse
	requireT.NoError(ctx.Codec.UnmarshalJSON(buf.Bytes(), &res))
	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit Issue tx for query", res)

	eventIssuedName := proto.MessageName(&types.EventIssued{})
	for i := range res.Events {
		if res.Events[i].Type != eventIssuedName {
			continue
		}
		eventsTokenIssued, err := event.FindTypedEvents[*types.EventIssued](res.Events)
		requireT.NoError(err)
		return eventsTokenIssued[0].Denom
	}
	requireT.Failf("event: %s not found in the issue response", eventIssuedName)

	return ""
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewInt64Coin(testNetwork.Config.BondDenom, 1000000)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
