package ibc

import (
	"context"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIBCAssetFTSendCommissionAndBurnRate(t *testing.T) {
	t.Parallel()

	channelsInfo := AwaitForIBCConfig(t)
	coreumToGaiaChannelID := channelsInfo.CoreumToGaiaChannelID
	gaiaToCoreumChannelID := channelsInfo.GaiaToCoreumChannelID

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumBankClient := banktypes.NewQueryClient(coreumChain.ClientContext)

	coreumIssuer := coreumChain.GenAccount()
	issueFee := getIssueFee(ctx, t, coreumChain.ClientContext).Amount
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumIssuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
			&ibctransfertypes.MsgTransfer{},
		},
		Amount: issueFee,
	}))

	coreumSender := coreumChain.GenAccount()
	// coreumRecipient := coreumChain.GenAccount()
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&ibctransfertypes.MsgTransfer{},
		},
	}))

	gaiaRecipient := gaiaChain.GenAccount()
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             coreumIssuer.String(),
		Symbol:             "mysymbol",
		Subunit:            "mysubunit",
		Precision:          8,
		InitialAmount:      sdk.NewInt(1_000_000),
		BurnRate:           sdk.MustNewDecFromStr("0.1"),
		SendCommissionRate: sdk.MustNewDecFromStr("0.2"),
	}
	_, err := client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumIssuer),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumIssuer)

	// send some asset ft to the coreum recipient
	sendToGaiaCoin := sdk.NewCoin(denom, sdk.NewInt(1000))
	burntAmount := issueMsg.BurnRate.Mul(sendToGaiaCoin.Amount.ToDec()).TruncateInt()
	sendCommissionAmount := issueMsg.SendCommissionRate.Mul(sendToGaiaCoin.Amount.ToDec()).TruncateInt()
	extraAmount := sdk.NewInt(77) // some amount to be left at the end
	msgSend := &banktypes.MsgSend{
		FromAddress: coreumIssuer.String(),
		ToAddress:   coreumSender.String(),
		// amount to send + burn rate + send commission rate + some amount to test that it's left
		Amount: sdk.NewCoins(sdk.NewCoin(denom, sendToGaiaCoin.Amount.
			Add(burntAmount).
			Add(sendCommissionAmount).
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

	coreumIssuerBalanceBeforeTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	_, err = ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumIssuer, coreumToGaiaChannelID, sendToGaiaCoin, gaiaChain, gaiaRecipient)
	requireT.NoError(err)

	executedGaiaRecipientBalance := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), sendToGaiaCoin.Amount)
	err = AwaitForBalance(ctx, gaiaChain, gaiaRecipient, executedGaiaRecipientBalance)
	requireT.NoError(err)

	// check that we don't apply the commissions since the sender is issuer
	coreumIssuerBalanceAfterTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(
		coreumIssuerBalanceBeforeTransferRes.Balance.Amount.Sub(sendToGaiaCoin.Amount).String(),
		coreumIssuerBalanceAfterTransferRes.Balance.Amount.String(),
	)

	// send form the coreum sender now to apply the commission/burn rates
	_, err = ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumSender, coreumToGaiaChannelID, sendToGaiaCoin, gaiaChain, gaiaRecipient)
	requireT.NoError(err)

	// now we expect the double balance
	executedGaiaRecipientBalance = sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), sendToGaiaCoin.Amount.MulRaw(2))
	err = AwaitForBalance(ctx, gaiaChain, gaiaRecipient, executedGaiaRecipientBalance)
	requireT.NoError(err)

	// check that the amount that is left is the extra amount so the rest was spent on the commission/burn rates
	coreumSenderBalanceAfterTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumSender.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(extraAmount.String(), coreumSenderBalanceAfterTransferRes.Balance.Amount.String())

	// check that issuer has received the commission
	coreumIssuerBalanceAfterSenderTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	requireT.Equal(
		coreumIssuerBalanceAfterTransferRes.Balance.Amount.Add(sendCommissionAmount).String(),
		coreumIssuerBalanceAfterSenderTransferRes.Balance.Amount.String(),
	)

	sentToCoreumCoin := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, denom), sendToGaiaCoin.Amount)
	// send back to the issuer
	coreumIssuerBalanceBeforeTransferBackRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	_, err = ExecuteIBCTransfer(ctx, gaiaChain, gaiaRecipient, gaiaToCoreumChannelID, sentToCoreumCoin, coreumChain.Chain, coreumIssuer)
	requireT.NoError(err)

	expectedCoreumIssuerBalanceAfterTransferBack := sdk.NewCoin(denom, coreumIssuerBalanceBeforeTransferBackRes.Balance.Amount.Add(sentToCoreumCoin.Amount))
	err = AwaitForBalance(ctx, coreumChain.Chain, coreumIssuer, expectedCoreumIssuerBalanceAfterTransferBack)
	requireT.NoError(err)

	// send back to the sender
	coreumSenderBalanceBeforeTransferBackRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumSender.String(),
		Denom:   denom,
	})
	requireT.NoError(err)

	_, err = ExecuteIBCTransfer(ctx, gaiaChain, gaiaRecipient, gaiaToCoreumChannelID, sentToCoreumCoin, coreumChain.Chain, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalanceAfterTransferBack := sdk.NewCoin(denom, coreumSenderBalanceBeforeTransferBackRes.Balance.Amount.Add(sentToCoreumCoin.Amount))
	err = AwaitForBalance(ctx, coreumChain.Chain, coreumSender, expectedCoreumSenderBalanceAfterTransferBack)
	requireT.NoError(err)
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
