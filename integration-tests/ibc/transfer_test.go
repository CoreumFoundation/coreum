//go:build integrationtests

package ibc

import (
	"context"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
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

	sendToGaiaCoin := coreumChain.NewCoin(sdkmath.NewInt(1000))
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	})

	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdkmath.NewInt(1000000)), // coin for the fees
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
	sendToCoreumCoin := gaiaChain.NewCoin(sdkmath.NewInt(1000))

	// Generate accounts
	gaiaAccount1 := gaiaChain.GenAccount()
	gaiaAccount2 := gaiaChain.GenAccount()
	coreumToCoreumSender := coreumChain.GenAccount()
	coreumToGaiaSender := coreumChain.GenAccount()

	// Fund accounts
	coreumChain.FundAccountWithOptions(ctx, t, coreumToCoreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
	})

	coreumChain.FundAccountWithOptions(ctx, t, coreumToGaiaSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaAccount1,
		Amount:  sendToCoreumCoin.Add(gaiaChain.NewCoin(sdkmath.NewInt(1000000))), // coin to send + coin for the fee
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
	osmosisChain := chains.Osmosis

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	retryCtx, retryCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer retryCancel()

	// This is the retry loop where we try to trigger a timeout condition for IBC transfer.
	// We can't reproduce it with 100% probability, so we may need to try it many times.
	// On every trial we send funds from one chain to the other. Then we observe accounts on both chains
	// to find if IBC transfer completed successfully or timed out. If tokens were delivered to the recipient
	// we must retry. Otherwise, if tokens were returned back to the sender, we might continue the test.
	err := retry.Do(retryCtx, time.Millisecond, func() error {
		coreumSender := coreumChain.GenAccount()
		osmosisRecipient := osmosisChain.GenAccount()

		sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(1000))
		coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
			Amount:   sendToOsmosisCoin.Amount,
		})

		_, err := coreumChain.ExecuteTimingOutIBCTransfer(ctx, t, coreumSender, sendToOsmosisCoin, osmosisChain.ChainContext, osmosisRecipient)
		switch {
		case err == nil:
		case strings.Contains(err.Error(), ibcchanneltypes.ErrPacketTimeout.Error()):
			return retry.Retryable(err)
		default:
			requireT.NoError(err)
		}

		parallelCtx, parallelCancel := context.WithCancel(ctx)
		defer parallelCancel()
		errCh := make(chan error, 1)
		go func() {
			// In this goroutine we check if funds were returned back to the sender.
			// If this happens it means timeout occurred.

			defer parallelCancel()
			if err := coreumChain.AwaitForBalance(parallelCtx, t, coreumSender, sendToOsmosisCoin); err != nil {
				select {
				case errCh <- retry.Retryable(err):
				default:
				}
			} else {
				errCh <- nil
			}
		}()
		go func() {
			// In this goroutine we check if funds were delivered to the other chain.
			// If this happens it means timeout didn't occur and we must try again.

			if err := osmosisChain.AwaitForBalance(parallelCtx, t, osmosisRecipient, sdk.NewCoin(convertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom), sendToOsmosisCoin.Amount)); err == nil {
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

		// At this point we are sure that timeout happened.

		// funds should not be received on gaia
		bankClient := banktypes.NewQueryClient(osmosisChain.ClientContext)
		resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: osmosisChain.MustConvertToBech32Address(osmosisRecipient),
			Denom:   convertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		})
		requireT.NoError(err)
		requireT.Equal("0", resp.Balance.Amount.String())

		return nil
	})
	requireT.NoError(err)
}

func TestRejectedTransfer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	gaiaChain := chains.Gaia

	gaiaToCoreumChannelID := gaiaChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)

	// Bank module rejects transfers targeting some module accounts. We use this feature to test that
	// this type of IBC transfer is rejected by the receiving chain.
	moduleAddress := authtypes.NewModuleAddress(ibctransfertypes.ModuleName)
	coreumSender := coreumChain.GenAccount()
	gaiaRecipient := gaiaChain.GenAccount()

	sendToGaiaCoin := coreumChain.NewCoin(sdkmath.NewInt(1000))
	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
		Amount:   sendToGaiaCoin.Amount,
	})
	gaiaChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: gaiaRecipient,
		Amount:  gaiaChain.NewCoin(sdk.NewIntFromUint64(100000)),
	})

	_, err := coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// funds should be returned to coreum
	requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumSender, sendToGaiaCoin))

	// funds should not be received on gaia
	ibcGaiaDenom := convertToIBCDenom(gaiaToCoreumChannelID, sendToGaiaCoin.Denom)
	bankClient := banktypes.NewQueryClient(gaiaChain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: gaiaChain.MustConvertToBech32Address(moduleAddress),
		Denom:   ibcGaiaDenom,
	})
	requireT.NoError(err)
	requireT.Equal("0", resp.Balance.Amount.String())

	// test that the reverse transfer from gaia to coreum is blocked too

	coreumChain.FundAccountWithOptions(ctx, t, coreumSender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&ibctransfertypes.MsgTransfer{}},
	})

	sendToCoreumCoin := sdk.NewCoin(ibcGaiaDenom, sendToGaiaCoin.Amount)
	_, err = coreumChain.ExecuteIBCTransfer(ctx, t, coreumSender, sendToGaiaCoin, gaiaChain.ChainContext, gaiaRecipient)
	requireT.NoError(err)
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sendToCoreumCoin))

	_, err = gaiaChain.ExecuteIBCTransfer(ctx, t, gaiaRecipient, sendToCoreumCoin, coreumChain.ChainContext, moduleAddress)
	requireT.NoError(err)

	// funds should be returned to gaia
	requireT.NoError(gaiaChain.AwaitForBalance(ctx, t, gaiaRecipient, sendToCoreumCoin))

	// funds should not be received on coreum
	bankClient = banktypes.NewQueryClient(coreumChain.ClientContext)
	resp, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: coreumChain.MustConvertToBech32Address(moduleAddress),
		Denom:   sendToGaiaCoin.Denom,
	})
	requireT.NoError(err)
	requireT.Equal("0", resp.Balance.Amount.String())
}
