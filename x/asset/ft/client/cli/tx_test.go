package cli_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/app"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
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
			types.Feature_burning,
		},
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
		Version:            0,
		URI:                "https://my-token-meta.invalid/1 ",
		URIHash:            "e000624",
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	var resp types.QueryTokenResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryToken(), []string{denom}, &resp))
	// set generated values
	token.Denom = denom
	token.Issuer = resp.Token.Issuer
	token.Version = resp.Token.Version
	requireT.Equal(token, resp.Token)
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
			types.Feature_minting,
			types.Feature_burning,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	issuer := testNetwork.Validators[0].Address

	// mint new tokens
	coinToMint := sdk.NewInt64Coin(denom, 100)
	args := append([]string{coinToMint.String()}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	var balanceRsp banktypes.QueryAllBalancesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String()}, &balanceRsp))
	requireT.Equal(sdkmath.NewInt(877).String(), balanceRsp.Balances.AmountOf(denom).String())

	var supplyRsp sdk.Coin
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		bankcli.GetCmdQueryTotalSupply(),
		[]string{"--denom", denom},
		&supplyRsp,
	))
	requireT.Equal(sdk.NewInt64Coin(denom, 877).String(), supplyRsp.String())

	// mint to recipient
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	args = append([]string{coinToMint.String(), "--recipient", recipient.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		bankcli.GetBalancesCmd(),
		[]string{recipient.String()},
		&balanceRsp,
	))
	requireT.Equal(sdkmath.NewInt(100).String(), balanceRsp.Balances.AmountOf(denom).String())

	// burn tokens
	coinToMint = sdk.NewInt64Coin(denom, 200)
	args = append([]string{coinToMint.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, bankcli.GetBalancesCmd(), []string{issuer.String()}, &balanceRsp))
	requireT.Equal(sdkmath.NewInt(677).String(), balanceRsp.Balances.AmountOf(denom).String())

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		bankcli.GetCmdQueryTotalSupply(),
		[]string{"--denom", denom},
		&supplyRsp,
	))
	requireT.Equal(sdk.NewInt64Coin(denom, 777).String(), supplyRsp.String())
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
			types.Feature_freezing,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze part of the token
	coinToFreeze := sdk.NewInt64Coin(denom, 100)
	args := append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
	requireT.NoError(err)

	// query frozen balance
	var respFrozen types.QueryFrozenBalanceResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	))
	requireT.Equal(coinToFreeze.String(), respFrozen.Balance.String())

	// query balance
	var respBalance types.QueryBalanceResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryBalance(),
		[]string{recipient.String(), denom},
		&respBalance,
	))
	requireT.Equal(coinToFreeze.Amount.String(), respBalance.Frozen.String())

	// issue and freeze more to test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		newDenom := issue(requireT, ctx, token, initialAmount, testNetwork)
		coinToFreeze = sdk.NewInt64Coin(newDenom, 100)
		args = append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
		_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
		requireT.NoError(err)
	}

	var balancesResp types.QueryFrozenBalancesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryFrozenBalances(),
		[]string{recipient.String()},
		&balancesResp,
	))
	requireT.Len(balancesResp.Balances, 3)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryFrozenBalances(),
		[]string{recipient.String(), "--limit", "1"},
		&balancesResp,
	))
	requireT.Len(balancesResp.Balances, 1)

	// unfreeze part of the frozen token
	unfreezeTokens := sdk.NewInt64Coin(denom, 75)
	args = append([]string{recipient.String(), unfreezeTokens.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	))
	requireT.Equal(sdk.NewInt64Coin(denom, 25).String(), respFrozen.Balance.String())

	// set absolute frozen amount
	setFrozenTokens := sdk.NewInt64Coin(denom, 100)
	args = append([]string{recipient.String(), setFrozenTokens.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxSetFrozen(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	))
	requireT.Equal(setFrozenTokens.String(), respFrozen.Balance.String())
}

func TestGloballyFreezeUnfreeze(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_freezing,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	// globally freeze the token
	args := append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxGloballyFreeze(), args)
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryToken(), []string{denom}, &resp))
	requireT.True(resp.Token.GloballyFrozen)

	// globally unfreeze the token
	args = append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxGloballyUnfreeze(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryToken(), []string{denom}, &resp))
	requireT.False(resp.Token.GloballyFrozen)
}

