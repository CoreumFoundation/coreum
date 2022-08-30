package gov

import (
	"context"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"math/big"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestProposalParamChange checks that param change proposal works correctly
func TestProposalParamChange(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()
	voter3 := testing.RandomWallet()

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Fund wallets
			initialBalance, err := types.NewCoin(big.NewInt(1000), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}

			// FIXME (wojtek): Temporary code for transition
			if chain.Fund != nil {
				chain.Fund(proposer, initialBalance)
				chain.Fund(voter1, initialBalance)
				chain.Fund(voter2, initialBalance)
				chain.Fund(voter3, initialBalance)
			}

			if err = chain.Network.FundAccount(proposer.Key.PubKey(), initialBalance.String()); err != nil {
				return err
			}

			if err = chain.Network.FundAccount(voter1.Key.PubKey(), initialBalance.String()); err != nil {
				return err
			}

			if err = chain.Network.FundAccount(voter2.Key.PubKey(), initialBalance.String()); err != nil {
				return err
			}

			if err = chain.Network.FundAccount(voter3.Key.PubKey(), initialBalance.String()); err != nil {
				return err
			}

			return nil
		},

		// Second function runs test
		func(ctx context.Context, t testing.T) {
			// Create client so we can send transactions and query state
			coredClient := chain.Client

			// Submit a param change proposal
			txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
				Base: tx.BaseInput{
					Signer:   proposer,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{Amount: chain.Network.FeeModel().InitialGasPrice.BigInt(), Denom: chain.Network.TokenSymbol()},
				},
				Proposer:       proposer,
				InitialDeposit: types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
				Content: paramproposal.NewParameterChangeProposal(
					"test", "test", []paramproposal.ParamChange{
						{
							Subspace: "staking",
							Key:      "UnbondingTime",
							Value:    "172800000000000",
						},
					},
				),
			})
			require.NoError(t, err)
			result, err := coredClient.Broadcast(ctx, txBytes)
			require.NoError(t, err)

			logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

			// Query wallets for current balance
			balancesProposer, err := coredClient.QueryBankBalances(ctx, proposer)
			require.NoError(t, err)

			// balancesReceiver, err := coredClient.QueryBankBalances(ctx, receiver)
			// require.NoError(t, err)

			// Test that tokens disappeared from sender's wallet
			// - 10core were transferred to receiver
			// - 180000000core were taken as fee
			assert.Equal(t, "900", balancesProposer[chain.Network.TokenSymbol()].Amount.String())

			// Test that tokens reached receiver's wallet
			// assert.Equal(t, "20", balancesReceiver[chain.Network.TokenSymbol()].Amount.String())
		}
}
