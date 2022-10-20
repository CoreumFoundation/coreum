package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// FIXME (wojtek): add test verifying that transfer fails if sender is out of balance.

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient := chain.GenAccount()

	senderInitialAmount := sdk.NewInt(100)
	recipientInitialAmount := sdk.NewInt(10)
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			sender,
			chain.NewCoin(testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
				chain.GasLimitByMsgs(&banktypes.MsgSend{}),
				1,
				senderInitialAmount,
			)),
		),
		testing.NewFundedAccount(
			recipient,
			chain.NewCoin(recipientInitialAmount),
		),
	))

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
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	balancesRecipient, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	assert.Equal(t, senderInitialAmount.Sub(amountToSend).String(), balancesSender.Balance.Amount.String())
	assert.Equal(t, recipientInitialAmount.Add(amountToSend).String(), balancesRecipient.Balance.Amount.String())
}
