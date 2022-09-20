package staking

import (
	"context"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"strconv"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// TestProposalParamChange checks that param change proposal works correctly.
func TestProposalParamChange(ctx context.Context, t testing.T, chain *testing.Chain) {
	const targetMaxValidators = 201
	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer, err := chain.Governance.CreateProposer(ctx)
	requireT.NoError(err)

	// Create proposition to change max validators value.
	proposalID, err := chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxValidators", "Propose changing MaxValidators in the staking module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), strconv.Itoa(targetMaxValidators)),
		},
	))
	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Int("proposalID", proposalID))

	// Wait for voting period to be started.
	proposal, err := chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusVotingPeriod, uint64(proposalID))
	assert.Equal(t, govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	proposal, err = chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusPassed, uint64(proposalID))
	requireT.Equal(govtypes.StatusPassed, proposal.Status)

	// Check the proposed change is applied
	stakingParams, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(uint32(targetMaxValidators), stakingParams.Params.MaxValidators)
}
