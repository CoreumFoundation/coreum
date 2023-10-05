//go:build integrationtests

package modules

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	feemodeltypes "github.com/CoreumFoundation/coreum/v3/x/feemodel/types"
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

	requireT := require.New(t)
	assertT := assert.New(t)
	feeModelClient := feemodeltypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	// For the test we need to create the proposal twice.
	proposerBalance = proposerBalance.Add(proposerBalance)
	requireT.NoError(err)
	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(proposer, proposerBalance))

	feeModelParamsRes, err := feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	oldParams := feeModelParamsRes.Params

	// Create invalid proposal MaxGasPriceMultiplier = 1.
	newParams := oldParams
	newParams.Model.MaxGasPriceMultiplier = sdk.OneDec()

	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(
		ctx, proposer,
		[]sdk.Msg{&feemodeltypes.MsgUpdateParams{
			Params:    newParams,
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		}},
		"-", "-", "-",
	)

	requireT.NoError(err)
	_, err = chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.ErrorIs(err, govtypes.ErrInvalidProposalMsg)

	// Create proposal to change MaxDiscount.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	targetMaxDiscount := sdk.MustNewDecFromStr("0.12345")
	newParams = feeModelParamsRes.Params
	newParams.Model.MaxDiscount = targetMaxDiscount
	requireT.NoError(err)
	chain.Governance.ProposalFromMsgAndVote(
		ctx, t, nil,
		"-", "-", "-", govtypesv1.OptionYes,
		&feemodeltypes.MsgUpdateParams{
			Params:    newParams,
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		},
	)

	// Check the proposed change is applied.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	assertT.Equal(feeModelParamsRes.Params.Model.MaxDiscount.String(), targetMaxDiscount.String())
	assertT.Equal(feeModelParamsRes.Params.Model.String(), feeModelParamsRes.Params.Model.String())
}

func getFeemodelParams(ctx context.Context, t *testing.T, clientCtx client.Context) feemodeltypes.ModelParams {
	queryClient := feemodeltypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.Model
}
