//go:build integrationtests

package ibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
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
	coreumRecipient := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	}))

	t.Logf("Sending from coreum to gaia")
	txRes, err := ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumSender, coreumToGaiaChannelID, sendToGaiaCoin, gaiaChain, gaiaRecipient)
	requireT.NoError(err)
	requireT.EqualValues(txRes.GasUsed, coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{}))

	t.Logf("Waiting for balance on gaia")
	gaiaRecipientBalance, err := QueryNonZeroIBCBalance(ctx, gaiaChain, gaiaRecipient, ConvertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom))
	requireT.NoError(err)
	assert.EqualValues(t, sendToGaiaCoin.Amount.String(), gaiaRecipientBalance.Amount.String())
	t.Logf("Reveiced %s on gaia", gaiaRecipientBalance.String())

	t.Logf("Sending %s back from gaia to coreum", gaiaRecipientBalance.String())
	_, err = ExecuteIBCTransfer(ctx, gaiaChain, gaiaRecipient, gaiaToCoreumChannelID, gaiaRecipientBalance, coreumChain.Chain, coreumRecipient)
	requireT.NoError(err)

	t.Logf("Waiting for balance on coreum")
	coreumRecipientBalance, err := QueryNonZeroIBCBalance(ctx, coreumChain.Chain, coreumRecipient, sendToGaiaCoin.Denom)
	requireT.NoError(err)
	assert.EqualValues(t, gaiaRecipientBalance.Amount.String(), coreumRecipientBalance.Amount.String())
	t.Logf("Reveiced %s on coreum", coreumRecipientBalance.String())
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
	gaiaRecipient := gaiaChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()

	sendToCoreumCoin := gaiaChain.NewCoin(sdk.NewInt(1000))
	requireT.NoError(coreumChain.FundAccountsWithOptions(ctx, coreumRecipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	}))

	requireT.NoError(gaiaChain.Faucet.FundAccounts(ctx, integrationtests.FundedAccount{
		Address: gaiaSender,
		Amount:  sendToCoreumCoin,
	}))

	t.Logf("Sending from gaia to coreum")
	_, err := ExecuteIBCTransfer(ctx, gaiaChain, gaiaSender, gaiaToCoreumChannelID, sendToCoreumCoin, coreumChain.Chain, coreumRecipient)
	requireT.NoError(err)

	t.Logf("Waiting for balance on coreum")
	coreumRecipientBalance, err := QueryNonZeroIBCBalance(ctx, coreumChain.Chain, coreumRecipient, ConvertToIBCDenom(coreumToGaiaChannelID, sendToCoreumCoin.Denom))
	requireT.NoError(err)
	assert.EqualValues(t, sendToCoreumCoin.Amount.String(), coreumRecipientBalance.Amount.String())
	t.Logf("Reveiced %s on coreum", coreumRecipientBalance.String())

	t.Logf("Sending %s back from coreum to gaia", coreumRecipientBalance.String())
	_, err = ExecuteIBCTransfer(ctx, coreumChain.Chain, coreumRecipient, coreumToGaiaChannelID, coreumRecipientBalance, gaiaChain, gaiaRecipient)
	requireT.NoError(err)

	t.Logf("Waiting for balance on gaia")
	gaiaRecipientBalance, err := QueryNonZeroIBCBalance(ctx, gaiaChain, gaiaRecipient, sendToCoreumCoin.Denom)
	requireT.NoError(err)
	assert.EqualValues(t, gaiaRecipientBalance.Amount.String(), coreumRecipientBalance.Amount.String())
	t.Logf("Reveiced %s on gaia", coreumRecipientBalance.String())
}
