//go:build integrationtests

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

func TestIBCFromCoreumToGaiaAndBack(t *testing.T) {
	t.Parallel()

	channelsInfo := AwaitForIBCConfig(t)
	coreumToGaiaChannelID := channelsInfo.CoreumToGaiaChannelID
	gaiaToCoreumChannelID := channelsInfo.GaiaToCoreumChannelID

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	}))

	txRes, err := ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumSender, coreumToGaiaChannelID, sendToGaiaCoin, gaiaChain, gaiaRecipient)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{}))

	expectedGaiaRecipientBalance := sdk.NewCoin(ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)
	err = AwaitForBalance(ctx, gaiaChain, gaiaRecipient, expectedGaiaRecipientBalance)
	requireT.NoError(err)
	_, err = ExecuteIBCTransfer(ctx, gaiaChain, gaiaRecipient, gaiaToCoreumChannelID, expectedGaiaRecipientBalance, coreumChain.Chain, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalance := sdk.NewCoin(sendToGaiaCoin.Denom, expectedGaiaRecipientBalance.Amount)
	err = AwaitForBalance(ctx, coreumChain.Chain, coreumSender, expectedCoreumSenderBalance)
	requireT.NoError(err)
}

func TestIBCFromGaiaToCoreumAndBack(t *testing.T) {
	t.Parallel()

	channelsInfo := AwaitForIBCConfig(t)
	coreumToGaiaChannelID := channelsInfo.CoreumToGaiaChannelID
	gaiaToCoreumChannelID := channelsInfo.GaiaToCoreumChannelID

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

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

	_, err := ExecuteIBCTransfer(ctx, gaiaChain, gaiaSender, gaiaToCoreumChannelID, sendToCoreumCoin, coreumChain.Chain, coreumRecipient)
	requireT.NoError(err)

	expectedCoreumRecipientBalance := sdk.NewCoin(ConvertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	err = AwaitForBalance(ctx, coreumChain.Chain, coreumRecipient, expectedCoreumRecipientBalance)
	requireT.NoError(err)

	_, err = ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumRecipient, coreumToGaiaChannelID, expectedCoreumRecipientBalance, gaiaChain, gaiaSender)
	requireT.NoError(err)

	expectedGaiaSenderBalance := sdk.NewCoin(sendToCoreumCoin.Denom, expectedCoreumRecipientBalance.Amount)
	err = AwaitForBalance(ctx, gaiaChain, gaiaSender, expectedGaiaSenderBalance)
	requireT.NoError(err)
}
