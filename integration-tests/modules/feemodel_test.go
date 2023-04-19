//go:build integrationtests

package modules

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestFeeModelQueryingMinGasPrice check that it's possible to query current minimum gas price required by the network.
func TestFeeModelQueryingMinGasPrice(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	feemodelClient := feemodeltypes.NewQueryClient(chain.ClientContext)
	res, err := feemodelClient.MinGasPrice(ctx, &feemodeltypes.QueryMinGasPriceRequest{})
	require.NoError(t, err)

	logger.Get(ctx).Info("Queried minimum gas price required", zap.Stringer("gasPrice", res.MinGasPrice))

	params := chain.NetworkConfig.Fee.FeeModel.Params()
	model := feemodeltypes.NewModel(params)

	require.False(t, res.MinGasPrice.Amount.IsNil())
	assert.True(t, res.MinGasPrice.Amount.GTE(model.CalculateGasPriceWithMaxDiscount()))
	assert.True(t, res.MinGasPrice.Amount.LTE(model.CalculateMaxGasPrice()))
	assert.Equal(t, chain.NetworkConfig.Denom, res.MinGasPrice.Denom)
}

// TestFeeModelProposalParamChange checks that feemodel param change proposal works correctly.
func TestFeeModelProposalParamChange(t *testing.T) {
	integrationtests.SkipUnsafe(t)
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	targetMaxDiscount := sdk.MustNewDecFromStr("0.12345")

	requireT := require.New(t)
	feeModelClient := feemodeltypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	// For the test we need to create the proposal twice.
	proposerBalance = proposerBalance.Add(proposerBalance)
	requireT.NoError(err)
	err = chain.Faucet.FundAccounts(ctx, integrationtests.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	feeModelParamsRes, err := feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)

	// Create invalid proposal MaxGasPrice = InitialGasPrice.
	feeModelParams := feeModelParamsRes.Params.Model
	feeModelParams.MaxGasPriceMultiplier = sdk.OneDec()
	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(ctx, proposer, paramproposal.NewParameterChangeProposal("Invalid proposal", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	))
	requireT.NoError(err)
	_, err = chain.Governance.Propose(ctx, proposalMsg)
	requireT.True(govtypes.ErrInvalidProposalContent.Is(err))

	// Create proposal to change MaxDiscount.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	feeModelParams = feeModelParamsRes.Params.Model
	feeModelParams.MaxDiscount = targetMaxDiscount
	requireT.NoError(err)
	proposalMsg, err = chain.Governance.NewMsgSubmitProposal(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxDiscount", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	))
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, proposalMsg)
	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Uint64("proposalID", proposalID))

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, finalStatus)

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
