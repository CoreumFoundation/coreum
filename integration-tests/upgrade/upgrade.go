package upgrade

import (
	"context"
	"strings"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
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

	infoBefore, err := info(ctx, chain.ClientContext.Client())
	requireT.NoError(err)
	require.False(t, strings.HasSuffix(infoBefore.Version, "-upgrade"))
	upgradeHeight := infoBefore.LastBlockHeight + 30

	// Create new proposer.
	proposer := chain.RandomWallet()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

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

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	log.Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	// Verify that upgrade plan is there waiting to be applied.
	currentPlan, err = upgradeClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	requireT.NoError(err)
	requireT.NotNil(currentPlan.Plan)
	assert.Equal(t, "upgrade", currentPlan.Plan.Name)
	assert.Equal(t, upgradeHeight, currentPlan.Plan.Height)

	infoWaiting, err := info(ctx, chain.ClientContext.Client())
	requireT.NoError(err)
	log.Info("Waiting for upgrade", zap.Int64("upgradeHeight", upgradeHeight), zap.Int64("currentHeight", infoWaiting.LastBlockHeight))

	retryCtx, cancel := context.WithTimeout(ctx, 3*time.Second*time.Duration(upgradeHeight-infoWaiting.LastBlockHeight))
	defer cancel()
	var infoAfter abci.ResponseInfo
	err = retry.Do(retryCtx, time.Second, func() error {
		var err error
		infoAfter, err = info(ctx, chain.ClientContext.Client())
		if err != nil {
			return retry.Retryable(err)
		}
		if infoAfter.LastBlockHeight >= upgradeHeight {
			return nil
		}
		return retry.Retryable(errors.Errorf("waiting for upgraded block %d, current block: %d", upgradeHeight, infoAfter.LastBlockHeight))
	})
	requireT.NoError(err)

	// Verify that upgrade was applied on chain.
	appliedPlan, err := upgradeClient.AppliedPlan(ctx, &upgradetypes.QueryAppliedPlanRequest{
		Name: "upgrade",
	})
	requireT.NoError(err)
	assert.Equal(t, upgradeHeight, appliedPlan.Height)

	// Verify that node was restarted by cosmovisor and new version is running.
	assert.Equal(t, infoBefore.Version+"-upgrade", infoAfter.Version)
}

func info(ctx context.Context, client client.Client) (abci.ResponseInfo, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	i, err := client.ABCIInfo(requestCtx)
	if err != nil {
		return abci.ResponseInfo{}, errors.WithStack(err)
	}
	return i.Response, nil
}
