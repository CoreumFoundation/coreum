package staking

import (
	"context"
	"strconv"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// TestStakingProposalParamChange checks that staking param change proposal works correctly.
func TestStakingProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	const targetMaxValidators = 201
	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	// Create proposition to change max validators value.
	proposalID, err := chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxValidators", "Propose changing MaxValidators in the staking module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), strconv.Itoa(targetMaxValidators)),
		},
	))
	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Int("proposalID", proposalID))

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	// Check the proposed change is applied.
	stakingParams, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(uint32(targetMaxValidators), stakingParams.Params.MaxValidators)
}
