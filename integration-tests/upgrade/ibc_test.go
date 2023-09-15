//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	integrationtestsibc "github.com/CoreumFoundation/coreum/v3/integration-tests/ibc"
)

type ibcUpgradeTest struct {
	coreumSender, gaiaSender sdk.AccAddress

	sentToGaiaCoin, sentToCoreumCoin sdk.Coin
}

func (iut *ibcUpgradeTest) Before(t *testing.T) {
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumToGaiaChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, gaiaChain.ChainSettings.ChainID)
	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	coreumSender := coreumChain.GenAccount()    // account to send IBC transfer from Coreum to Gaia
	coreumRecipient := coreumChain.GenAccount() // account to receive IBC transfer form Gaia to Coreum

	gaiaSender := gaiaChain.GenAccount()    // account to send IBC transfer from Gaia to Coreum
	gaiaRecipient := gaiaChain.GenAccount() // account to receive IBC transfer form Coreum to Gaia

	sendToGaiaCoin := coreumChain.NewCoin(sdkmath.NewInt(1000))
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	})
	coreumChain.FundAccountWithOptions(ctx, t, coreumRecipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})

	sendToCoreumCoin := gaiaChain.NewCoin(sdkmath.NewInt(1500))
	gaiaFeesCoin := gaiaChain.NewCoin(sdkmath.NewInt(1000000))
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaSender,
		Amount:  sendToCoreumCoin.Add(gaiaFeesCoin), // amount to send + coin for the fees
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaFeesCoin, // coin for the fees
	})

	// IBC transfer from Coreum to Gaia.
	_, err := coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)

	expectedGaiaRecipientBalance := sdk.NewCoin(integrationtestsibc.ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, expectedGaiaRecipientBalance))

	// IBC transfer from Gaia to Coreum.
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaSender, sendToCoreumCoin, coreumChain.Chain.ChainContext, coreumRecipient)
	requireT.NoError(err)

	expectedCoreumRecipientBalance := sdk.NewCoin(integrationtestsibc.ConvertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom), sendToCoreumCoin.Amount)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, expectedCoreumRecipientBalance))

	iut.sentToGaiaCoin = expectedGaiaRecipientBalance
	iut.sentToCoreumCoin = expectedCoreumRecipientBalance

	// Since we will be returning funds back from destination chain to original, recipients become senders
	iut.coreumSender = coreumRecipient
	iut.gaiaSender = gaiaRecipient
}

func (iut *ibcUpgradeTest) After(t *testing.T) {
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)

	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	coreumRecipient := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	_, err := coreumChain.ExecuteIBCTransfer(ctx, t, iut.coreumSender, iut.sentToCoreumCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, gaiaChain.NewCoin(iut.sentToCoreumCoin.Amount)))

	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, iut.gaiaSender, iut.sentToGaiaCoin, coreumChain.Chain.ChainContext, coreumRecipient)
	requireT.NoError(err)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, coreumChain.NewCoin(iut.sentToGaiaCoin.Amount)))
}
