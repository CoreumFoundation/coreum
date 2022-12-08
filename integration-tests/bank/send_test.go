//go:build integration

// FIXME(dhil) here we set the profile integration since we don't run the tests by default
package bank

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	integrationtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// FIXME (wojtek): add test verifying that transfer fails if sender is out of balance.

// TestCoreSend checks that core is transferred correctly between wallets
func TestCoreSend(t *testing.T) {
	t.Parallel()

	// FIXME(dhil) Additionally we can rename integration-tests -> integration-tests
	// and move integration-tests/testing to integration-tests root  to simplify the imports
	ctx, chain := integrationtests.NewTestingContext(t)
	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	senderInitialAmount := sdk.NewInt(100)
	recipientInitialAmount := sdk.NewInt(10)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtesting.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   senderInitialAmount,
	}))
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, recipient, integrationtesting.BalancesOptions{
		Amount: recipientInitialAmount,
	}))

	// transfer tokens from sender to recipient
	amountToSend := sdk.NewInt(10)
	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", result.TxHash))

	// Query wallets for current balance
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	balancesSender, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	require.NoError(t, err)

	balancesRecipient, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	require.NoError(t, err)

	assert.Equal(t, senderInitialAmount.Sub(amountToSend).String(), balancesSender.Balance.Amount.String())
	assert.Equal(t, recipientInitialAmount.Add(amountToSend).String(), balancesRecipient.Balance.Amount.String())
}
