package gov

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
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
func TestProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create client so we can send transactions and query state
	coredClient := chain.Client

	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()
	voter3 := testing.RandomWallet()

	// Fund wallets
	initialBalance, err := types.NewCoin(big.NewInt(20000000000), chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{Wallet: proposer, Amount: initialBalance}))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{Wallet: voter1, Amount: initialBalance}))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{Wallet: voter2, Amount: initialBalance}))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.FundedAccount{Wallet: voter3, Amount: initialBalance}))

	// Set account deposit amount
	depositAmount := types.Coin{Denom: chain.NetworkConfig.TokenSymbol, Amount: big.NewInt(2500000)}

	// Submit a param change proposal
	txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBase(chain, proposer),
		Proposer:       proposer,
		InitialDeposit: depositAmount,
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
	proposalID, err := coredClient.GetProposalByTx(ctx, result.TxHash)
	require.NoError(t, err)

	logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

	// Vote for the proposal
	depositProposal(ctx, t, chain, voter1, depositAmount, proposalID)
	depositProposal(ctx, t, chain, voter2, depositAmount, proposalID)
	depositProposal(ctx, t, chain, voter3, depositAmount, proposalID)

	logger.Get(ctx).Info("3 depositors have deposited amounts successfully")

	// Vote for the proposal
	balanceVoter1 := voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposalID)
	balanceVoter2 := voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposalID)
	balanceVoter3 := voteProposal(ctx, t, chain, voter3, govtypes.OptionYes, proposalID)

	logger.Get(ctx).Info("3 voters have voted successfully")

	// Query wallets for current balance
	balancesProposer, err := coredClient.QueryBankBalances(ctx, proposer)
	require.NoError(t, err)

	// Test that tokens disappeared from proposer's wallet
	// - 10core were deposited
	// - 187500000core were taken as fee
	assert.Equal(t, "19810000000", balancesProposer[chain.NetworkConfig.TokenSymbol].Amount.String())

	// Test that tokens disappeared from voter's wallet
	// - 187500000core were taken as fee
	assert.Equal(t, "19622500000", balanceVoter1.Amount.String())
	assert.Equal(t, "19622500000", balanceVoter2.Amount.String())
	assert.Equal(t, "19622500000", balanceVoter3.Amount.String())
}

func depositProposal(ctx context.Context, t testing.T, chain testing.Chain, depositor types.Wallet, amount types.Coin, id uint64) {
	coredClient := chain.Client
	txBytes, err := coredClient.PrepareTxSubmitProposalDeposit(ctx, client.TxSubmitProposalDepositInput{
		Base:       buildBase(chain, depositor),
		Depositor:  depositor,
		ProposalID: id,
		Amount:     amount,
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}

func voteProposal(ctx context.Context, t testing.T, chain testing.Chain, voter types.Wallet, option govtypes.VoteOption, id uint64) types.Coin {
	coredClient := chain.Client
	txBytes, err := coredClient.PrepareTxSubmitProposalVote(ctx, client.TxSubmitProposalVoteInput{
		Base:       buildBase(chain, voter),
		Voter:      voter,
		ProposalID: id,
		Option:     option,
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	// Check vote
	votes, err := coredClient.QueryProposalVotes(ctx, id)
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

func buildBase(chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice: types.Coin{
			Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice.BigInt(),
			Denom:  chain.NetworkConfig.TokenSymbol,
		},
	}
}
