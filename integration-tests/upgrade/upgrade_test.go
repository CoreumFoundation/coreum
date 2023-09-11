//go:build integrationtests

package upgrade

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	appupgradev3 "github.com/CoreumFoundation/coreum/v3/app/upgrade/v3"
	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
)

type upgradeTest interface {
	Before(t *testing.T)
	After(t *testing.T)
}

// TestUpgrade that after accepting upgrade proposal cosmovisor starts a new version of cored.
func TestUpgrade(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	tmQueryClient := tmservice.NewServiceClient(chain.ClientContext)
	infoRes, err := tmQueryClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	requireT.NoError(err)

	switch infoRes.ApplicationVersion.Version {
	case "v2.0.2":
		upgradeV3(t)
	default:
		requireT.Failf("not supported version: %s", infoRes.ApplicationVersion.Version)
	}
}

func upgradeV3(t *testing.T) {
	tests := []upgradeTest{
		&paramsMigrationTest{},
		&wasmMigrationTest{},
	}

	for _, test := range tests {
		test.Before(t)
	}

	runUpgrade(t, "v2.0.2", appupgradev3.Name, 30)

	for _, test := range tests {
		test.After(t)
	}
}

// Note that inside this method we use deprecated Block attributed of GetLatestBlockResponse (latestBlockRes.Block)
// because we interact with older version of SDK before upgrade, and it doesn't have new SdkBlock attribute set.
// We also use deprecated v1beta1 gov because v1 doesn't exist in cored v2.0.2.
func runUpgrade(
	t *testing.T,
	oldBinaryVersion string,
	upgradeName string,
	blocksToWait int64,
) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	upgradeClient := upgradetypes.NewQueryClient(chain.ClientContext)

	// Verify that there is no ongoing upgrade plan.
	currentPlan, err := upgradeClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	requireT.NoError(err)
	requireT.Nil(currentPlan.Plan)

	tmQueryClient := tmservice.NewServiceClient(chain.ClientContext)
	infoBeforeRes, err := tmQueryClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	requireT.NoError(err)
	// we start with the old binary version
	require.Equal(t, oldBinaryVersion, infoBeforeRes.ApplicationVersion.Version)

	latestBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	upgradeHeight := latestBlockRes.Block.Header.Height + blocksToWait //nolint:staticcheck

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.LegacyGovernance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	t.Logf("Creating proposal for upgrading, upgradeName:%s, upgradeHeight:%d", upgradeName, upgradeHeight)

	proposalMsg, err := chain.LegacyGovernance.NewMsgSubmitProposalV1Beta1(
		ctx,
		proposer,
		upgradetypes.NewSoftwareUpgradeProposal(
			"Upgrade "+upgradeName,
			"Running "+upgradeName+" in integration tests",
			upgradetypes.Plan{
				Name:   upgradeName,
				Height: upgradeHeight,
			},
		))

	requireT.NoError(err)
	proposalID, err := chain.LegacyGovernance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)
	t.Logf("Upgrade proposal has been submitted, proposalID:%d", proposalID)

	// Verify that voting period started.
	proposal, err := chain.LegacyGovernance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.LegacyGovernance.VoteAll(ctx, govtypesv1beta1.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	t.Logf("Voters have voted successfully, waiting for voting period to be finished, votingEndTime: %s", proposal.VotingEndTime)

	// Wait for proposal result.
	finalStatus, err := chain.LegacyGovernance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusPassed, finalStatus)

	// Verify that upgrade plan is there waiting to be applied.
	currentPlan, err = upgradeClient.CurrentPlan(ctx, &upgradetypes.QueryCurrentPlanRequest{})
	requireT.NoError(err)
	requireT.NotNil(currentPlan.Plan)
	assert.Equal(t, upgradeName, currentPlan.Plan.Name)
	assert.Equal(t, upgradeHeight, currentPlan.Plan.Height)

	// Verify that we are before the upgrade
	infoWaitingBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	requireT.NoError(err)
	requireT.Less(infoWaitingBlockRes.Block.Header.Height, upgradeHeight) //nolint:staticcheck

	retryCtx, cancel := context.WithTimeout(ctx, 6*time.Second*time.Duration(upgradeHeight-infoWaitingBlockRes.Block.Header.Height)) //nolint:staticcheck
	defer cancel()
	t.Logf("Waiting for upgrade, upgradeHeight:%d, currentHeight:%d", upgradeHeight, infoWaitingBlockRes.Block.Header.Height) //nolint:staticcheck
	err = retry.Do(retryCtx, time.Second, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var err error
		infoAfterBlockRes, err := tmQueryClient.GetLatestBlock(requestCtx, &tmservice.GetLatestBlockRequest{})
		if err != nil {
			return retry.Retryable(err)
		}
		if infoAfterBlockRes.Block.Header.Height >= upgradeHeight+1 { //nolint:staticcheck
			return nil
		}
		return retry.Retryable(errors.Errorf("waiting for upgraded block %d, current block: %d", upgradeHeight, infoAfterBlockRes.Block.Header.Height)) //nolint:staticcheck
	})
	requireT.NoError(err)

	// Verify that upgrade was applied on chain.
	appliedPlan, err := upgradeClient.AppliedPlan(ctx, &upgradetypes.QueryAppliedPlanRequest{
		Name: upgradeName,
	})
	requireT.NoError(err)
	assert.Equal(t, upgradeHeight, appliedPlan.Height)
	t.Logf("Upgrade passed, applied plan height: %d", appliedPlan.Height)

	// The new binary isn't equal to initial
	infoAfterRes, err := tmQueryClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	requireT.NoError(err)
	t.Logf("New binary version: %s", infoAfterRes.ApplicationVersion.Version)
	assert.NotEqual(t, infoAfterRes.ApplicationVersion.Version, infoBeforeRes.ApplicationVersion.Version)
}
