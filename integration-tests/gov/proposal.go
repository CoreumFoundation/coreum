package gov

import (
	"context"
	"math/big"
	"time"

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

const (
	// gasLimitThreshold is the threshold added to the gas limit number to add a little more space if a transaction
	// would cost a little more that was expected.
	gasLimitThreshold = 20000
)

// TestProposalParamChange checks that param change proposal works correctly
func TestProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create client so we can send transactions and query state
	coredClient := chain.Client

	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()

	// Calculate a voter balance based on min amount to be delegated
	validators, err := coredClient.GetValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	totalDelegated := sdk.NewInt(0)
	for _, validator := range validators {
		totalDelegated = totalDelegated.Add(validator.Tokens)
	}
	voterDelegateAmount := totalDelegated.MulRaw(52).QuoRaw(100).QuoRaw(2)

	// Prepare initial balances
	proposerInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		getGasLimit(chain),
		1,
		sdk.NewInt(11000000),
	)
	voterInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		getGasLimit(chain),
		3,
		voterDelegateAmount,
	)

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

	// Delegate coins
	valAddress, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegateAmount := testing.MustNewCoin(t, voterDelegateAmount, chain.NetworkConfig.TokenSymbol)
	delegateCoins(ctx, t, chain, voter1, valAddress, delegateAmount)
	delegateCoins(ctx, t, chain, voter2, valAddress, delegateAmount)

	// Submit a param change proposal
	initialDeposit := testing.MustNewCoin(t, sdk.NewInt(10000000), chain.NetworkConfig.TokenSymbol)
	txBytes, err := coredClient.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBase(t, chain, proposer),
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
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
	assert.Equal(t,
		proposerInitialBalance.Sub(getBaseTransactionFee(chain)).Sub(sdk.NewIntFromBigInt(initialDeposit.Amount)).BigInt(),
		balancesProposer[chain.NetworkConfig.TokenSymbol].Amount,
	)

	logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash))

	// Wait for voting period to be started
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusVotingPeriod, testing.MinDepositPeriod, proposal.ProposalId)
	assert.Equal(t, govtypes.StatusVotingPeriod, proposal.Status)

	// Vote for the proposal
	voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposal.ProposalId)
	voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposal.ProposalId)

	logger.Get(ctx).Info("2 voters have voted successfully")

	// Wait for proposal result
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusPassed, testing.MinVotingPeriod, proposal.ProposalId)
	assert.Equal(t, govtypes.StatusPassed, proposal.Status)
	assert.Equal(t, big.NewInt(0).Mul(delegateAmount.Amount, big.NewInt(2)), proposal.FinalTallyResult.Yes.BigInt())
	assert.Equal(t, int64(0), proposal.FinalTallyResult.No.Int64())
	assert.Equal(t, int64(0), proposal.FinalTallyResult.Abstain.Int64())
	assert.Equal(t, int64(0), proposal.FinalTallyResult.NoWithVeto.Int64())
}

func delegateCoins(ctx context.Context, t testing.T, chain testing.Chain, delegator types.Wallet, validator sdk.ValAddress, amount types.Coin) {
	coredClient := chain.Client
	txBytes, err := coredClient.PrepareTxSubmitDelegation(ctx, client.TxSubmitDelegationInput{
		Base:      buildBase(t, chain, delegator),
		Delegator: delegator,
		Validator: validator,
		Amount:    amount,
	})
	require.NoError(t, err)
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}

func voteProposal(ctx context.Context, t testing.T, chain testing.Chain, voter types.Wallet, option govtypes.VoteOption, proposalID uint64) {
	coredClient := chain.Client

	// Query voter initial balance
	initialBalances, err := coredClient.QueryBankBalances(ctx, voter)
	require.NoError(t, err)

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
	finalBalances, err := coredClient.QueryBankBalances(ctx, voter)
	require.NoError(t, err)

	// Check balance
	assert.Equal(t,
		big.NewInt(0).Sub(
			initialBalances[chain.NetworkConfig.TokenSymbol].Amount,
			getBaseTransactionFee(chain).BigInt(),
		),
		finalBalances[chain.NetworkConfig.TokenSymbol].Amount,
	)
}

func waitForProposalStatus(ctx context.Context, t testing.T, chain testing.Chain, status govtypes.ProposalStatus, duration time.Duration, proposalID uint64) *govtypes.Proposal {
	coredClient := chain.Client
	timeout := time.NewTimer(duration)
	ticker := time.NewTicker(time.Second / 4)
	for range ticker.C {
		select {
		case <-timeout.C:
			t.Errorf("waiting for %s status is timed out for proposal %d", status, proposalID)
			t.FailNow()
		default:
			proposal, err := coredClient.GetProposal(ctx, proposalID)
			require.NoError(t, err)

			if proposal.Status == status {
				return proposal
			}
		}
	}
	t.Errorf("waiting for %s status is timed out for proposal %d", status, proposalID)
	t.FailNow()
	return nil
}

func getBaseTransactionFee(chain testing.Chain) sdk.Int {
	return chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice.Mul(
		sdk.NewIntFromUint64(getGasLimit(chain)),
	)
}

func buildBase(t testing.T, chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: getGasLimit(chain),
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice, chain.NetworkConfig.TokenSymbol),
	}
}

func getGasLimit(chain testing.Chain) uint64 {
	return chain.NetworkConfig.Fee.DeterministicGas.BankSend + gasLimitThreshold
}
