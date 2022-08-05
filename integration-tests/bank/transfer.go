package bank

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
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
