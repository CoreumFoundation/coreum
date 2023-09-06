package integrationtests

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/CoreumFoundation/coreum/v2/testutil/event"
)

// GovernanceLegacy keep the test chain predefined account for the governance operations via v1beta1 API.
// This structure will be removed in the future once we:
// 1. Fully migrate new params (which are stored inside each module instead of params module).
// 2. Migrate to new proposal types. Mostly initialize by NewMsgSubmitProposalV1 method.
// 3. Get rid of interactions with cored v2 (inside upgrade tests) which uses v1beta1 API.
type GovernanceLegacy struct {
	chainCtx       ChainContext
	faucet         Faucet
	govClient      govtypesv1beta1.QueryClient
	stakerAccounts []sdk.AccAddress
	muCh           chan struct{}
}

// NewGovernanceLegacy returns the new instance of GovernanceLegacy.
func NewGovernanceLegacy(gov Governance) GovernanceLegacy {
	return GovernanceLegacy{
		chainCtx:       gov.chainCtx,
		faucet:         gov.faucet,
		stakerAccounts: gov.stakerAccounts,
		govClient:      govtypesv1beta1.NewQueryClient(gov.chainCtx.ClientContext),
		muCh:           gov.muCh,
	}
}

// ComputeProposerBalance computes the balance required for the proposer.
func (g GovernanceLegacy) ComputeProposerBalance(ctx context.Context) (sdk.Coin, error) {
	govParams, err := g.QueryGovParams(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	minDeposit := govParams.DepositParams.MinDeposit[0]
	return g.chainCtx.NewCoin(minDeposit.Amount.Add(g.chainCtx.ChainSettings.GasPrice.Mul(sdk.NewDec(int64(submitProposalGas))).Ceil().RoundInt())), nil
}

// UpdateParams goes through proposal process to update parameters.
func (g GovernanceLegacy) UpdateParams(ctx context.Context, t *testing.T, description string, updates []paramproposal.ParamChange) {
	t.Helper()
	// Fund accounts.
	proposer := g.chainCtx.GenAccount()
	proposerBalance, err := g.ComputeProposerBalance(ctx)
	require.NoError(t, err)

	g.faucet.FundAccounts(ctx, t, NewFundedAccount(proposer, proposerBalance))

	g.ProposeAndVote(ctx, t, proposer,
		paramproposal.NewParameterChangeProposal("Updating parameters", description, updates), govtypesv1beta1.OptionYes)
}

// ProposeAndVote create a new proposal, votes from all stakers accounts and awaits for the final status.
func (g GovernanceLegacy) ProposeAndVote(ctx context.Context, t *testing.T, proposer sdk.AccAddress, content govtypesv1beta1.Content, option govtypesv1beta1.VoteOption) {
	t.Helper()
	proposalMsg, err := g.NewMsgSubmitProposalV1Beta1(ctx, proposer, content)
	require.NoError(t, err)

	proposalID, err := g.Propose(ctx, t, proposalMsg)
	require.NoError(t, err)

	proposal, err := g.GetProposal(ctx, proposalID)
	require.NoError(t, err)

	if govtypesv1beta1.StatusVotingPeriod != proposal.Status {
		t.Fatalf("unexpected proposal status after creation: %s", proposal.Status)
	}

	err = g.VoteAll(ctx, option, proposal.ProposalId)
	require.NoError(t, err)

	t.Logf("Voters have voted successfully, waiting for voting period to be finished, votingEndTime:%s", proposal.VotingEndTime)

	finalStatus, err := g.WaitForVotingToFinalize(ctx, proposalID)
	require.NoError(t, err)

	if finalStatus != govtypesv1beta1.StatusPassed {
		t.Fatalf("unexpected proposal status after voting: %s, expected: %s", finalStatus, govtypesv1beta1.StatusPassed)
	}

	t.Logf("Proposal has been submitted, proposalID: %d", proposalID)
}

// Propose creates a new proposal.
func (g GovernanceLegacy) Propose(ctx context.Context, t *testing.T, msg *govtypesv1beta1.MsgSubmitProposal) (uint64, error) {
	SkipUnsafe(t)

	txf := g.chainCtx.TxFactory().WithGas(submitProposalGas)
	result, err := client.BroadcastTx(
		ctx,
		g.chainCtx.ClientContext.WithFromAddress(msg.GetProposer()),
		txf,
		msg,
	)
	if err != nil {
		return 0, err
	}

	proposalID, err := event.FindUint64EventAttribute(result.Events, govtypes.EventTypeSubmitProposal, govtypes.AttributeKeyProposalID)
	if err != nil {
		return 0, err
	}

	return proposalID, nil
}

// NewParamsChangeProposal returns a legacy update parameters proposal.
func (g GovernanceLegacy) NewParamsChangeProposal(
	ctx context.Context,
	t *testing.T,
	proposer sdk.AccAddress,
	title string,
	description string,
	metadata string,
	updates []paramproposal.ParamChange,
) *govtypesv1beta1.MsgSubmitProposal {
	t.Helper()

	legacyContent := paramproposal.NewParameterChangeProposal(
		title,
		description,
		updates,
	)

	proposalMsg, err := g.NewMsgSubmitProposalV1Beta1(
		ctx,
		proposer,
		legacyContent,
	)
	require.NoError(t, err)
	return proposalMsg
}

// NewMsgSubmitProposalV1Beta1 - is a helper which initializes govtypesv1beta1.MsgSubmitProposal with govtypesv1beta1.Content.
func (g GovernanceLegacy) NewMsgSubmitProposalV1Beta1(
	ctx context.Context,
	proposer sdk.AccAddress,
	content govtypesv1beta1.Content,
) (*govtypesv1beta1.MsgSubmitProposal, error) {
	govParams, err := g.QueryGovParams(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := govtypesv1beta1.NewMsgSubmitProposal(content, govParams.DepositParams.MinDeposit, proposer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return msg, nil
}

// NewMsgSubmitProposalV1 - is a helper which initializes govtypesv1.MsgSubmitProposal with govtypesv1beta1.Content.
func (g GovernanceLegacy) NewMsgSubmitProposalV1(
	ctx context.Context,
	proposer sdk.AccAddress,
	content govtypesv1beta1.Content, // That is the single place where we use govtypesv1beta1 in gov.go. Can we avoid it ?
) (*govtypesv1.MsgSubmitProposal, error) {
	msgExecLegacy, err := govtypesv1.NewLegacyContent(content, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	govParams, err := g.QueryGovParams(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgExecLegacy},
		govParams.DepositParams.MinDeposit,
		proposer.String(),
		content.GetDescription(),
		content.GetTitle(),
		content.GetTitle(),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return msg, nil
}

// VoteAll votes for the proposalID from all voting accounts with the provided VoteOption.
func (g GovernanceLegacy) VoteAll(ctx context.Context, option govtypesv1beta1.VoteOption, proposalID uint64) error {
	return g.voteAll(ctx, func(voter sdk.AccAddress) sdk.Msg {
		return &govtypesv1beta1.MsgVote{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Option:     option,
		}
	})
}

// VoteAllWeighted votes for the proposalID from all voting accounts with the provided WeightedVoteOptions.
func (g GovernanceLegacy) VoteAllWeighted(ctx context.Context, options govtypesv1beta1.WeightedVoteOptions, proposalID uint64) error {
	return g.voteAll(ctx, func(voter sdk.AccAddress) sdk.Msg {
		return &govtypesv1beta1.MsgVoteWeighted{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Options:    options,
		}
	})
}

func (g GovernanceLegacy) voteAll(ctx context.Context, msgFunc func(sdk.AccAddress) sdk.Msg) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-g.muCh:
		defer func() {
			g.muCh <- struct{}{}
		}()
	}

	txHashes := make([]string, 0, len(g.stakerAccounts))
	for _, staker := range g.stakerAccounts {
		msg := msgFunc(staker)

		txf := g.chainCtx.TxFactory().
			WithSimulateAndExecute(true)

		clientCtx := g.chainCtx.ClientContext.
			WithBroadcastMode(flags.BroadcastSync)

		res, err := client.BroadcastTx(ctx, clientCtx.WithFromAddress(staker), txf, msg)
		if err != nil {
			return err
		}
		txHashes = append(txHashes, res.TxHash)
	}

	// await for the first error
	for _, txHash := range txHashes {
		_, err := client.AwaitTx(ctx, g.chainCtx.ClientContext, txHash)
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForVotingToFinalize waits for the proposal status to change to final.
// Final statuses are: StatusPassed, StatusRejected or StatusFailed.
func (g GovernanceLegacy) WaitForVotingToFinalize(ctx context.Context, proposalID uint64) (govtypesv1beta1.ProposalStatus, error) {
	proposal, err := g.GetProposal(ctx, proposalID)
	if err != nil {
		return proposal.Status, err
	}

	tmQueryClient := tmservice.NewServiceClient(g.chainCtx.ClientContext)
	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	if err != nil {
		return proposal.Status, errors.WithStack(err)
	}
	if blockRes.Block.Header.Time.Before(proposal.VotingEndTime) { //nolint:staticcheck
		waitCtx, waitCancel := context.WithTimeout(ctx, proposal.VotingEndTime.Sub(blockRes.Block.Header.Time)) //nolint:staticcheck
		defer waitCancel()

		<-waitCtx.Done()
		if ctx.Err() != nil {
			return proposal.Status, ctx.Err()
		}
	}

	retryCtx, retryCancel := context.WithTimeout(ctx, 10*time.Second)
	defer retryCancel()

	err = retry.Do(retryCtx, time.Second, func() error {
		proposal, err = g.GetProposal(ctx, proposalID)
		if err != nil {
			return err
		}

		switch proposal.Status {
		case govtypesv1beta1.StatusPassed, govtypesv1beta1.StatusRejected, govtypesv1beta1.StatusFailed:
			return nil
		default:
			return retry.Retryable(errors.Errorf("waiting for one of final statuses but current one is %s", proposal.Status))
		}
	})
	if err != nil {
		return proposal.Status, err
	}
	return proposal.Status, nil
}

// GetProposal returns proposal by ID.
func (g GovernanceLegacy) GetProposal(ctx context.Context, proposalID uint64) (govtypesv1beta1.Proposal, error) {
	resp, err := g.govClient.Proposal(ctx, &govtypesv1beta1.QueryProposalRequest{
		ProposalId: proposalID,
	})
	if err != nil {
		return govtypesv1beta1.Proposal{}, err
	}

	return resp.Proposal, nil
}

// QueryGovParams returns all governance params.
func (g GovernanceLegacy) QueryGovParams(ctx context.Context) (govtypesv1beta1.Params, error) {
	govClient := g.govClient

	votingParams, err := govClient.Params(ctx, &govtypesv1beta1.QueryParamsRequest{
		ParamsType: govtypesv1beta1.ParamVoting,
	})
	if err != nil {
		return govtypesv1beta1.Params{}, errors.WithStack(err)
	}

	depositParams, err := govClient.Params(ctx, &govtypesv1beta1.QueryParamsRequest{
		ParamsType: govtypesv1beta1.ParamDeposit,
	})
	if err != nil {
		return govtypesv1beta1.Params{}, errors.WithStack(err)
	}

	taillyParams, err := govClient.Params(ctx, &govtypesv1beta1.QueryParamsRequest{
		ParamsType: govtypesv1beta1.ParamTallying,
	})
	if err != nil {
		return govtypesv1beta1.Params{}, errors.WithStack(err)
	}

	return govtypesv1beta1.Params{
		VotingParams:  votingParams.VotingParams,
		DepositParams: depositParams.DepositParams,
		TallyParams:   taillyParams.TallyParams,
	}, nil
}
