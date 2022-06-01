package transfers

import (
	"context"
	"math/big"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
)

// VerifyInitialBalance checks that initial balance is set by genesis block
func VerifyInitialBalance(chain apps.Cored) (testing.PrepareFunc, testing.RunFunc) {
	var wallet cored.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Create new random wallet with predefined balance added to genesis block
			wallet, _ = chain.AddWallet("100core")
			return nil
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Query for current balance available on the wallet
			balances, err := client.QBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances["core"].Amount.String())
		}
}

// TransferCore checks that core is transferred correctly between wallets
func TransferCore(chain apps.Cored) (testing.PrepareFunc, testing.RunFunc) {
	var sender, receiver cored.Wallet

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Create two random wallets with predefined amounts of core
			sender, _ = chain.AddWallet("100core")
			receiver, _ = chain.AddWallet("10core")
			return nil
		},

		// Second function runs test
		func(ctx context.Context, t *testing.T) {
			// Wait until chain is healthy
			testing.WaitUntilHealthy(ctx, t, 20*time.Second, chain)

			// Create client so we can send transactions and query state
			client := chain.Client()

			// Transfer 10 cores from sender to receiver
			txHash, err := client.TxBankSend(ctx, sender, receiver, cored.Balance{Denom: "core", Amount: big.NewInt(10)})
			require.NoError(t, err)

			logger.Get(ctx).Info("Transfer executed", zap.String("txHash", txHash))

			// Query wallets for current balance
			balancesSender, err := client.QBankBalances(ctx, sender)
			require.NoError(t, err)

			balancesReceiver, err := client.QBankBalances(ctx, receiver)
			require.NoError(t, err)

			// Test that tokens disappeared from sender's wallet
			assert.Equal(t, "90", balancesSender["core"].Amount.String())

			// Test that tokens reached receiver's wallet
			assert.Equal(t, "20", balancesReceiver["core"].Amount.String())
		}
}
