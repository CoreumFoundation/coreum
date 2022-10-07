package feemodel

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestFeeModelProposalParamChange checks that feemodel param change proposal works correctly.
func TestFeeModelProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	targetMaxDiscount := sdk.MustNewDecFromStr("0.12345")

	requireT := require.New(t)
	feeModelClient := feemodeltypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.RandomWallet()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	// For the test we need to create the proposal twice.
	proposerBalance = proposerBalance.Add(proposerBalance)
	requireT.NoError(err)
	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	feeModelParamsRes, err := feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)

	// Create invalid proposal MaxGasPrice = InitialGasPrice.
	feeModelParams := feeModelParamsRes.Params.Model
	feeModelParams.MaxGasPriceMultiplier = sdk.OneDec()
	_, err = chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Invalid proposal", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	))
	requireT.True(client.IsErr(err, govtypes.ErrInvalidProposalContent))

	// Create proposal to change MaxDiscount.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	feeModelParams = feeModelParamsRes.Params.Model
	feeModelParams.MaxDiscount = targetMaxDiscount
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxDiscount", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	))
	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Int("proposalID", proposalID))

	// Wait for voting period to be started.
	proposal, err := chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusVotingPeriod, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	proposal, err = chain.Governance.WaitForProposalStatus(ctx, govtypes.StatusPassed, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, proposal.Status)

	// Check the proposed change is applied.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(feeModelParams.String(), feeModelParamsRes.Params.Model.String())
}

func marshalParamChangeProposal(requireT *require.Assertions, modelParams feemodeltypes.ModelParams) string {
	str, err := tmjson.Marshal(modelParams)
	requireT.NoError(err)
	return string(str)
}
