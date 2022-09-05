package gov

import (
	"context"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestProposalParamChange checks that param change proposal works correctly
func TestProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create client so we can send transactions and query state
	coredClient := chain.Client

	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()

	// Prepare initial balances
	proposerInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(20000000000),
	)
	voterInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(835000000000000), // 1670000000000000
	)

	// Make sure the total balance of two voters is > 66% of the total supply.
	// This is needed to have successful voting process.
	totalSupply, err := coredClient.GetTotalSupply(ctx, chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)
	minTotalBalance := totalSupply.MulRaw(66).ModRaw(100)
	totalVotersBalance := voterInitialBalance.MulRaw(2)
	require.True(t, minTotalBalance.LT(totalVotersBalance), "%s > %s", minTotalBalance, totalVotersBalance)

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{
		Wallet: proposer,
		Amount: testing.MustNewCoin(t, proposerInitialBalance, chain.NetworkConfig.TokenSymbol),
	}, testing.FundedAccount{
		Wallet: voter1,
		Amount: testing.MustNewCoin(t, voterInitialBalance, chain.NetworkConfig.TokenSymbol),
	}, testing.FundedAccount{
		Wallet: voter2,
		Amount: testing.MustNewCoin(t, voterInitialBalance, chain.NetworkConfig.TokenSymbol),
	}))
	t.Cleanup(func() {
		assert.NoError(t, chain.Faucet.CleanupAccounts(ctx, proposer, voter1, voter2))
	})

	// Set account deposit amount
	depositAmount := testing.MustNewCoin(t, sdk.NewInt(3500000), chain.NetworkConfig.TokenSymbol)

	// Submit a param change proposal
	txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBase(t, chain, proposer),
		Proposer:       proposer,
		InitialDeposit: depositAmount,
		Content: paramproposal.NewParameterChangeProposal(
			"Change UnbondingTime",
			"Propose changing UnbondingTime in the staking module",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange("staking", "UnbondingTime", "\"172800000000000\""),
			},
		),
	})
	require.NoError(t, err)
	result, err := coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
	proposal, err := coredClient.GetProposalByTx(ctx, result.TxHash)
	require.NoError(t, err)

	// Check proposer balance
	balancesProposer, err := coredClient.QueryBankBalances(ctx, proposer)
	require.NoError(t, err)
	assert.Equal(t, "19996500000", balancesProposer[chain.NetworkConfig.TokenSymbol].Amount.String())

	logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

	// Make proposal deposits
	depositProposal(ctx, t, chain, voter1, depositAmount, proposal.ProposalId)
	depositProposal(ctx, t, chain, voter2, depositAmount, proposal.ProposalId)

	logger.Get(ctx).Info("2 depositors have deposited amounts successfully")

	// Wait for voting period to be started
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusVotingPeriod, testing.MinDepositPeriod, proposal.ProposalId)

	// Vote for the proposal
	balanceVoter1 := voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposal.ProposalId)
	balanceVoter2 := voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposal.ProposalId)

	// Test that tokens disappeared from voter's wallet
	// - 187500000core were taken as fee
	assert.Equal(t, "834999809000000", balanceVoter1.Amount.String())
	assert.Equal(t, "834999809000000", balanceVoter2.Amount.String())

	logger.Get(ctx).Info("2 voters have voted successfully")

	// Wait for proposal result
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusPassed, testing.MinVotingPeriod, proposal.ProposalId)
}

func depositProposal(ctx context.Context, t testing.T, chain testing.Chain, depositor types.Wallet, amount types.Coin, proposalID uint64) {
	coredClient := chain.Client
	txBytes, err := coredClient.PrepareTxSubmitProposalDeposit(ctx, client.TxSubmitProposalDepositInput{
		Base:       buildBase(t, chain, depositor),
		Depositor:  depositor,
		ProposalID: proposalID,
		Amount:     amount,
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}

func voteProposal(ctx context.Context, t testing.T, chain testing.Chain, voter types.Wallet, option govtypes.VoteOption, proposalID uint64) types.Coin {
	coredClient := chain.Client
	txBytes, err := coredClient.PrepareTxSubmitProposalVote(ctx, client.TxSubmitProposalVoteInput{
		Base:       buildBase(t, chain, voter),
		Voter:      voter,
		ProposalID: proposalID,
		Option:     option,
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	// Check vote
	votes, err := coredClient.QueryProposalVotes(ctx, proposalID)
	require.NoError(t, err)
	voterVotes, ok := votes[voter.Key.Address()]
	require.True(t, ok)
	require.Len(t, voterVotes, 1)
	require.Equal(t, voterVotes[0].Option, govtypes.OptionYes)
	require.Equal(t, voterVotes[0].Weight, sdk.NewDec(1))

	// Query wallets for current balance
	balances, err := coredClient.QueryBankBalances(ctx, voter)
	require.NoError(t, err)

	return balances[chain.NetworkConfig.TokenSymbol]
}

func waitForProposalStatus(ctx context.Context, t testing.T, chain testing.Chain, status govtypes.ProposalStatus, duration time.Duration, proposalID uint64) *govtypes.Proposal {
	coredClient := chain.Client
	timeout := time.NewTimer(duration)
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		select {
		case <-timeout.C:
			t.Errorf("timed out proposal %d", proposalID)
			t.FailNow()
		default:
			proposal, err := coredClient.GetProposal(ctx, proposalID)
			require.NoError(t, err)

			if proposal.Status == status {
				return proposal
			}
		}
	}
	t.Errorf("timed out proposal %d", proposalID)
	t.FailNow()
	return nil
}

func buildBase(t testing.T, chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, chain.NetworkConfig.TokenSymbol),
	}
}
