package ibc

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestIBCAssetFTSendCommissionAndBurnRate(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t, true)
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
	gaiaRecipient := gaiaChain.GenAccount()
	osmosisRecipient := osmosisChain.GenAccount()

	coreumBankClient := banktypes.NewQueryClient(coreumChain.ClientContext)

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

	// issue balance before the IBC transfer
	coreumIssuerBalanceBeforeIBCTransfersRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)

	// send from issuer and non issuer to gaia
	sendToPeerChainFromCoreumFTIssuerAndNonIssuer(
		ctx, requireT, coreumIssuer, coreumSender, coreumChain.ChainContext, sendCoin, gaiaChain.ChainContext, gaiaRecipient, gaiaToCoreumChannelID, coreumToGaiaEscrowAddress,
	)

	// send from issuer to osmosis
	sendToPeerChainFromCoreumFTIssuerAndNonIssuer(
		ctx, requireT, coreumIssuer, coreumSender, coreumChain.ChainContext, sendCoin, osmosisChain.ChainContext, osmosisRecipient, osmosisToCoreumChannelID, coreumToOsmosisEscrowAddress,
	)

	// validate two commissions
	coreumIssuerBalanceAfterSenderToChainsTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(
		coreumIssuerBalanceBeforeIBCTransfersRes.Balance.Amount.Add(sendCommissionAmount.MulRaw(2)).Sub(sendCoin.Amount.MulRaw(2)).String(),
		coreumIssuerBalanceAfterSenderToChainsTransferRes.Balance.Amount.String(),
	)

	// send back from gaia to validate zero commission
	sendFromPeerChainAndValidateZeroCommissionOnEscrow(ctx, requireT, coreumIssuer, coreumSender, coreumChain.ChainContext, sendCoin, gaiaChain.ChainContext, gaiaRecipient, gaiaToCoreumChannelID, coreumToGaiaEscrowAddress)

	// send back from osmosis to validate zero commission
	sendFromPeerChainAndValidateZeroCommissionOnEscrow(ctx, requireT, coreumIssuer, coreumSender, coreumChain.ChainContext, sendCoin, osmosisChain.ChainContext, osmosisRecipient, osmosisToCoreumChannelID, coreumToOsmosisEscrowAddress)

	// validate two commissions (no additional commission)
	coreumIssuerBalanceAfterSenderToChainsTransferRes, err = coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(
		coreumIssuerBalanceBeforeIBCTransfersRes.Balance.Amount.Add(sendCommissionAmount.MulRaw(2)).String(),
		coreumIssuerBalanceAfterSenderToChainsTransferRes.Balance.Amount.String(),
	)
}

func sendToPeerChainFromCoreumFTIssuerAndNonIssuer(
	ctx context.Context,
	requireT *require.Assertions,
	coreumIssuer sdk.AccAddress,
	coreumSender sdk.AccAddress,
	coreumChainCtx integrationtests.ChainContext,
	sendCoin sdk.Coin,
	peerChainCtx integrationtests.ChainContext,
	peerChainRecipient sdk.AccAddress,
	peerChainToCoreumChannelID string,
	coreumToPeerChainEscrowAddress sdk.AccAddress) {
	coreumBankClient := banktypes.NewQueryClient(coreumChainCtx.ClientContext)
	coreumIssuerBalanceBeforeTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)

	_, err = coreumChainCtx.ExecuteIBCTransfer(ctx, coreumIssuer, sendCoin, peerChainCtx, peerChainRecipient)
	requireT.NoError(err)
	expectedRecipientBalance := sdk.NewCoin(integrationtests.ConvertToIBCDenom(peerChainToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)
	err = peerChainCtx.AwaitForBalance(ctx, peerChainRecipient, expectedRecipientBalance)
	requireT.NoError(err)
	// check that amount is locked on the escrow account
	escrowAddressRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChainCtx.ConvertToBech32Address(coreumToPeerChainEscrowAddress),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(sendCoin.Amount.String(), escrowAddressRes.Balance.Amount.String())
	// check that we don't apply the commissions since the sender is issuer
	coreumIssuerBalanceAfterTransferRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(
		coreumIssuerBalanceBeforeTransferRes.Balance.Amount.Sub(sendCoin.Amount).String(),
		coreumIssuerBalanceAfterTransferRes.Balance.Amount.String(),
	)

	// send from non-issuer
	_, err = coreumChainCtx.ExecuteIBCTransfer(ctx, coreumSender, sendCoin, peerChainCtx, peerChainRecipient)
	requireT.NoError(err)

	expectedOsmosisRecipientBalance := sdk.NewCoin(integrationtests.ConvertToIBCDenom(peerChainToCoreumChannelID, sendCoin.Denom), sendCoin.Amount.MulRaw(2))
	err = peerChainCtx.AwaitForBalance(ctx, peerChainRecipient, expectedOsmosisRecipientBalance)
	requireT.NoError(err)

	// validate escrow balance on the osmosis channel
	coreumToOsmosisEscrowAddressRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChainCtx.ConvertToBech32Address(coreumToPeerChainEscrowAddress),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(sendCoin.Amount.MulRaw(2).String(), coreumToOsmosisEscrowAddressRes.Balance.Amount.String())
}

