//go:build integrationtests

package ibc

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestIBCAssetFTSendCommissionAndBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t, false)
	requireT := require.New(t)

	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia
	osmosisChain := chains.Osmosis

	gaiaToCoreumChannelID, err := gaiaChain.GetIBCChannelID(ctx, coreumChain.ChainSettings.ChainID)
	requireT.NoError(err)
	coreumToGaiaChannelID, err := coreumChain.GetIBCChannelID(ctx, gaiaChain.ChainSettings.ChainID)
	requireT.NoError(err)
	osmosisToCoreumChannelID, err := osmosisChain.GetIBCChannelID(ctx, coreumChain.ChainSettings.ChainID)
	requireT.NoError(err)
	coreumToOsmosisChannelID, err := coreumChain.GetIBCChannelID(ctx, osmosisChain.ChainSettings.ChainID)
	requireT.NoError(err)

	coreumToGaiaEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToGaiaChannelID)
	coreumToOsmosisEscrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, coreumToOsmosisChannelID)

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient1 := gaiaChain.GenAccount()
	gaiaRecipient2 := gaiaChain.GenAccount()
	osmosisRecipient1 := osmosisChain.GenAccount()
	osmosisRecipient2 := osmosisChain.GenAccount()

	coreumIssuer := coreumChain.GenAccount()
	issueFee := getIssueFee(ctx, t, coreumChain.ClientContext).Amount
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	}))

	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
			&ibctransfertypes.MsgTransfer{},
		},
	}))

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             coreumIssuer.String(),
		Symbol:             "mysymbol",
		Subunit:            "mysubunit",
		Precision:          8,
		InitialAmount:      sdk.NewInt(1_000_000),
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

	sendCoin := sdk.NewCoin(denom, sdk.NewInt(1000))
	burntAmount := issueMsg.BurnRate.Mul(sendCoin.Amount.ToDec()).TruncateInt()
	sendCommissionAmount := issueMsg.SendCommissionRate.Mul(sendCoin.Amount.ToDec()).TruncateInt()
	extraAmount := sdk.NewInt(77) // some amount to be left at the end
	msgSend := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		// amount to send + burn rate + send commission rate + some amount to test with none-zero reminder
		Amount: sdk.NewCoins(sdk.NewCoin(denom,
			sendCoin.Amount.MulRaw(2).
				Add(burntAmount.MulRaw(2)).
				Add(sendCommissionAmount.MulRaw(2)).
				Add(extraAmount)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	receiveCoinGaia := sdk.NewCoin(convertToIBCDenom(gaiaToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)
	receiveCoinOsmosis := sdk.NewCoin(convertToIBCDenom(osmosisToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)

	// ********** Coreum to Gaia ********** //
	// IBC transfer from FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		gaiaChain.ChainContext,
		gaiaRecipient1,
		receiveCoinGaia,
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumIssuer):              sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount,
		},
		map[string]sdk.Int{
			gaiaChain.ConvertToBech32Address(gaiaRecipient1): sendCoin.Amount,
		},
	)

	// IBC transfer from non-FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		gaiaChain.ChainContext,
		gaiaRecipient2,
		receiveCoinGaia,
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumSender):              sendCoin.Amount.Add(sendCommissionAmount).Add(burntAmount).Neg(),
			coreumChain.ConvertToBech32Address(coreumIssuer):              sendCommissionAmount,
			coreumChain.ConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount,
		},
		map[string]sdk.Int{
			gaiaChain.ConvertToBech32Address(gaiaRecipient2): sendCoin.Amount,
		},
	)

	// ********** Coreum to Osmosis ********** //
	// IBC transfer from FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		osmosisChain.ChainContext,
		osmosisRecipient1,
		receiveCoinOsmosis,
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumIssuer):                 sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount,
		},
		map[string]sdk.Int{
			osmosisChain.ConvertToBech32Address(osmosisRecipient1): sendCoin.Amount,
		},
	)

	// IBC transfer from non-FT issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		osmosisChain.ChainContext,
		osmosisRecipient2,
		receiveCoinOsmosis,
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumSender):                 sendCoin.Amount.Add(sendCommissionAmount).Add(burntAmount).Neg(),
			coreumChain.ConvertToBech32Address(coreumIssuer):                 sendCommissionAmount,
			coreumChain.ConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount,
		},
		map[string]sdk.Int{
			osmosisChain.ConvertToBech32Address(osmosisRecipient2): sendCoin.Amount,
		},
	)

	// ********** Gaia to Coreum (send back) ********** //
	// IBC transfer back to issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		gaiaChain.ChainContext,
		gaiaRecipient1,
		receiveCoinGaia,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		map[string]sdk.Int{
			gaiaChain.ConvertToBech32Address(gaiaRecipient1): sendCoin.Amount.Neg(),
		},
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumIssuer):              sendCoin.Amount,
		},
	)

	// IBC transfer back to non-issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		gaiaChain.ChainContext,
		gaiaRecipient2,
		receiveCoinGaia,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		map[string]sdk.Int{
			gaiaChain.ConvertToBech32Address(gaiaRecipient2): sendCoin.Amount.Neg(),
		},
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumToGaiaEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumSender):              sendCoin.Amount,
			coreumChain.ConvertToBech32Address(coreumIssuer):              sdk.ZeroInt(),
		},
	)

	// ********** Osmosis to Coreum (send back) ********** //
	// IBC transfer back to issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		osmosisChain.ChainContext,
		osmosisRecipient1,
		receiveCoinOsmosis,
		coreumChain.ChainContext,
		coreumIssuer,
		sendCoin,
		map[string]sdk.Int{
			osmosisChain.ConvertToBech32Address(osmosisRecipient1): sendCoin.Amount.Neg(),
		},
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumIssuer):                 sendCoin.Amount,
		},
	)

	// IBC transfer back to non-issuer address.
	ibcTransferAndAssertBalanceChanges(
		ctx,
		t,
		osmosisChain.ChainContext,
		osmosisRecipient2,
		receiveCoinOsmosis,
		coreumChain.ChainContext,
		coreumSender,
		sendCoin,
		map[string]sdk.Int{
			osmosisChain.ConvertToBech32Address(osmosisRecipient2): sendCoin.Amount.Neg(),
		},
		map[string]sdk.Int{
			coreumChain.ConvertToBech32Address(coreumToOsmosisEscrowAddress): sendCoin.Amount.Neg(),
			coreumChain.ConvertToBech32Address(coreumSender):                 sendCoin.Amount,
			coreumChain.ConvertToBech32Address(coreumIssuer):                 sdk.ZeroInt(),
		},
	)
}

