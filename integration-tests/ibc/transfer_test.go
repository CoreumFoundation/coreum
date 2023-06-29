//go:build integrationtests

package ibc

import (
	"context"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
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
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, expectedGaiaRecipientBalance))
	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, expectedGaiaRecipientBalance, coreumChain.Chain.ChainContext, coreumSender)
	requireT.NoError(err)

	expectedCoreumSenderBalance := sdk.NewCoin(sendToGaiaCoin.Denom, expectedGaiaRecipientBalance.Amount)
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, expectedCoreumSenderBalance))
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
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumToCoreumSender, expectedBalanceAtCoreum))

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
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaAccount2, expectedGaiaSenderBalance))
}

func TestTimedOutTransfer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 30*time.Second)
	defer retryCancel()

	err := retry.Do(retryCtx, time.Millisecond, func() error {
		coreumSender := coreumChain.GenAccount()
		gaiaRecipient := gaiaChain.GenAccount()

		sendToGaiaCoin := coreumChain.NewCoin(sdk.NewInt(1000))
		coreumChain.FundAccountsWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
			Amount:   sendToGaiaCoin.Amount,
		})

		_, err := coreumChain.ExecuteTimingOutIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), ibcchanneltypes.ErrPacketTimeout.Error()):
			return retry.Retryable(err)
		default:
			requireT.NoError(err, t)
		}

		parallelCtx, parallelCancel := context.WithCancel(ctx)
		defer parallelCancel()
		errCh := make(chan error, 1)
		go func() {
			defer parallelCancel()
			if err := coreumChain.AwaitForBalance(parallelCtx, t, coreumSender, sendToGaiaCoin); err != nil {
				select {
				case errCh <- retry.Retryable(err):
				default:
				}
			} else {
				errCh <- nil
			}
		}()
		go func() {
			if err := gaiaChain.AwaitForBalance(parallelCtx, t, gaiaRecipient, sdk.NewCoin(convertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom), sendToGaiaCoin.Amount)); err == nil {
				select {
				case errCh <- retry.Retryable(errors.New("timeout didn't happen")):
					parallelCancel()
				default:
				}
			}
		}()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}

		// funds should not be received on gaia
		bankClient := banktypes.NewQueryClient(gaiaChain.ClientContext)
		resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: gaiaChain.ConvertToBech32Address(gaiaRecipient),
			Denom:   convertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom),
		})
		requireT.NoError(err)
		requireT.Equal("0", resp.Balance.Amount.String())

		return nil
	})
	requireT.NoError(err)
}
