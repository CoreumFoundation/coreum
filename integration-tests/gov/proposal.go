package gov

import (
	"context"

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
		sdk.NewInt(20000000000),
	)

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{
		Wallet: proposer,
		Amount: testing.MustNewCoin(t, proposerInitialBalance, chain.NetworkConfig.TokenSymbol),
	}))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{
		Wallet: voter1,
		Amount: testing.MustNewCoin(t, voterInitialBalance, chain.NetworkConfig.TokenSymbol),
	}))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{
		Wallet: voter2,
		Amount: testing.MustNewCoin(t, voterInitialBalance, chain.NetworkConfig.TokenSymbol),
	}))

	// Set account deposit amount
	depositAmount := testing.MustNewCoin(t, sdk.NewInt(2500000), chain.NetworkConfig.TokenSymbol)

	// Submit a param change proposal
	txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBase(t, chain, proposer),
		Proposer:       proposer,
		InitialDeposit: depositAmount,
		Content: paramproposal.NewParameterChangeProposal(
			"Change UnbondingTime", "Propose changing UnbondingTime in the staking module", []paramproposal.ParamChange{
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
	proposalID, err := coredClient.GetProposalByTx(ctx, result.TxHash)
	require.NoError(t, err)

	// Query wallets for current balance
	balancesProposer, err := coredClient.QueryBankBalances(ctx, proposer)
	require.NoError(t, err)

	// Test that tokens disappeared from proposer's wallet
	// - 10core were deposited
	// - 187500000core were taken as fee
	assert.Equal(t, "19810000000", balancesProposer[chain.NetworkConfig.TokenSymbol].Amount.String())

	logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

	// Make proposal deposits
	depositProposal(ctx, t, chain, voter1, depositAmount, proposalID)
	depositProposal(ctx, t, chain, voter2, depositAmount, proposalID)

	logger.Get(ctx).Info("2 depositors have deposited amounts successfully")

	// Vote for the proposal
	balanceVoter1 := voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposalID)
	balanceVoter2 := voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposalID)

	logger.Get(ctx).Info("2 voters have voted successfully")

	// Test that tokens disappeared from voter's wallet
	// - 187500000core were taken as fee
	assert.Equal(t, "19622500000", balanceVoter1.Amount.String())
	assert.Equal(t, "19622500000", balanceVoter2.Amount.String())
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

func buildBase(t testing.T, chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, chain.NetworkConfig.TokenSymbol),
	}
}
