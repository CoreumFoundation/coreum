package auth

import (
	"context"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestTooLowGasPrice verifies that transaction does not enter mempool if offered gas price is below minimum level
// specified by the fee model of the network
func TestTooLowGasPrice(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin2(types.NewInt(180000100), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			amount, err := types.NewCoin2(types.NewInt(10), chain.Network.TokenSymbol())
			require.NoError(t, err)

			gasPrice := chain.Network.MinDiscountedGasPrice().Sub(types.NewInt(1))
			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{Amount: gasPrice, Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   amount,
			})
			require.NoError(t, err)

			// Broadcast should fail because gas price is too low for transaction to enter mempool
			_, err = coredClient.Broadcast(ctx, txBytes)
			require.True(t, client.IsInsufficientFeeError(err))
		}
}
