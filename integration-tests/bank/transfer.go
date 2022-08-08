package bank

import (
	"context"
	"math/big"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestInitialBalance checks that initial balance is set by genesis block
func TestInitialBalance(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	// Create new random wallet
	wallet := testing.RandomWallet()

	// First returned function prepares initial well-known state
	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(big.NewInt(100), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(wallet.Key.PubKey(), initialBalance.String())
		},

		// Second returned function runs test
		func(ctx context.Context, t testing.T) {
			// Query for current balance available on the wallet
			balances, err := chain.Client.QueryBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances[chain.Network.TokenSymbol()].Amount.String())
		}
}

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	// Create two random wallets
	sender := testing.RandomWallet()
	receiver := testing.RandomWallet()

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Fund wallets
			senderInitialBalance, err := types.NewCoin(big.NewInt(180000100), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}

			receiverInitialBalance, err := types.NewCoin(big.NewInt(10), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}

			if err := chain.Network.FundAccount(sender.Key.PubKey(), senderInitialBalance.String()); err != nil {
				return err
			}
			return chain.Network.FundAccount(receiver.Key.PubKey(), receiverInitialBalance.String())
		},

		// Second function runs test
		func(ctx context.Context, t testing.T) {
			// Create client so we can send transactions and query state
			coredClient := chain.Client

			// Transfer 10 cores from sender to receiver
			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{Amount: chain.Network.InitialGasPrice(), Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: receiver,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)
			result, err := coredClient.Broadcast(ctx, txBytes)
			require.NoError(t, err)

			logger.Get(ctx).Info("Transfer executed", zap.String("txHash", result.TxHash))

			// Query wallets for current balance
			balancesSender, err := coredClient.QueryBankBalances(ctx, sender)
			require.NoError(t, err)

			balancesReceiver, err := coredClient.QueryBankBalances(ctx, receiver)
			require.NoError(t, err)

			// Test that tokens disappeared from sender's wallet
			// - 10core were transferred to receiver
			// - 180000000core were taken as fee
			assert.Equal(t, "90", balancesSender[chain.Network.TokenSymbol()].Amount.String())

			// Test that tokens reached receiver's wallet
			assert.Equal(t, "20", balancesReceiver[chain.Network.TokenSymbol()].Amount.String())
		}
}
