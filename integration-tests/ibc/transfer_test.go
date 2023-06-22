//go:build integrationtests

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
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

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
	coreumChain.FundAccountsWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	})

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdk.NewInt(1000000)), // coin for the fees
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
// gaiaAccount1 [IBC]-> coreumToCoreumSender [bank.Send]-> coreumToGaiaSender [IBC]-> gaiaAccount2.
func TestIBCTransferFromGaiaToCoreumAndBack(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)
	sendToCoreumCoin := gaiaChain.NewCoin(sdk.NewInt(1000))

	// Generate accounts
	gaiaAccount1 := gaiaChain.GenAccount()
	gaiaAccount2 := gaiaChain.GenAccount()
	coreumToCoreumSender := coreumChain.GenAccount()
	coreumToGaiaSender := coreumChain.GenAccount()

	// Fund accounts
	coreumChain.FundAccountsWithOptions(ctx, t, coreumToCoreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
	})

	coreumChain.FundAccountsWithOptions(ctx, t, coreumToGaiaSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaAccount1,
		Amount:  sendToCoreumCoin.Add(gaiaChain.NewCoin(sdk.NewInt(1000000))), // coin to send + coin for the fee
	})

	// Send from gaiaAccount to coreumToCoreumSender
	_, err := gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaAccount1, sendToCoreumCoin, coreumChain.Chain.ChainContext, coreumToCoreumSender)
	requireT.NoError(err)

	expectedBalanceAtCoreum := sdk.NewCoin(convertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	coreumChain.AwaitForBalance(ctx, t, coreumToCoreumSender, expectedBalanceAtCoreum)

	// Send from coreumToCoreumSender to coreumToGaiaSender
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

	bankClient := banktypes.NewQueryClient(coreumChain.ClientContext)
	queryBalanceResponse, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumToGaiaSender.String(),
		Denom:   expectedBalanceAtCoreum.Denom,
	})
	requireT.NoError(err)
	assert.Equal(t, expectedBalanceAtCoreum.Amount.String(), queryBalanceResponse.Balance.Amount.String())

	// Send from coreumToGaiaSender back to gaiaAccount
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumToGaiaSender, expectedBalanceAtCoreum, gaiaChain.ChainContext, gaiaAccount2)
	requireT.NoError(err)
	expectedGaiaSenderBalance := sdk.NewCoin(sendToCoreumCoin.Denom, expectedBalanceAtCoreum.Amount)
	gaiaChain.AwaitForBalance(ctx, t, gaiaAccount2, expectedGaiaSenderBalance)
}
