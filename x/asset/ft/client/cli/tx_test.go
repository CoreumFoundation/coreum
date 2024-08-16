package cli_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmcli "github.com/CosmWasm/wasmd/x/wasm/client/cli"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
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
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.2"),
		Version:            0,
		URI:                "https://my-token-meta.invalid/1 ",
		URIHash:            "e000624",
	}

	ctx := testNetwork.Validators[0].ClientCtx
	initialAmount := sdkmath.NewInt(100)
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	var resp types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryToken(), []string{denom}, &resp)
	// set generated values
	token.Denom = denom
	token.Issuer = resp.Token.Issuer
	token.Version = resp.Token.Version
	token.Admin = resp.Token.Admin
	requireT.Equal(token, resp.Token)
}

func TestIssueWithExtension(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	args := []string{
		"../../keeper/test-contracts/asset-extension/artifacts/asset_extension.wasm",
		fmt.Sprintf("--%s=%s", flags.FlagGas, "auto"),
	}
	args = append(args, txValidator1Args(testNetwork)...)
	res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, wasmcli.StoreCodeCmd(), args)
	requireT.NoError(err)

	codeID, err := event.FindUint64EventAttribute(res.Events, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
	requireT.NoError(err)

	token := types.Token{
		Symbol:      "btc" + uuid.NewString()[:4],
		Subunit:     "satoshi" + uuid.NewString()[:4],
		Precision:   8,
		Description: "description",
		Features: []types.Feature{
			types.Feature_burning,
			types.Feature_extension,
		},
		BurnRate:           sdkmath.LegacyMustNewDecFromStr("0.1"),
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.2"),
		Version:            0,
		URI:                "https://my-token-meta.invalid/1 ",
		URIHash:            "e000624",
	}

	//nolint:tagliatelle // these will be exposed to rust and must be snake case.
	issuanceMsg := struct {
		ExtraData string `json:"extra_data"`
	}{
		ExtraData: "test",
	}

	issuanceMsgBytes, err := json.Marshal(issuanceMsg)
	requireT.NoError(err)

	initialAmount := sdkmath.NewInt(100)
	extension := &types.ExtensionIssueSettings{
		CodeId:      codeID,
		Label:       "testing-extension",
		Funds:       sdk.NewCoins(sdk.NewCoin(testNetwork.Config.BondDenom, sdkmath.NewInt(10))),
		IssuanceMsg: issuanceMsgBytes,
	}
	denom := issue(requireT, ctx, token, initialAmount, extension, testNetwork)

	var resp types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryToken(), []string{denom}, &resp)
	// set generated values
	token.Denom = denom
	token.Issuer = resp.Token.Issuer
	token.Version = resp.Token.Version
	token.Admin = resp.Token.Admin
	token.ExtensionCWAddress = resp.Token.ExtensionCWAddress
	requireT.Equal(token, resp.Token)
	requireT.NotEmpty(resp.Token.ExtensionCWAddress)

	args = []string{resp.Token.ExtensionCWAddress, `{"query_issuance_msg":{}}`}
	var queryResp wasmtypes.QuerySmartContractStateResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, wasmcli.GetCmdGetContractStateSmart(), args, &queryResp)
	requireT.NoError(json.Unmarshal(queryResp.Data, &issuanceMsg))
	requireT.Equal("test", issuanceMsg.ExtraData)
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)
	issuer := testNetwork.Validators[0].Address

	// mint new tokens
	coinToMint := sdk.NewInt64Coin(denom, 100)
	args := append([]string{coinToMint.String()}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	var balanceRes banktypes.QueryAllBalancesResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "balances", issuer.String()}, &balanceRes)
	requireT.Equal(sdkmath.NewInt(877).String(), balanceRes.Balances.AmountOf(denom).String())

	var supplyRes banktypes.QuerySupplyOfResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "total-supply-of", denom}, &supplyRes)
	requireT.Equal(sdk.NewInt64Coin(denom, 877).String(), supplyRes.Amount.String())

	// mint to recipient
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	args = append([]string{coinToMint.String(), "--recipient", recipient.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxMint(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "balances", recipient.String()}, &balanceRes)
	requireT.Equal(sdkmath.NewInt(100).String(), balanceRes.Balances.AmountOf(denom).String())

	// burn tokens
	coinToMint = sdk.NewInt64Coin(denom, 200)
	args = append([]string{coinToMint.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxBurn(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "balances", issuer.String()}, &balanceRes)
	requireT.Equal(sdkmath.NewInt(677).String(), balanceRes.Balances.AmountOf(denom).String())

	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "total-supply-of", denom}, &supplyRes)
	requireT.Equal(sdk.NewInt64Coin(denom, 777).String(), supplyRes.Amount.String())
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)
	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// freeze part of the token
	coinToFreeze := sdk.NewInt64Coin(denom, 100)
	args := append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
	requireT.NoError(err)

	// query frozen balance
	var respFrozen types.QueryFrozenBalanceResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	)
	requireT.Equal(coinToFreeze.String(), respFrozen.Balance.String())

	// query balance
	var respBalance types.QueryBalanceResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryBalance(),
		[]string{recipient.String(), denom},
		&respBalance,
	)
	requireT.Equal(coinToFreeze.Amount.String(), respBalance.Frozen.String())

	// issue and freeze more to test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		newDenom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)
		coinToFreeze = sdk.NewInt64Coin(newDenom, 100)
		args = append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
		_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
		requireT.NoError(err)
	}

	var balancesResp types.QueryFrozenBalancesResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryFrozenBalances(),
		[]string{recipient.String()},
		&balancesResp,
	)
	requireT.Len(balancesResp.Balances, 3)

	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryFrozenBalances(),
		[]string{recipient.String(), "--limit", "1"},
		&balancesResp,
	)
	requireT.Len(balancesResp.Balances, 1)

	// unfreeze part of the frozen token
	unfreezeTokens := sdk.NewInt64Coin(denom, 75)
	args = append([]string{recipient.String(), unfreezeTokens.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxUnfreeze(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	)
	requireT.Equal(sdk.NewInt64Coin(denom, 25).String(), respFrozen.Balance.String())

	// set absolute frozen amount
	setFrozenTokens := sdk.NewInt64Coin(denom, 100)
	args = append([]string{recipient.String(), setFrozenTokens.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxSetFrozen(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryFrozenBalance(),
		[]string{recipient.String(), denom},
		&respFrozen,
	)
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	// globally freeze the token
	args := append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxGloballyFreeze(), args)
	requireT.NoError(err)

	var resp types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryToken(), []string{denom}, &resp)
	requireT.True(resp.Token.GloballyFrozen)

	// globally unfreeze the token
	args = append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxGloballyUnfreeze(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryToken(), []string{denom}, &resp)
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)
	account := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	coin := sdk.NewInt64Coin(denom, 100)

	valAddr := testNetwork.Validators[0].Address.String()
	args := append(
		[]string{valAddr, account.String(), coin.String()},
		txValidator1Args(testNetwork)...,
	)
	_, err := coreumclitestutil.ExecTxCmd(
		ctx,
		testNetwork,
		bankcli.NewSendTxCmd(authcodec.NewBech32Codec(app.ChosenNetwork.Provider.GetAddressPrefix())),
		args,
	)
	requireT.NoError(err)

	var balanceRes banktypes.QueryAllBalancesResponse
	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "balances", account.String()}, &balanceRes)
	requireT.Equal(sdkmath.NewInt(100).String(), balanceRes.Balances.AmountOf(denom).String())

	args = append([]string{account.String(), coin.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClawback(), args)
	requireT.NoError(err)

	coreumclitestutil.ExecRootQueryCmd(t, ctx, []string{banktypes.ModuleName, "balances", account.String()}, &balanceRes)
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.Balances.AmountOf(denom).String())
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
	_ = issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// test pagination
	for i := 0; i < 2; i++ {
		token.Symbol = fmt.Sprintf("btc%d%s", i, uuid.NewString()[:4])
		token.Subunit = fmt.Sprintf("satoshi%d%s", i, uuid.NewString()[:4])
		denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

		coinToWhitelist := sdk.NewInt64Coin(denom, 100)
		args := append([]string{recipient.String(), coinToWhitelist.String()}, txValidator1Args(testNetwork)...)
		_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxSetWhitelistedLimit(), args)
		requireT.NoError(err)

		// query whitelisted balance
		var respWhitelisted types.QueryWhitelistedBalanceResponse
		coreumclitestutil.ExecQueryCmd(
			t,
			ctx,
			cli.CmdQueryWhitelistedBalance(),
			[]string{recipient.String(), denom},
			&respWhitelisted,
		)
		requireT.Equal(coinToWhitelist.String(), respWhitelisted.Balance.String())

		// query balance
		var respBalance types.QueryBalanceResponse
		coreumclitestutil.ExecQueryCmd(
			t,
			ctx,
			cli.CmdQueryBalance(),
			[]string{recipient.String(), denom},
			&respBalance,
		)
		requireT.Equal(coinToWhitelist.Amount.String(), respBalance.Whitelisted.String())
	}

	var balancesResp types.QueryWhitelistedBalancesResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryWhitelistedBalances(),
		[]string{recipient.String()},
		&balancesResp,
	)
	requireT.Len(balancesResp.Balances, 2)

	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryWhitelistedBalances(),
		[]string{recipient.String(), "--limit", "1"},
		&balancesResp,
	)
	requireT.Len(balancesResp.Balances, 1)
}

