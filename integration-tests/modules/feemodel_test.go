//go:build integrationtests

package modules

import (
	"context"
	"testing"

	tmjson "github.com/cometbft/cometbft/libs/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	feemodeltypes "github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

// TestFeeModelQueryingMinGasPrice check that it's possible to query current minimum gas price required by the network.
func TestFeeModelQueryingMinGasPrice(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	feemodelClient := feemodeltypes.NewQueryClient(chain.ClientContext)
	res, err := feemodelClient.MinGasPrice(ctx, &feemodeltypes.QueryMinGasPriceRequest{})
	require.NoError(t, err)

	t.Logf("Queried minimum gas price required, gasPrice:%s", res.MinGasPrice)

	model := feemodeltypes.NewModel(getFeemodelParams(ctx, t, chain.ClientContext))

	require.False(t, res.MinGasPrice.Amount.IsNil())
	assert.True(t, res.MinGasPrice.Amount.GTE(model.CalculateGasPriceWithMaxDiscount()))
	assert.True(t, res.MinGasPrice.Amount.LTE(model.CalculateMaxGasPrice()))
	assert.Equal(t, chain.ChainSettings.Denom, res.MinGasPrice.Denom)
}

// TestFeeModelQueryingGasPriceRecommendation check that recommendation end point is called correctly.
func TestFeeModelQueryingGasPriceRecommendation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	feemodelClient := feemodeltypes.NewQueryClient(chain.ClientContext)
	res, err := feemodelClient.RecommendedGasPrice(ctx, &feemodeltypes.QueryRecommendedGasPriceRequest{AfterBlocks: 50})
	requireT.NoError(err)
	requireT.NotNil(res)

	model := feemodeltypes.NewModel(getFeemodelParams(ctx, t, chain.ClientContext))
	requireT.GreaterOrEqual(res.GetHigh().Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
	requireT.LessOrEqual(res.GetHigh().Amount.MustFloat64(), model.CalculateMaxGasPrice().MustFloat64())
	requireT.GreaterOrEqual(res.GetLow().Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
	requireT.LessOrEqual(res.GetLow().Amount.MustFloat64(), model.CalculateMaxGasPrice().MustFloat64())
	requireT.GreaterOrEqual(res.GetMed().Amount.MustFloat64(), model.CalculateGasPriceWithMaxDiscount().MustFloat64())
	requireT.LessOrEqual(res.GetMed().Amount.MustFloat64(), model.CalculateMaxGasPrice().MustFloat64())

	requireT.LessOrEqual(res.GetLow().Amount.MustFloat64(), res.GetMed().Amount.MustFloat64())
	requireT.LessOrEqual(res.GetMed().Amount.MustFloat64(), res.GetHigh().Amount.MustFloat64())
}

// TestFeeModelProposalParamChange checks that feemodel param change proposal works correctly.
func TestFeeModelProposalParamChange(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	targetMaxDiscount := sdk.MustNewDecFromStr("0.12345")

	requireT := require.New(t)
	feeModelClient := feemodeltypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	// For the test we need to create the proposal twice.
	proposerBalance = proposerBalance.Add(proposerBalance)
	requireT.NoError(err)
	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	feeModelParamsRes, err := feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)

	// Create invalid proposal MaxGasPrice = InitialGasPrice.
	feeModelParams := feeModelParamsRes.Params.Model
	feeModelParams.MaxGasPriceMultiplier = sdk.OneDec()
	proposalMsg := chain.LegacyGovernance.NewParamsChangeProposal(ctx, t, proposer, "Invalid proposal", "-", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	)
	requireT.NoError(err)
	_, err = chain.LegacyGovernance.Propose(ctx, t, proposalMsg)
	requireT.True(govtypes.ErrInvalidProposalContent.Is(err))

	// Create proposal to change MaxDiscount.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	feeModelParams = feeModelParamsRes.Params.Model
	feeModelParams.MaxDiscount = targetMaxDiscount
	requireT.NoError(err)
	proposalMsg = chain.LegacyGovernance.NewParamsChangeProposal(
		ctx, t, proposer, "Change MaxDiscount", "-", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), marshalParamChangeProposal(requireT, feeModelParams),
			),
		},
	)
	requireT.NoError(err)
	proposalID, err := chain.LegacyGovernance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)
	t.Logf("Proposal has been submitted, proposalID:%d", proposalID)

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, proposal.Id)
	requireT.NoError(err)

	t.Logf("Voters have voted successfully, waiting for voting period to be finished, votingEndTime:%s", proposal.VotingEndTime)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

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

func getFeemodelParams(ctx context.Context, t *testing.T, clientCtx client.Context) feemodeltypes.ModelParams {
	queryClient := feemodeltypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.Model
}
