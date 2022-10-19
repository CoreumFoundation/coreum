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
	sender := chain.GenAccount()

	maxBlockGas := chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&banktypes.MsgSend{}),
		1,
		sdk.NewInt(maxBlockGas+100),
	)))))

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	gasPriceWithMaxDiscount := chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice.
		Mul(sdk.OneDec().Sub(chain.NetworkConfig.Fee.FeeModel.Params().MaxDiscount))

	// the gas price is too low
	_, err := tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.NewDecCoin(gasPriceWithMaxDiscount.QuoInt64(2)).String()),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// no gas price
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// more gas than MaxBlockGas
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas+1)),
		msg)
	// TODO(dhil) here we get the Internal error -> "tx (***) not found" and the test takes the "txTimeout" time, validate that it's expected
	require.Error(t, err)

	// gas equal MaxBlockGas, the tx should pass
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas)),
		msg)
	require.NoError(t, err)
}
