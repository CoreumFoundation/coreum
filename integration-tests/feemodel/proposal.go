package feemodel

import (
	"bytes"
	"context"
	"text/template"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestFeeModelProposalParamChange checks that feemodel param change proposal works correctly.
func TestFeeModelProposalParamChange(ctx context.Context, t testing.T, chain *testing.Chain) {
	targetMaxDiscount := sdk.MustNewDecFromStr("0.12345")

	requireT := require.New(t)
	feeModelClient := feemodeltypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer, err := chain.Governance.CreateProposer(ctx)
	requireT.NoError(err)

	feeModelParamsRes, err := feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)

	// Create invalid proposal MaxGasPrice = InitialGasPrice.
	feeModelParams := feeModelParamsRes.Params.Model
	feeModelParams.MaxGasPrice = feeModelParams.InitialGasPrice
	_, err = chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Invalid proposal", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), toProposalString(requireT, feeModelParams),
			),
		},
	))
	requireT.Error(err)
	require.True(t, client.IsErr(err, govtypes.ErrInvalidProposalContent))

	// Re-create new proposer to have enough deposit for the next proposal.
	proposer, err = chain.Governance.CreateProposer(ctx)
	requireT.NoError(err)

	// Create proposal to change MaxDiscount.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	feeModelParams = feeModelParamsRes.Params.Model
	feeModelParams.MaxDiscount = targetMaxDiscount
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxDiscount", "-",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(
				feemodeltypes.ModuleName, string(feemodeltypes.KeyModel), toProposalString(requireT, feeModelParams),
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

func toProposalString(requireT *require.Assertions, modelParams feemodeltypes.ModelParams) string {
	// the template is used here since the encoding expect the int to be wrapped to double quotes, but uint32 without.
	proposalTemplate := `
{
  "initial_gas_price": "{{.InitialGasPrice}}",
  "max_gas_price": "{{.MaxGasPrice}}",
  "max_discount": "{{.MaxDiscount}}",
  "escalation_start_block_gas":"{{.EscalationStartBlockGas}}",
  "max_block_gas":"{{.MaxBlockGas}}",
  "short_ema_block_length":{{.ShortEmaBlockLength}},
  "long_ema_block_length":{{.LongEmaBlockLength}}
}
`
	buf := new(bytes.Buffer)
	err := template.Must(template.New("").Parse(proposalTemplate)).Execute(buf, modelParams)
	requireT.NoError(err)

	return buf.String()
}
