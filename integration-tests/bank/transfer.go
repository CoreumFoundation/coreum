package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(ctx context.Context, t testing.T, chain testing.Chain) {
	sender, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)
	recipient, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	senderInitialBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	recipientInitialBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	gasPrice, err := tx.GetGasPrice(ctx, chain.ClientContext)
	require.NoError(t, err)

	// try to send the x2 balance
	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(senderInitialBalance.Balance.Add(*senderInitialBalance.Balance)),
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(gasPrice.String()),
		msg,
	)
	require.ErrorIs(t, cosmoserrors.ErrInsufficientFunds, err)

	// update the sender balance since some tokens are spend on prev reverted tx
	senderInitialBalance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	// transfer tokens from sender to recipient
	amountToSend := sdk.NewInt(10)
	msg = &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	gasPrice, err = tx.GetGasPrice(ctx, chain.ClientContext)
	require.NoError(t, err)
	txResult, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(gasPrice.String()),
		msg,
	)
	require.NoError(t, err)
	spentOnTxs := testing.ComputeFeeAmount(gasPrice.Amount, uint64(txResult.GasUsed))

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", txResult.TxHash))

	senderBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: sender.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	recipientBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.NetworkConfig.BaseDenom,
	})
	require.NoError(t, err)

	require.Equal(t, senderInitialBalance.Balance.Amount.Sub(amountToSend).Sub(spentOnTxs).String(), senderBalance.Balance.Amount.String())
	require.Equal(t, recipientInitialBalance.Balance.Amount.Add(amountToSend).String(), recipientBalance.Balance.Amount.String())
}
