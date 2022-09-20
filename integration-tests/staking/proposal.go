package staking

import (
	"context"
	"strconv"
	"time"

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

	// Create two random wallets
	proposer := testing.RandomWallet()

	// Prepare initial balances
	govParams, err := chain.Governance.GetParams(ctx)

	minDeposit := govParams.DepositParams.MinDeposit
	minDepositAmount := minDeposit.AmountOf(chain.NetworkConfig.TokenSymbol)

	proposerInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		minDepositAmount,
	)

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(proposer, chain.NewCoin(proposerInitialBalance)),
	))

	// Submit a param change proposal

	txBytes, err := chain.Client.PrepareTxSubmitProposal(ctx, client.TxSubmitProposalInput{
		Base:           buildBaseTxInput(chain, proposer),
		Proposer:       proposer,
		InitialDeposit: chain.NewCoin(minDepositAmount),
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

	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	require.NoError(t, err)

	logger.Get(ctx).Info("2 voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result
	proposal = waitForProposalStatus(ctx, t, chain, govtypes.StatusPassed, time.Until(proposal.VotingEndTime), proposal.ProposalId)
	assert.Equal(t, govtypes.StatusPassed, proposal.Status)

	// Check the proposed change is applied
	stakingParams, err := chain.Client.GetStakingParams(ctx)
	require.NoError(t, err)
	require.Equal(t, uint32(proposedMaxValidators), stakingParams.MaxValidators)
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