func TestTransferAdmin(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	newAdmin := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	// transfer admin from issuer to new admin
	args := append([]string{newAdmin.String(), denom}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxTransferAdmin(), args)
	requireT.NoError(err)

	// query token
	var respToken types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryToken(),
		[]string{denom},
		&respToken,
	)
	requireT.Equal(newAdmin.String(), respToken.Token.Admin)

	// try to transfer admin from issuer to new admin again
	args = append([]string{newAdmin.String(), denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxTransferAdmin(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// try to freeze part of the token by issuer which is not admin anymore
	coinToFreeze := sdk.NewInt64Coin(denom, 100)
	args = append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
}

func TestClearAdmin(t *testing.T) {
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

	// clear admin
	args := append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClearAdmin(), args)
	requireT.NoError(err)

	// query token
	var respToken types.QueryTokenResponse
	coreumclitestutil.ExecQueryCmd(
		t,
		ctx,
		cli.CmdQueryToken(),
		[]string{denom},
		&respToken,
	)
	requireT.Empty(respToken.Token.Admin)

	// try to clear admin again
	args = append([]string{denom}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxClearAdmin(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)

	recipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// try to freeze part of the token by previous admin
	coinToFreeze := sdk.NewInt64Coin(denom, 100)
	args = append([]string{recipient.String(), coinToFreeze.String()}, txValidator1Args(testNetwork)...)
	_, err = coreumclitestutil.ExecTxCmd(ctx, testNetwork, cli.CmdTxFreeze(), args)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
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
	denom := issue(requireT, ctx, token, initialAmount, nil, testNetwork)

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
	extensionSettings *types.ExtensionIssueSettings,
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
	if extensionSettings != nil && extensionSettings.CodeId > 0 {
		args = parseExtensionArgs(args, extensionSettings)
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

func parseExtensionArgs(args []string, extensionSettings *types.ExtensionIssueSettings) []string {
	args = append(args,
		fmt.Sprintf("--%s=%d", cli.ExtensionCodeID, extensionSettings.CodeId),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 2000000),
	)
	if len(extensionSettings.Label) > 0 {
		args = append(args, fmt.Sprintf("--%s=%s", cli.ExtensionLabel, extensionSettings.Label))
	}
	if extensionSettings.Funds != nil && extensionSettings.Funds.IsAllPositive() {
		args = append(args, fmt.Sprintf("--%s=%s", cli.ExtensionFunds, extensionSettings.Funds.String()))
	}
	if len(extensionSettings.IssuanceMsg) > 0 {
		if jsonEncodedMessage, err := extensionSettings.IssuanceMsg.MarshalJSON(); err == nil {
			args = append(args, fmt.Sprintf("--%s=%s", cli.ExtensionIssuanceMsg, string(jsonEncodedMessage)))
		}
	}
	return args
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, testNetwork.Validators[0].Address.String()),
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewInt64Coin(testNetwork.Config.BondDenom, 1000000)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}
