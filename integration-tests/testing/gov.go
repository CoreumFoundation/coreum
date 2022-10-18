package testing

import (
	"context"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// Governance keep the test chain predefined account for the governance operations.
type Governance struct {
	chainContext   ChainContext
	govClient      govtypes.QueryClient
	stakerAccounts []sdk.AccAddress
	muCh           chan struct{}
}

// NewGovernance returns the new instance of Governance.
func NewGovernance(chainContext ChainContext, stakerMnemonics []string) Governance {
	stakerAccounts := make([]sdk.AccAddress, 0, len(stakerMnemonics))
	for _, stakerMnemonic := range stakerMnemonics {
		stakerAccounts = append(stakerAccounts, chainContext.ImportMnemonic(stakerMnemonic))
	}

	gov := Governance{
		chainContext:   chainContext,
		stakerAccounts: stakerAccounts,
		govClient:      govtypes.NewQueryClient(chainContext.ClientContext),
		muCh:           make(chan struct{}, 1),
	}
	gov.muCh <- struct{}{}

	return gov
}

// ComputeProposerBalance computes the balance required for the proposer.
func (g Governance) ComputeProposerBalance(ctx context.Context) (sdk.Coin, error) {
	govParams, err := queryGovParams(ctx, g.govClient)
	if err != nil {
		return sdk.Coin{}, err
	}

	minDeposit := govParams.DepositParams.MinDeposit[0]
	initialGasPrice := NetworkConfig.Fee.FeeModel.Params().InitialGasPrice
	proposerInitialBalance := ComputeNeededBalance(
		initialGasPrice,
		g.chainContext.GasLimitByMsgs(&govtypes.MsgSubmitProposal{}),
		1,
		minDeposit.Amount,
	)

	return g.chainContext.NewCoin(proposerInitialBalance), nil
}

// Propose creates a new proposal.
func (g Governance) Propose(ctx context.Context, proposer sdk.AccAddress, content govtypes.Content) (int, error) {
	govParams, err := queryGovParams(ctx, g.govClient)
	if err != nil {
		return 0, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, govParams.DepositParams.MinDeposit, proposer)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	txf := g.chainContext.TxFactory().
		WithGas(g.chainContext.GasLimitByMsgs(&govtypes.MsgSubmitProposal{}))
	result, err := tx.BroadcastTx(
		ctx,
		g.chainContext.ClientContext.
			WithFromAddress(proposer),
		txf,
		msg,
	)
	if err != nil {
		return 0, err
	}

	proposalIDStr, ok := client.FindEventAttribute(sdk.StringifyEvents(result.Events), govtypes.EventTypeSubmitProposal, govtypes.AttributeKeyProposalID)
	if !ok {
		return 0, errors.New("can find proposal id in the broadcast response")
	}
	proposalID, err := strconv.Atoi(proposalIDStr)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return proposalID, nil
}

// VoteAll votes for the proposalID from all voting accounts with the provided VoteOption.
func (g Governance) VoteAll(ctx context.Context, option govtypes.VoteOption, proposalID uint64) error {
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
		msg := &govtypes.MsgVote{
			ProposalId: proposalID,
			Voter:      staker.String(),
			Option:     option,
		}

		txf := g.chainContext.TxFactory().
			WithGas(g.chainContext.GasLimitByMsgs(msg))

		clientCtx := g.chainContext.ClientContext.
			WithBroadcastMode(flags.BroadcastSync)

		res, err := tx.BroadcastTx(ctx, clientCtx.WithFromAddress(staker), txf, msg)
		if err != nil {
			return err
		}
		txHashes = append(txHashes, res.TxHash)
	}

	// await for the first error
	for _, txHash := range txHashes {
		_, err := tx.AwaitTx(ctx, g.chainContext.ClientContext, txHash)
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForVotingToPass waits for the voting to pass.
func (g Governance) WaitForVotingToPass(ctx context.Context, proposalID uint64) error {
	proposal, err := g.GetProposal(ctx, proposalID)
	if err != nil {
		return err
	}

	block, err := g.chainContext.ClientContext.Client().Block(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}
	if block.Block.Time.Before(proposal.VotingEndTime) {
		waitCtx, waitCancel := context.WithTimeout(ctx, proposal.VotingEndTime.Sub(block.Block.Time))
		defer waitCancel()

		<-waitCtx.Done()
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	retryCtx, retryCancel := context.WithTimeout(ctx, 10*time.Second)
	defer retryCancel()

	err = retry.Do(retryCtx, time.Second, func() error {
		proposal, err := g.GetProposal(ctx, proposalID)
		if err != nil {
			return err
		}

		switch proposal.Status {
		case govtypes.StatusPassed:
			return nil
		case govtypes.StatusFailed:
			return errors.New("voting failed")
		default:
			return retry.Retryable(errors.Errorf("waiting for status %s but current one is %s", govtypes.StatusPassed, proposal.Status))
		}
	})
	if err != nil {
		return err
	}
	return nil
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

func queryGovParams(ctx context.Context, govClient govtypes.QueryClient) (govtypes.Params, error) {
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
