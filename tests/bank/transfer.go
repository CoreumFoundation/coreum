package bank

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/tests/testing"
)

// TestInitialBalance checks that initial balance is set by genesis block
func TestInitialBalance(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	// Create new random wallet
	wallet := testing.RandomWallet()

	// First returned value is the slice of objects representing prerequisites for the test
	return func(ctx context.Context) error {
			return chain.Network.FundAccount(wallet.Key.PubKey(),
				testing.MustCoin(types.NewCoin(big.NewInt(100), chain.Network.TokenSymbol())).String())
		},

		// Second returned value is the function running test
		func(ctx context.Context, t testing.T) {
			// Query for current balance available on the wallet
			balances, err := chain.Client.QueryBankBalances(ctx, wallet)
			require.NoError(t, err)

			// Test that wallet owns expected balance
			assert.Equal(t, "100", balances[chain.Network.TokenSymbol()].Amount.String())
		}
}
