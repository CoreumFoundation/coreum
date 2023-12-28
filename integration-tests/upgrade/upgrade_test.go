//go:build integrationtests

package upgrade

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	appupgradev3 "github.com/CoreumFoundation/coreum/v4/app/upgrade/v3"
	appupgradev4 "github.com/CoreumFoundation/coreum/v4/app/upgrade/v4"
	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
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
	case "v3.0.2":
		upgradeV3ToV4(t)
	case "v2.0.2":
		upgradeV2ToV3(t)
	default:
		requireT.Failf("not supported version", "version: %s", infoRes.ApplicationVersion.Version)
	}
}

func upgradeV3ToV4(t *testing.T) {
	tests := []upgradeTest{}

	for _, test := range tests {
		test.Before(t)
	}

	runUpgrade(t, appupgradev4.Name, 30)

	for _, test := range tests {
		test.After(t)
	}
}

func upgradeV2ToV3(t *testing.T) {
	runLegacyUpgrade(t, appupgradev3.Name, 30)
}

func runUpgrade(
	t *testing.T,
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

	latestBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	upgradeHeight := latestBlockRes.Block.Header.Height + blocksToWait //nolint:staticcheck

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(proposer, proposerBalance))

	t.Logf("Creating proposal for upgrading, upgradeName:%s, upgradeHeight:%d", upgradeName, upgradeHeight)

	msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Plan: upgradetypes.Plan{
			Name:   upgradeName,
			Height: upgradeHeight,
		},
	}

	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{msgUpgrade},
		"Upgrade chain",
		"Upgrade "+upgradeName,
		"Running "+upgradeName+" in integration tests",
	)

	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)
	t.Logf("Upgrade proposal has been submitted, proposalID:%d", proposalID)

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, proposal.Id)
	requireT.NoError(err)

	t.Logf(
		"Voters have voted successfully, waiting for voting period to be finished, votingEndTime: %s",
		proposal.VotingEndTime,
	)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

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

	//nolint:staticcheck
	retryCtx, cancel := context.WithTimeout(
		ctx,
		6*time.Second*time.Duration(upgradeHeight-infoWaitingBlockRes.Block.Header.Height),
	)
	defer cancel()
	//nolint:staticcheck
	t.Logf(
		"Waiting for upgrade, upgradeHeight:%d, currentHeight:%d",
		upgradeHeight,
		infoWaitingBlockRes.Block.Header.Height,
	)
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
		//nolint:staticcheck
		return retry.Retryable(errors.Errorf(
			"waiting for upgraded block %d, current block: %d",
			upgradeHeight,
			infoAfterBlockRes.Block.Header.Height,
		))
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

func runLegacyUpgrade(
	t *testing.T,
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

	latestBlockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	upgradeHeight := latestBlockRes.Block.Header.Height + blocksToWait //nolint:staticcheck

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.LegacyGovernance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(proposer, proposerBalance))

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

	t.Logf("Voters have voted successfully, waiting for voting period to be finished, votingEndTime: %s",
		proposal.VotingEndTime)

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
	//nolint:staticcheck
	requireT.Less(infoWaitingBlockRes.Block.Header.Height, upgradeHeight)

	//nolint:staticcheck
	retryCtx, cancel := context.WithTimeout(ctx,
		6*time.Second*time.Duration(upgradeHeight-infoWaitingBlockRes.Block.Header.Height))
	defer cancel()
	//nolint:staticcheck
	t.Logf("Waiting for upgrade, upgradeHeight:%d, currentHeight:%d",
		upgradeHeight, infoWaitingBlockRes.Block.Header.Height)
	err = retry.Do(retryCtx, time.Second, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		var err error
		infoAfterBlockRes, err := tmQueryClient.GetLatestBlock(requestCtx, &tmservice.GetLatestBlockRequest{})
		if err != nil {
			return retry.Retryable(err)
		}
		//nolint:staticcheck
		if infoAfterBlockRes.Block.Header.Height >= upgradeHeight+1 {
			return nil
		}
		//nolint:staticcheck
		return retry.Retryable(errors.Errorf("waiting for upgraded block %d, current block: %d",
			upgradeHeight, infoAfterBlockRes.Block.Header.Height))
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