func sendFromPeerChainAndValidateZeroCommissionOnEscrow(
	ctx context.Context,
	requireT *require.Assertions,
	coreumIssuer sdk.AccAddress,
	coreumSender sdk.AccAddress,
	coreumChainCtx integrationtests.ChainContext,
	sendCoin sdk.Coin,
	peerChainCtx integrationtests.ChainContext,
	peerChainRecipient sdk.AccAddress,
	peerChainToCoreumChannelID string,
	coreumToPeerChainEscrowAddress sdk.AccAddress) {
	coreumBankClient := banktypes.NewQueryClient(coreumChainCtx.ClientContext)
	sentFromPeerChainToCoreumCoin := sdk.NewCoin(integrationtests.ConvertToIBCDenom(peerChainToCoreumChannelID, sendCoin.Denom), sendCoin.Amount)
	coreumIssuerBalanceBeforeTransferBackRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumIssuer.String(),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	// check that escrow balance is decreased now
	coreumToPeerChainEscrowAddressBeforeTranserRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChainCtx.ConvertToBech32Address(coreumToPeerChainEscrowAddress),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)

	_, err = peerChainCtx.ExecuteIBCTransfer(ctx, peerChainRecipient, sentFromPeerChainToCoreumCoin, coreumChainCtx, coreumIssuer)
	requireT.NoError(err)

	// check new issuer balance (no commission)
	expectedCoreumIssuerBalanceAfterTransferBack := sdk.NewCoin(
		sendCoin.Denom,
		coreumIssuerBalanceBeforeTransferBackRes.Balance.Amount.Add(sentFromPeerChainToCoreumCoin.Amount),
	)
	err = coreumChainCtx.AwaitForBalance(ctx, coreumIssuer, expectedCoreumIssuerBalanceAfterTransferBack)
	requireT.NoError(err)
	// check that escrow balance is decreased now
	coreumToPeerChainEscrowAddressRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChainCtx.ConvertToBech32Address(coreumToPeerChainEscrowAddress),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(
		coreumToPeerChainEscrowAddressBeforeTranserRes.Balance.Amount.Sub(sendCoin.Amount).String(),
		coreumToPeerChainEscrowAddressRes.Balance.Amount.String(),
	)
	// check new sender balance
	coreumSenderBalanceBeforeTransferBackRes, err := coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumSender.String(),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)

	_, err = peerChainCtx.ExecuteIBCTransfer(ctx, peerChainRecipient, sentFromPeerChainToCoreumCoin, coreumChainCtx, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalanceAfterTransferBack := sdk.NewCoin(sendCoin.Denom, coreumSenderBalanceBeforeTransferBackRes.Balance.Amount.Add(sentFromPeerChainToCoreumCoin.Amount))
	err = coreumChainCtx.AwaitForBalance(ctx, coreumSender, expectedCoreumSenderBalanceAfterTransferBack)
	requireT.NoError(err)

	// check zero balance on escrow address
	coreumToPeerChainEscrowAddressRes, err = coreumBankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChainCtx.ConvertToBech32Address(coreumToPeerChainEscrowAddress),
		Denom:   sendCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.ZeroInt().String(), coreumToPeerChainEscrowAddressRes.Balance.Amount.String())
}

func getIssueFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.IssueFee
}
