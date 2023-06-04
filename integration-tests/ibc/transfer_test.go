//go:build integrationtests

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

func TestIBCTransferFromCoreumToGaiaAndBack(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.GetIBCChannelID(ctx, t, coreumChain.ChainSettings.ChainID)

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
	coreumChain.FundAccountsWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	})

	txRes, err := coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{}))

	expectedGaiaRecipientBalance := sdk.NewCoin(convertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)
	gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, expectedGaiaRecipientBalance)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, expectedGaiaRecipientBalance, coreumChain.Chain.ChainContext, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalance := sdk.NewCoin(sendToGaiaCoin.Denom, expectedGaiaRecipientBalance.Amount)
	coreumChain.AwaitForBalance(ctx, t, coreumSender, expectedCoreumSenderBalance)
}

// TestIBCTransferFromGaiaToCoreumAndBack checks IBC transfer in the following order:
// gaiaAccount [IBC]-> coreumToCoreumSender [bank.Send]-> coreumToGaiaSender [IBC]-> gaiaAccount.
func TestIBCTransferFromGaiaToCoreumAndBack(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumToGaiaChannelID := coreumChain.GetIBCChannelID(ctx, t, gaiaChain.ChainSettings.ChainID)
	sendToCoreumCoin := gaiaChain.NewCoin(sdk.NewInt(1000))

	// Generate accounts
	gaiaAccount := gaiaChain.GenAccount()
	coreumToCoreumSender := coreumChain.GenAccount()
	coreumToGaiaSender := coreumChain.GenAccount()

	// Fund accounts
	coreumChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: coreumToCoreumSender,
		Amount:  coreumChain.NewCoin(sdk.NewInt(1000000)),
	})
	coreumChain.FundAccountsWithOptions(ctx, t, coreumToGaiaSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaAccount,
		Amount:  sendToCoreumCoin,
	})

	// 1. Send from gaiaAccount to coreumToCoreumSender
	_, err := gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaAccount, sendToCoreumCoin, coreumChain.Chain.ChainContext, coreumToCoreumSender)
	requireT.NoError(err)

	expectedBalanceAtCoreum := sdk.NewCoin(convertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	coreumChain.AwaitForBalance(ctx, t, coreumToCoreumSender, expectedBalanceAtCoreum)

	// 2. Send from coreumToCoreumSender to coreumToGaiaSender
	sendMsg := &banktypes.MsgSend{
		FromAddress: coreumToCoreumSender.String(),
		ToAddress:   coreumToGaiaSender.String(),
		Amount:      []sdk.Coin{expectedBalanceAtCoreum},
	}
	_, err = client.BroadcastTx(
		ctx,
		coreumChain.ClientContext.WithFromAddress(coreumToCoreumSender),
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	coreumChain.AwaitForBalance(ctx, t, coreumToGaiaSender, expectedBalanceAtCoreum)

	// 3. Send from coreumToGaiaSender back to gaiaAccount
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumToGaiaSender, expectedBalanceAtCoreum, gaiaChain.ChainContext, gaiaAccount)
	requireT.NoError(err)

	expectedGaiaSenderBalance := sdk.NewCoin(sendToCoreumCoin.Denom, expectedBalanceAtCoreum.Amount)
	gaiaChain.AwaitForBalance(ctx, t, gaiaAccount, expectedGaiaSenderBalance)
}
