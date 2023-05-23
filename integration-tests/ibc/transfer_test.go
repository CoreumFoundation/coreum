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

	ctx, chains := integrationtests.NewChainsTestingContext(t, false)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID, err := gaiaChain.GetIBCChannelID(ctx, coreumChain.ChainSettings.ChainID)
	requireT.NoError(err)

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	}))

	txRes, err := coreumChain.ExecuteIBCTransfer(ctx, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{}))

	expectedGaiaRecipientBalance := sdk.NewCoin(convertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)
	err = gaiaChain.AwaitForBalance(ctx, gaiaRecipient, expectedGaiaRecipientBalance)
	requireT.NoError(err)
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, gaiaRecipient, expectedGaiaRecipientBalance, coreumChain.Chain.ChainContext, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalance := sdk.NewCoin(sendToGaiaCoin.Denom, expectedGaiaRecipientBalance.Amount)
	err = coreumChain.AwaitForBalance(ctx, coreumSender, expectedCoreumSenderBalance)
	requireT.NoError(err)
}

func TestIBCTransferFromGaiaToCoreumAndBack(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t, false)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumToGaiaChannelID, err := coreumChain.GetIBCChannelID(ctx, gaiaChain.ChainSettings.ChainID)
	requireT.NoError(err)

	gaiaSender := gaiaChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()

	sendToCoreumCoin := gaiaChain.NewCoin(sdk.NewInt(1000))
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumRecipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	}))

	requireT.NoError(gaiaChain.Faucet.FundAccounts(ctx, integrationtests.FundedAccount{
		Address: gaiaSender,
		Amount:  sendToCoreumCoin,
	}))

	_, err = gaiaChain.ExecuteIBCTransfer(ctx, gaiaSender, sendToCoreumCoin, coreumChain.Chain.ChainContext, coreumRecipient)
	requireT.NoError(err)

	expectedCoreumRecipientBalance := sdk.NewCoin(convertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	err = coreumChain.AwaitForBalance(ctx, coreumRecipient, expectedCoreumRecipientBalance)
	requireT.NoError(err)

	_, err = coreumChain.ExecuteIBCTransfer(ctx, coreumRecipient, expectedCoreumRecipientBalance, gaiaChain.ChainContext, gaiaSender)
	requireT.NoError(err)

	expectedGaiaSenderBalance := sdk.NewCoin(sendToCoreumCoin.Denom, expectedCoreumRecipientBalance.Amount)
	err = gaiaChain.AwaitForBalance(ctx, gaiaSender, expectedGaiaSenderBalance)
	requireT.NoError(err)
}
