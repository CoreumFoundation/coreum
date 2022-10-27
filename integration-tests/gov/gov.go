package gov

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestProposalWithDepositAndWeightedVotes - todo
func TestProposalWithDepositAndWeightedVotes(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)

	missingDepositAmount := chain.NewCoin(sdk.NewInt(10))

	gov := chain.Governance

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount)

	depositor := chain.GenAccount()
	depositorBalance := chain.NewCoin(
		testing.ComputeNeededBalance(
			chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
			chain.GasLimitByMsgs(&govtypes.MsgDeposit{}),
			1,
			missingDepositAmount.Amount,
		),
	)

	err = chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(proposer, proposerBalance),
		testing.NewFundedAccount(depositor, depositorBalance),
	)
	requireT.NoError(err)

	govParams, err := gov.QueryGovParams(ctx)
	requireT.NoError(err)

	msg, err := govtypes.NewMsgSubmitProposal(
		govtypes.NewTextProposal("abc", "def"),
		govParams.DepositParams.MinDeposit.Sub(depositorBalance).,
		proposer,
	)
	chain.Governance.ProposeV2(ctx, )
	chain.Governance.Propose(ctx, proposer, govtypes.NewTextProposal())

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
	requireT.True(govtypes.ErrInvalidProposalContent.Is(err))

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

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// Wait for proposal result.
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	// Check the proposed change is applied.
	feeModelParamsRes, err = feeModelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(feeModelParams.String(), feeModelParamsRes.Params.Model.String())
}
