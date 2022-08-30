package gov

import (
	"context"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

	fundWallet := func(wallet types.Wallet, balance *big.Int) {
		initialBalance, err := types.NewCoin(big.NewInt(20000000000), chain.Network.TokenSymbol())
		must.OK(err)

		if chain.Fund != nil {
			chain.Fund(proposer, initialBalance)
		}

		err = chain.Network.FundAccount(proposer.Key.PubKey(), initialBalance.String())
		must.OK(err)
	}

	// First function prepares initial well-known state
	return func(ctx context.Context) error {
			// Fund wallets
			fundWallet(proposer, big.NewInt(20000000000))
			fundWallet(voter1, big.NewInt(20000000000))
			fundWallet(voter2, big.NewInt(20000000000))
			fundWallet(voter3, big.NewInt(20000000000))
			return nil
		},

		// Second function runs test
		func(ctx context.Context, t testing.T) {
			// Create client so we can send transactions and query state
			coredClient := chain.Client

			vote := func(voter types.Wallet, option govtypes.VoteOption, id uint64) (string, types.Coin) {
				txBytes, err := coredClient.PrepareTxSubmitProposalVote(ctx, client.TxSubmitProposalVoteInput{
					Base: tx.BaseInput{
						Signer:   voter,
						GasLimit: chain.Network.DeterministicGas().BankSend,
						GasPrice: types.Coin{
							Amount: chain.Network.FeeModel().InitialGasPrice.BigInt(),
							Denom:  chain.Network.TokenSymbol(),
						},
					},
					Voter:      voter,
					ProposalID: id,
					Option:     option,
				})
				require.NoError(t, err)
				result, err := coredClient.Broadcast(ctx, txBytes)
				require.NoError(t, err)

				// Query wallets for current balance
				balancesProposer, err := coredClient.QueryBankBalances(ctx, voter)
				require.NoError(t, err)

				return result.TxHash, balancesProposer[chain.Network.TokenSymbol()]
			}

			// Submit a param change proposal
			txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
				Base: tx.BaseInput{
					Signer:   proposer,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{
						Amount: chain.Network.FeeModel().InitialGasPrice.BigInt(),
						Denom:  chain.Network.TokenSymbol(),
					},
				},
				Proposer:       proposer,
				InitialDeposit: types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
				Content: paramproposal.NewParameterChangeProposal(
					"test", "test", []paramproposal.ParamChange{
						{
							Subspace: "staking",
							Key:      "UnbondingTime",
							Value:    `"172800000000000"`,
						},
					},
				),
			})
			require.NoError(t, err)
			result, err := coredClient.Broadcast(ctx, txBytes)
			require.NoError(t, err)

			logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

			// Vote for the proposal
			proposalID := uint64(1) // TODO: Fetch proposal ID from the transaction
			_, balanceVoter1 := vote(voter1, govtypes.OptionYes, proposalID)
			_, balanceVoter2 := vote(voter2, govtypes.OptionYes, proposalID)
			_, balanceVoter3 := vote(voter3, govtypes.OptionYes, proposalID)

			logger.Get(ctx).Info("Proposal has 3 votes")

			// Query wallets for current balance
			balancesProposer, err := coredClient.QueryBankBalances(ctx, proposer)
			require.NoError(t, err)

			// Test that tokens disappeared from proposer's wallet
			// - 10core were deposited
			// - 187500000core were taken as fee
			assert.Equal(t, "19812499990", balancesProposer[chain.Network.TokenSymbol()].Amount.String())

			// Test that tokens disappeared from voter's wallet
			// - 187500000core were taken as fee
			assert.Equal(t, "19812500000", balanceVoter1.Amount.String())
			assert.Equal(t, "19812500000", balanceVoter2.Amount.String())
			assert.Equal(t, "19812500000", balanceVoter3.Amount.String())
		}
}
