package bank

import (
	"context"
	"math/big"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
)

// TestInitialBalance checks that initial balance is set by genesis block
func TestInitialBalance(chain cored.Cored) (testing.PrepareFunc, testing.RunFunc) {
	var wallet cored.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Create new random wallet with predefined balance added to genesis block
			wallet = chain.AddWallet("100core")
			return nil
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Query for current balance available on the wallet
			balances, err := client.QueryBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances["core"].Amount.String())
		}
}

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(chain cored.Cored) (testing.PrepareFunc, testing.RunFunc) {
	var sender, receiver cored.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Create two random wallets with predefined amounts of core
			sender = chain.AddWallet("100core")
			receiver = chain.AddWallet("10core")
			return nil
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Transfer 10 cores from sender to receiver
			txBytes, err := client.PrepareTxBankSend(ctx, sender, receiver, cored.Balance{Denom: "core", Amount: big.NewInt(10)})
			require.NoError(t, err)
			txHash, err := client.Broadcast(ctx, txBytes)
			require.NoError(t, err)

			logger.Get(ctx).Info("Transfer executed", zap.String("txHash", txHash))

			// Query wallets for current balance
			balancesSender, err := client.QueryBankBalances(ctx, sender)
			require.NoError(t, err)

			balancesReceiver, err := client.QueryBankBalances(ctx, receiver)
			require.NoError(t, err)

			// Test that tokens disappeared from sender's wallet
			assert.Equal(t, "90", balancesSender["core"].Amount.String())

			// Test that tokens reached receiver's wallet
			assert.Equal(t, "20", balancesReceiver["core"].Amount.String())
		}
}
