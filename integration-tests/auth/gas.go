package auth

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TODO (wojtek): once we have other coins add test verifying that transaction offering fee in coin other then CORE is rejected

// TestFeeLimits verifies that invalid message gas won't be accepted.
func TestFeeLimits(ctx context.Context, t testing.T, chain testing.Chain) {
	sender, err := chain.GenFundedAccount(ctx)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	gasPrice, err := tx.GetGasPrice(ctx, chain.ClientContext)
	require.NoError(t, err)

	// the gas price is too low
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.NewDecCoin(gasPrice.Amount.QuoInt64(2)).String()),
		msg)
	require.ErrorIs(t, cosmoserrors.ErrInsufficientFee, err)

	// no gas price
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.ErrorIs(t, err, cosmoserrors.ErrInsufficientFee)

	// more gas than MaxBlockGas
	maxBlockGas := chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas

	// gas equal MaxBlockGas, the tx should pass
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas)).
			WithGasPrices(gasPrice.String()),
		msg)
	require.NoError(t, err)

	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas+1)).
			WithGasPrices(gasPrice.String()),
		msg)
	// TODO(dhil) here we get the Internal error -> "tx (***) not found" and the test takes the "txTimeout" time, validate that it's expected
	require.Error(t, err)
}
