//go:build integrationtests

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
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

func TestIBCTransferFromGaiaToCoreumAndBack(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumToGaiaChannelID := coreumChain.GetIBCChannelID(ctx, t, gaiaChain.ChainSettings.ChainID)

	gaiaSender := gaiaChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()

	sendToCoreumCoin := gaiaChain.NewCoin(sdk.NewInt(1000))
	coreumChain.FundAccountsWithOptions(ctx, t, coreumRecipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaSender,
		Amount:  sendToCoreumCoin,
	})

	_, err := gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaSender, sendToCoreumCoin, coreumChain.Chain.ChainContext, coreumRecipient)
	requireT.NoError(err)

	expectedCoreumRecipientBalance := sdk.NewCoin(convertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	coreumChain.AwaitForBalance(ctx, t, coreumRecipient, expectedCoreumRecipientBalance)

	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumRecipient, expectedCoreumRecipientBalance, gaiaChain.ChainContext, gaiaSender)
	requireT.NoError(err)

	expectedGaiaSenderBalance := sdk.NewCoin(sendToCoreumCoin.Denom, expectedCoreumRecipientBalance.Amount)
	gaiaChain.AwaitForBalance(ctx, t, gaiaSender, expectedGaiaSenderBalance)
}