func TestClawback(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_clawback,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)
	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	coin := sdk.NewInt64Coin(denom, 100)

	args := append([]string{testNetwork.Validators[0].Address.String(), account.String(), coin.String()}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, bankcli.NewSendTxCmd(), args)
	requireT.NoError(err)

	var balanceRsp banktypes.QueryAllBalancesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, bankcli.GetBalancesCmd(), []string{account.String()}, &balanceRsp))
	requireT.Equal(sdkmath.NewInt(100).String(), balanceRsp.Balances.AmountOf(denom).String())

	args = append([]string{account.String(), coin.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClawback(), args)
	requireT.NoError(err)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, bankcli.GetBalancesCmd(), []string{account.String()}, &balanceRsp))
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRsp.Balances.AmountOf(denom).String())
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
			types.Feature_whitelisting,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	_ = issue(requireT, ctx, token, initialAmount, testNetwork)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		denom := issue(requireT, ctx, token, initialAmount, testNetwork)

		coinToWhitelist := sdk.NewInt64Coin(denom, 100)
		args := append([]string{recipient.String(), coinToWhitelist.String()}, txValidator1Args(testNetwork)...)
		_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxSetWhitelistedLimit(), args)
		requireT.NoError(err)

		// query whitelisted balance
		var respWhitelisted types.QueryWhitelistedBalanceResponse
		requireT.NoError(coreumclitestutil.ExecQueryCmd(
			ctx,
			cli.CmdQueryWhitelistedBalance(),
			[]string{recipient.String(), denom},
			&respWhitelisted,
		))
		requireT.Equal(coinToWhitelist.String(), respWhitelisted.Balance.String())

		// query balance
		var respBalance types.QueryBalanceResponse
		requireT.NoError(coreumclitestutil.ExecQueryCmd(
			ctx,
			cli.CmdQueryBalance(),
			[]string{recipient.String(), denom},
			&respBalance,
		))
		requireT.Equal(coinToWhitelist.Amount.String(), respBalance.Whitelisted.String())
	}

	var balancesResp types.QueryWhitelistedBalancesResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryWhitelistedBalances(),
		[]string{recipient.String()},
		&balancesResp,
	))
	requireT.Len(balancesResp.Balances, 2)

	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx,
		cli.CmdQueryWhitelistedBalances(),
		[]string{recipient.String(), "--limit", "1"},
		&balancesResp,
	))
	requireT.Len(balancesResp.Balances, 1)
}

func TestUpgradeV1(t *testing.T) {
	requireT := require.New(t)
	networkCfg, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	app.ChosenNetwork = networkCfg
	testNetwork := network.New(t)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_freezing,
			types.Feature_ibc,
		},
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(777)
	denom := issue(requireT, ctx, token, initialAmount, testNetwork)

	// --ibc-enabled is missing
	args := append([]string{
		denom,
	}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUpgradeV1(), args)
	requireT.Error(err)

	// upgrade the token with --ibc-enabled=true
	args = append([]string{
		denom,
		fmt.Sprintf("--%s=true", cli.IBCEnabledFlag),
	}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUpgradeV1(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	// upgrade the token with --ibc-enabled=false
	args = append([]string{
		denom,
		fmt.Sprintf("--%s=false", cli.IBCEnabledFlag),
	}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUpgradeV1(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

func issue(
	requireT *require.Assertions,
	ctx client.Context,
	token types.Token,
	initialAmount sdkmath.Int,
	testNetwork *network.Network,
) string {
	features := make([]string, 0, len(token.Features))
	for _, feature := range token.Features {
		features = append(features, feature.String())
	}
	// args
	args := []string{
		token.Symbol,
		token.Subunit,
		strconv.FormatUint(uint64(token.Precision), 10),
		initialAmount.String(),
		token.Description,
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
	if token.URI != "" {
		args = append(args, fmt.Sprintf("--%s=%s", cli.URIFlag, token.URI))
	}
	if token.URIHash != "" {
		args = append(args, fmt.Sprintf("--%s=%s", cli.URIHashFlag, token.URIHash))
	}

	args = append(args, txValidator1Args(testNetwork)...)
	res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxIssue(), args)
	requireT.NoError(err)

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
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewInt64Coin(testNetwork.Config.BondDenom, 1000000)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
