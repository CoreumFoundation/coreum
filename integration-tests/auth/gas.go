package auth

import (
	"context"
	"math/big"

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
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.InitialGasPrice(),
				chain.Network.DeterministicGas().BankSend,
				1,
				big.NewInt(100),
			), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}

			// FIXME (wojtek): Temporary code for transition
			if chain.Fund != nil {
				chain.Fund(sender, initialBalance)
			}

			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			gasPrice := big.NewInt(0).Sub(chain.Network.MinDiscountedGasPrice(), big.NewInt(1))
			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{Amount: gasPrice, Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)

			// Broadcast should fail because gas price is too low for transaction to enter mempool
			_, err = coredClient.Broadcast(ctx, txBytes)
			require.True(t, client.IsInsufficientFeeError(err))
		}
}
