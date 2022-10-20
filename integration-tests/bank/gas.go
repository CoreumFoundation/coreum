package bank

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestTransferDeterministicGas checks that transfer takes the deterministic amount of gas.
func TestTransferDeterministicGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender, err := chain.GenFundedAccount(ctx)
	recipient := chain.GenAccount()
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)

	gasPrice, err := tx.GetGasPrice(ctx, clientCtx)
	require.NoError(t, err)

	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})

	// send with not enough gas
	_, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas-1). // gas less than expected
			WithGasPrices(gasPrice.String()),
		msg)
	require.ErrorIs(t, cosmoserrors.ErrOutOfGas, err)

	// send with enough gas
	res, err := tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankSendGas).
			WithGasPrices(gasPrice.String()),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankSendGas, uint64(res.GasUsed))
}

// TestTransferDeterministicGasTwoBankSends checks that transfer takes the deterministic amount of gas.
func TestTransferDeterministicGasTwoBankSends(ctx context.Context, t testing.T, chain testing.Chain) {
	gasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})

	sender, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)
	receiver1 := chain.GenAccount()
	receiver2 := chain.GenAccount()

	bankSend1 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   receiver1.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}
	bankSend2 := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   receiver2.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1000))),
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)

	gasPrice, err := tx.GetGasPrice(ctx, clientCtx)
	require.NoError(t, err)

	txf := chain.TxFactory().
		WithGas(gasExpected).
		WithGasPrices(gasPrice.String())
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, bankSend1, bankSend2)
	require.NoError(t, err)
	require.EqualValues(t, gasExpected, result.GasUsed)
}

// TestTransferGasEstimation checks that gas is correctly estimated for send message.
func TestTransferGasEstimation(ctx context.Context, t testing.T, chain testing.Chain) {
	sender, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)

	amountToSend := sdk.NewInt(1000)
	bankSendGas := chain.GasLimitByMsgs(&banktypes.MsgSend{})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	gasPrice, err := tx.GetGasPrice(ctx, chain.ClientContext)
	require.NoError(t, err)

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	_, estimatedGas, err := tx.CalculateGas(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithGas(bankSendGas).
			WithGasPrices(gasPrice.String()),
		msg)
	require.NoError(t, err)
	assert.Equal(t, bankSendGas, estimatedGas)
}
