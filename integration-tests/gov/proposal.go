package gov

import (
	"context"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	const proposedMaxValidators = 201
	minDepositMultiplier, err := sdk.NewDecFromStr("1.02")
	require.NoError(t, err)

	// Create two random wallets
	proposer := testing.RandomWallet()
	voter1 := testing.RandomWallet()
	voter2 := testing.RandomWallet()

	// Get gov tally params
	govTallyParams, err := chain.Client.GetGovTallyParams(ctx)
	require.NoError(t, err)

	// Calculate a voter balance based on min amount to be delegated
	bondedTokens, err := chain.Client.GetBondedTokens(ctx)
	require.NoError(t, err)
	voterDelegateAmount := bondedTokens.ToDec().Mul(govTallyParams.Threshold.Mul(minDepositMultiplier)).QuoInt64(2).RoundInt()

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
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(proposer, chain.NewCoin(proposerInitialBalance)),
		testing.NewFundedAccount(voter1, chain.NewCoin(voterInitialBalance)),
		testing.NewFundedAccount(voter2, chain.NewCoin(voterInitialBalance)),
	))

	// Delegate coins
	validators, err := chain.Client.GetValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)
	valAddress, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegateAmount := chain.NewCoin(voterDelegateAmount)
	delegateCoins(ctx, t, chain, voter1, valAddress, delegateAmount)
	delegateCoins(ctx, t, chain, voter2, valAddress, delegateAmount)

	// Submit a param change proposal
	initialDeposit := chain.NewCoin(minDepositAmount)
	txBytes, err := chain.Client.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBaseTxInput(chain, proposer),
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
		Content: paramproposal.NewParameterChangeProposal("Change MaxValidators", "Propose changing MaxValidators in the staking module",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), strconv.Itoa(proposedMaxValidators)),
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
	depositPeriod, err := time.ParseDuration(chain.NetworkConfig.GovConfig.ProposalConfig.MinDepositPeriod)
	require.NoError(t, err)
	proposal := waitForProposalStatus(ctx, t, chain, govtypes.StatusVotingPeriod, depositPeriod, uint64(proposalID))
	assert.Equal(t, govtypes.StatusVotingPeriod, proposal.Status)

	// Vote for the proposal
	voteProposal(ctx, t, chain, voter1, govtypes.OptionYes, proposal.ProposalId)
	voteProposal(ctx, t, chain, voter2, govtypes.OptionYes, proposal.ProposalId)

	logger.Get(ctx).Info("2 voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusPassed, time.Until(proposal.VotingEndTime), proposal.ProposalId)
	assert.Equal(t, govtypes.StatusPassed, proposal.Status)
	assert.Equal(t, proposal.FinalTallyResult, govtypes.TallyResult{
		Yes:        delegateAmount.Amount.MulRaw(2),
		Abstain:    sdk.NewInt(0),
		No:         sdk.NewInt(0),
		NoWithVeto: sdk.NewInt(0),
	})

	// Check the proposed change is applied
	stakingParams, err := chain.Client.GetStakingParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint32(proposedMaxValidators), stakingParams.MaxValidators)
}

func delegateCoins(ctx context.Context, t testing.T, chain testing.Chain, delegator types.Wallet, validator sdk.ValAddress, amount sdk.Coin) {
	txBytes, err := chain.Client.PrepareTxSubmitDelegation(ctx, client.TxSubmitDelegationInput{
		Base:      buildBaseTxInput(chain, delegator),
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
		Base:       buildBaseTxInput(chain, voter),
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
	require.True(t, ok, "%#v, %s", votes, voter.Key.Address())
	require.Len(t, voterVotes, 1)
	require.Equal(t, voterVotes[0].Option, govtypes.OptionYes)
	require.Equal(t, voterVotes[0].Weight, sdk.NewDec(1))
}

func waitForProposalStatus(ctx context.Context, t testing.T, chain testing.Chain, status govtypes.ProposalStatus, duration time.Duration, proposalID uint64) *govtypes.Proposal {
	var lastStatus govtypes.ProposalStatus
	timeout := time.NewTimer(duration + time.Second*10)
	ticker := time.NewTicker(time.Millisecond * 250)
	for range ticker.C {
		select {
		case <-ctx.Done():
			t.Errorf("canceled context")
			t.FailNow()
		case <-timeout.C:
			t.Errorf("waiting for %s status is timed out for proposal %d and final status %s", status, proposalID, lastStatus)
			t.FailNow()
		default:
			proposal, err := chain.Client.GetProposal(ctx, proposalID)
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

func buildBaseTxInput(chain testing.Chain, signer types.Wallet) tx.BaseInput {
	return tx.BaseInput{
		Signer:   signer,
		GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
	}
}
