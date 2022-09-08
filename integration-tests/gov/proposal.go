package gov

import (
	"context"
	"strconv"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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
	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()

	// Calculate a voter balance based on min amount to be delegated
	bondedTokens, err := chain.Client.GetBondedTokens(ctx)
	require.NoError(t, err)
	voterDelegateAmount := bondedTokens.MulRaw(52).QuoRaw(100).QuoRaw(2)

	// Prepare initial balances
	minDepositAmount, ok := sdk.NewIntFromString(chain.NetworkConfig.GovConfig.ProposalConfig.MinDepositAmount)
	require.True(t, ok)
	proposerInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		minDepositAmount,
	)
	voterInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
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
	validators, err := chain.Client.GetValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	valAddress, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegateAmount := testing.MustNewCoin(t, voterDelegateAmount, chain.NetworkConfig.TokenSymbol)
	delegateCoins(ctx, t, chain, voter1, valAddress, delegateAmount)
	delegateCoins(ctx, t, chain, voter2, valAddress, delegateAmount)

	// Submit a param change proposal
	initialDeposit := testing.MustNewCoin(t, minDepositAmount, chain.NetworkConfig.TokenSymbol)
	txBytes, err := chain.Client.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBaseTxInput(t, chain, proposer),
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
		Content: paramproposal.NewParameterChangeProposal(
			"Change UnbondingTime",
			"Propose changing UnbondingTime in the staking module",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyUnbondingTime), "\"172800000000000\""),
			},
		),
	})
	require.NoError(t, err)
	result, err := chain.Client.Broadcast(ctx, txBytes)
	require.NoError(t, err)
	proposalIDStr, ok := client.FindEventAttribute(result.EventLogs, govtypes.EventTypeSubmitProposal, govtypes.AttributeKeyProposalID)
	require.True(t, ok)
	proposalID, err := strconv.Atoi(proposalIDStr)
	require.NoError(t, err)

	logger.Get(ctx).Info("Proposal has been submitted", zap.String("txHash", result.TxHash), zap.Int("proposalID", proposalID))

	// Wait for voting period to be started
	proposal := waitForProposalStatus(ctx, t, chain, govtypes.StatusVotingPeriod, testing.MinDepositPeriod, uint64(proposalID))
	assert.Equal(t, govtypes.StatusVotingPeriod, proposal.Status)

	// Vote for the proposal
	voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposal.ProposalId)
	voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposal.ProposalId)

	logger.Get(ctx).Info("2 voters have voted successfully")

	// Wait for proposal result
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusPassed, testing.MinVotingPeriod, proposal.ProposalId)
	assert.Equal(t, govtypes.StatusPassed, proposal.Status)
	assert.Equal(t, proposal.FinalTallyResult, govtypes.TallyResult{
		Yes:        sdk.NewIntFromBigInt(delegateAmount.Amount).MulRaw(2),
		Abstain:    sdk.NewInt(0),
		No:         sdk.NewInt(0),
		NoWithVeto: sdk.NewInt(0),
	})
}

func delegateCoins(ctx context.Context, t testing.T, chain testing.Chain, delegator types.Wallet, validator sdk.ValAddress, amount types.Coin) {
	txBytes, err := chain.Client.PrepareTxSubmitDelegation(ctx, client.TxSubmitDelegationInput{
		Base:      buildBaseTxInput(t, chain, delegator),
		Delegator: delegator,
		Validator: validator,
		Amount:    amount,
	})
	require.NoError(t, err)
	_, err = chain.Client.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}

func voteProposal(ctx context.Context, t testing.T, chain testing.Chain, voter types.Wallet, option govtypes.VoteOption, proposalID uint64) {
	txBytes, err := chain.Client.PrepareTxSubmitProposalVote(ctx, client.TxSubmitProposalVoteInput{
		Base:       buildBaseTxInput(t, chain, voter),
		Voter:      voter,
		ProposalID: proposalID,
		Option:     option,
	})
	require.NoError(t, err)
	_, err = chain.Client.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	// Check vote
	votes, err := chain.Client.QueryProposalVotes(ctx, proposalID)
	require.NoError(t, err)
	voterVotes, ok := votes[voter.Key.Address()]
	require.True(t, ok)
	require.Len(t, voterVotes, 1)
	require.Equal(t, voterVotes[0].Option, govtypes.OptionYes)
	require.Equal(t, voterVotes[0].Weight, sdk.NewDec(1))
}

func waitForProposalStatus(ctx context.Context, t testing.T, chain testing.Chain, status govtypes.ProposalStatus, duration time.Duration, proposalID uint64) *govtypes.Proposal {
	coredClient := chain.Client
	var lastStatus govtypes.ProposalStatus
	timeout := time.NewTimer(duration + time.Second)
	ticker := time.NewTicker(time.Second / 4)
	for range ticker.C {
		select {
		case <-timeout.C:
			t.Errorf("waiting for %s status is timed out for proposal %d and final status %s", status, proposalID, lastStatus)
			t.FailNow()
		default:
			proposal, err := coredClient.GetProposal(ctx, proposalID)
			require.NoError(t, err)

			if lastStatus = proposal.Status; lastStatus == status {
				return proposal
			}
		}
	}
	t.Errorf("waiting for %s status is timed out for proposal %d and final status %s", status, proposalID, lastStatus)
	t.FailNow()
	return nil
}

func buildBaseTxInput(t testing.T, chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice, chain.NetworkConfig.TokenSymbol),
	}
}
