package integrationtests

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
)

// Governance keep the test chain predefined account for the governance operations.
type Governance struct {
	chainCtx       ChainContext
	govClient      govtypes.QueryClient
	stakerAccounts []sdk.AccAddress
	muCh           chan struct{}
}

// NewGovernance returns the new instance of Governance.
func NewGovernance(chainCtx ChainContext, stakerMnemonics []string) Governance {
	stakerAccounts := make([]sdk.AccAddress, 0, len(stakerMnemonics))
	for _, stakerMnemonic := range stakerMnemonics {
		stakerAccounts = append(stakerAccounts, chainCtx.ImportMnemonic(stakerMnemonic))
	}

	gov := Governance{
		chainCtx:       chainCtx,
		stakerAccounts: stakerAccounts,
		govClient:      govtypes.NewQueryClient(chainCtx.ClientContext),
		muCh:           make(chan struct{}, 1),
	}
	gov.muCh <- struct{}{}

	return gov
}

// ComputeProposerBalance computes the balance required for the proposer.
func (g Governance) ComputeProposerBalance(ctx context.Context) (sdk.Coin, error) {
	govParams, err := g.queryGovParams(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	minDeposit := govParams.DepositParams.MinDeposit[0]
	proposerInitialBalance := g.chainCtx.ComputeNeededBalanceFromOptions(BalancesOptions{
		Messages: []sdk.Msg{&govtypes.MsgSubmitProposal{}},
		Amount:   minDeposit.Amount,
	})

	return g.chainCtx.NewCoin(proposerInitialBalance), nil
}

// UpdateParams goes through proposal process to update parameters
func (g Governance) UpdateParams(ctx context.Context, description string, updates []paramproposal.ParamChange) error {
	// Fund accounts.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	if err != nil {
		return err
	}

	err = chain.Faucet.FundAccounts(ctx, NewFundedAccount(proposer, proposerBalance))
	if err != nil {
		return err
	}

	err = g.ProposeAndVote(ctx, proposer,
		paramproposal.NewParameterChangeProposal("Updating parameters", description, updates), govtypes.OptionYes)
	if err != nil {
		return err
	}

	return nil
}

// ProposeAndVote create a new proposal, votes from all stakers accounts and awaits for the final status.
func (g Governance) ProposeAndVote(ctx context.Context, proposer sdk.AccAddress, content govtypes.Content, option govtypes.VoteOption) error {
	proposalMsg, err := g.NewMsgSubmitProposal(ctx, proposer, content)
	if err != nil {
		return err
	}

	proposalID, err := g.Propose(ctx, proposalMsg)
	if err != nil {
		return err
	}

	proposal, err := g.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	if govtypes.StatusVotingPeriod != proposal.Status {
		return errors.Errorf("unexpected proposal status after creation: %s", proposal.Status)
	}

	err = g.VoteAll(ctx, option, proposal.ProposalId)
	if err != nil {
		return err
	}
	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	finalStatus, err := g.WaitForVotingToFinalize(ctx, proposalID)
	if err != nil {
		return err
	}
	if finalStatus != govtypes.StatusPassed {
		return errors.Errorf("unexpected proposal status after voting: %s, expected: %s", finalStatus, govtypes.StatusPassed)
	}

	logger.Get(ctx).Info("Proposal has been submitted", zap.Uint64("proposalID", proposalID))

	return nil
}

// Propose creates a new proposal.
func (g Governance) Propose(ctx context.Context, msg *govtypes.MsgSubmitProposal) (uint64, error) {
	txf := g.chainCtx.TxFactory().
		WithGas(g.chainCtx.GasLimitByMsgs(&govtypes.MsgSubmitProposal{}))
	result, err := tx.BroadcastTx(
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

// NewMsgSubmitProposal - is a helper which initializes MsgSubmitProposal with args passed and prefills min deposit.
func (g Governance) NewMsgSubmitProposal(ctx context.Context, proposer sdk.AccAddress, content govtypes.Content) (*govtypes.MsgSubmitProposal, error) {
	govParams, err := g.queryGovParams(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, govParams.DepositParams.MinDeposit, proposer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return msg, nil
}

// VoteAll votes for the proposalID from all voting accounts with the provided VoteOption.
func (g Governance) VoteAll(ctx context.Context, option govtypes.VoteOption, proposalID uint64) error {
	return g.voteAll(ctx, func(voter sdk.AccAddress) sdk.Msg {
		return &govtypes.MsgVote{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Option:     option,
		}
	})
}

// VoteAllWeighted votes for the proposalID from all voting accounts with the provided WeightedVoteOptions.
func (g Governance) VoteAllWeighted(ctx context.Context, options govtypes.WeightedVoteOptions, proposalID uint64) error {
	return g.voteAll(ctx, func(voter sdk.AccAddress) sdk.Msg {
		return &govtypes.MsgVoteWeighted{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Options:    options,
		}
	})
}

func (g Governance) voteAll(ctx context.Context, msgFunc func(sdk.AccAddress) sdk.Msg) error {
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
			WithGas(g.chainCtx.GasLimitByMsgs(msg))

		clientCtx := g.chainCtx.ClientContext.
			WithBroadcastMode(flags.BroadcastSync)

		res, err := tx.BroadcastTx(ctx, clientCtx.WithFromAddress(staker), txf, msg)
		if err != nil {
			return err
		}
		txHashes = append(txHashes, res.TxHash)
	}

	// await for the first error
	for _, txHash := range txHashes {
		_, err := tx.AwaitTx(ctx, g.chainCtx.ClientContext, txHash)
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForVotingToFinalize waits for the proposal status to change to final.
// Final statuses are: StatusPassed, StatusRejected or StatusFailed.
func (g Governance) WaitForVotingToFinalize(ctx context.Context, proposalID uint64) (govtypes.ProposalStatus, error) {
	proposal, err := g.GetProposal(ctx, proposalID)
	if err != nil {
		return proposal.Status, err
	}

	block, err := g.chainCtx.ClientContext.Client().Block(ctx, nil)
	if err != nil {
		return proposal.Status, errors.WithStack(err)
	}
	if block.Block.Time.Before(proposal.VotingEndTime) {
		waitCtx, waitCancel := context.WithTimeout(ctx, proposal.VotingEndTime.Sub(block.Block.Time))
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
		case govtypes.StatusPassed, govtypes.StatusRejected, govtypes.StatusFailed:
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

// GetProposal returns proposal by ID
func (g Governance) GetProposal(ctx context.Context, proposalID uint64) (govtypes.Proposal, error) {
	resp, err := g.govClient.Proposal(ctx, &govtypes.QueryProposalRequest{
		ProposalId: proposalID,
	})
	if err != nil {
		return govtypes.Proposal{}, err
	}

	return resp.Proposal, nil
}

func (g Governance) queryGovParams(ctx context.Context) (govtypes.Params, error) {
	govClient := g.govClient

	votingParams, err := govClient.Params(ctx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamVoting,
	})
	if err != nil {
		return govtypes.Params{}, errors.WithStack(err)
	}

	depositParams, err := govClient.Params(ctx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamDeposit,
	})
	if err != nil {
		return govtypes.Params{}, errors.WithStack(err)
	}

	taillyParams, err := govClient.Params(ctx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamTallying,
	})
	if err != nil {
		return govtypes.Params{}, errors.WithStack(err)
	}

	return govtypes.Params{
		VotingParams:  votingParams.VotingParams,
		DepositParams: depositParams.DepositParams,
		TallyParams:   taillyParams.TallyParams,
	}, nil
}
