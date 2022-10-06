package upgrade

import (
	"context"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// TestUpgrade that after accepting upgrade proposal cosmovisor starts a new version of cored.
func TestUpgrade(ctx context.Context, t testing.T, chain testing.Chain) {
	log := logger.Get(ctx)
	requireT := require.New(t)
	upgradeClient := upgradetypes.NewQueryClient(chain.ClientContext)

	// Verify that there is no ongoing upgrade plan.
	currentPlan, err := upgradeClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	requireT.NoError(err)
	requireT.Nil(currentPlan.Plan)

	// Create new proposer.
	proposer := chain.RandomWallet()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	height, err := latestHeight(ctx, chain.ClientContext.Client())
	requireT.NoError(err)
	upgradeHeight := height + 30

	log.Info("Creating proposal for upgrading", zap.Int64("upgradeHeight", upgradeHeight))

	// Create proposition to upgrade chain.
	proposalID, err := chain.Governance.Propose(ctx, proposer, upgradetypes.NewSoftwareUpgradeProposal("Upgrade test", "Testing if new version of node is started by cosmovisor",
		upgradetypes.Plan{
			Name:   "upgrade",
			Height: upgradeHeight,
		},
	))
	requireT.NoError(err)
	log.Info("Upgrade proposal has been submitted", zap.Int("proposalID", proposalID))

	// Wait for voting period to be started.
	proposal, err := chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusVotingPeriod, uint64(proposalID))
	requireT.NoError(err)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	log.Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	_, err = chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusPassed, uint64(proposalID))
	requireT.NoError(err)

	height, err = latestHeight(ctx, chain.ClientContext.Client())
	requireT.NoError(err)
	log.Info("Waiting for upgrade", zap.Int64("upgradeHeight", upgradeHeight), zap.Int64("currentHeight", height))

	retryCtx, cancel := context.WithTimeout(ctx, 3*time.Second*time.Duration(upgradeHeight-height))
	defer cancel()
	err = retry.Do(retryCtx, time.Second, func() error {
		height, err := latestHeight(ctx, chain.ClientContext.Client())
		if err != nil {
			return retry.Retryable(err)
		}
		if height >= upgradeHeight {
			return nil
		}
		return retry.Retryable(errors.Errorf("waiting for upgraded block %d, current block: %d", upgradeHeight, height))
	})
	requireT.NoError(err)

	// Verify that upgrade was applied on chain.
	currentPlan, err = upgradeClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	requireT.NoError(err)
	requireT.NotNil(currentPlan.Plan)
	assert.Equal(t, "upgrade", currentPlan.Plan.Name)
	assert.Equal(t, upgradeHeight, currentPlan.Plan.Height)

	// Verify that node was restarted by cosmovisor.
	assert.Equal(t, "upgrade", moniker(t, ctx, chain.ClientContext.Client()))
}

func latestHeight(ctx context.Context, client client.Client) (int64, error) {
	s, err := status(ctx, client)
	if err != nil {
		return 0, err
	}
	return s.SyncInfo.LatestBlockHeight, nil
}

func moniker(t testing.T, ctx context.Context, client client.Client) string {
	s, err := status(ctx, client)
	require.NoError(t, err)
	return s.NodeInfo.Moniker
}

func status(ctx context.Context, client client.Client) (*coretypes.ResultStatus, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	s, err := client.Status(requestCtx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return s, nil
}