func ibcTransferAndAssertBalanceChanges(
	ctx context.Context,
	t *testing.T,
	srcChainCtx integrationtests.ChainContext,
	srcSender sdk.AccAddress,
	sendCoin sdk.Coin,
	dstChainCtx integrationtests.ChainContext,
	dstChainRecipient sdk.AccAddress,
	receiveCoin sdk.Coin,
	srcExpectedBalanceChanges map[string]sdk.Int,
	dstExpectedBalanceChanges map[string]sdk.Int,
) {
	requireT := require.New(t)

	srcBalancesBeforeOperation := fetchBalanceForMultipleAddresses(ctx, t, srcChainCtx, sendCoin.Denom, lo.Keys(srcExpectedBalanceChanges))
	dstBalancesBeforeOperation := fetchBalanceForMultipleAddresses(ctx, t, dstChainCtx, receiveCoin.Denom, lo.Keys(dstExpectedBalanceChanges))

	dstBankClient := banktypes.NewQueryClient(dstChainCtx.ClientContext)
	dstChainRecipientBalanceBefore, err := dstBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: dstChainCtx.ConvertToBech32Address(dstChainRecipient),
		Denom:   receiveCoin.Denom,
	})
	requireT.NoError(err)
	dstChainRecipientBalanceExpected := dstChainRecipientBalanceBefore.Balance.Add(receiveCoin)

	_, err = srcChainCtx.ExecuteIBCTransfer(ctx, srcSender, sendCoin, dstChainCtx, dstChainRecipient)
	requireT.NoError(err)

	err = dstChainCtx.AwaitForBalance(ctx, dstChainRecipient, dstChainRecipientBalanceExpected)
	requireT.NoError(err)

	srcBalancesAfterOperation := fetchBalanceForMultipleAddresses(ctx, t, srcChainCtx, sendCoin.Denom, lo.Keys(srcExpectedBalanceChanges))
	dstBalancesAfterOperation := fetchBalanceForMultipleAddresses(ctx, t, dstChainCtx, receiveCoin.Denom, lo.Keys(dstExpectedBalanceChanges))

	assertBalanceChanges(t, srcExpectedBalanceChanges, srcBalancesBeforeOperation, srcBalancesAfterOperation)
	assertBalanceChanges(t, dstExpectedBalanceChanges, dstBalancesBeforeOperation, dstBalancesAfterOperation)
}

func fetchBalanceForMultipleAddresses(
	ctx context.Context,
	t *testing.T,
	chainCtx integrationtests.ChainContext,
	denom string,
	addresses []string,
) map[string]sdk.Int {
	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chainCtx.ClientContext)
	balances := make(map[string]sdk.Int, len(addresses))

	for _, addr := range addresses {
		balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: addr,
			Denom:   denom,
		})
		requireT.NoError(err)
		requireT.NotNil(balance.Balance)
		balances[addr] = balance.Balance.Amount
	}

	return balances
}

func assertBalanceChanges(t *testing.T, expectedBalanceChanges, balancesBefore, balancesAfter map[string]sdk.Int) {
	requireT := require.New(t)

	for addr, expectedBalanceChange := range expectedBalanceChanges {
		actualBalanceChange := balancesAfter[addr].Sub(balancesBefore[addr])
		requireT.Equal(expectedBalanceChange.String(), actualBalanceChange.String())
	}
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
