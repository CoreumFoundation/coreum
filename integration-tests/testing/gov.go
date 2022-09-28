package testing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"

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

// NewGovernance initializes the voter accounts to have enough voting power for the voting.
func NewGovernance(chainContext ChainContext, voterMnemonics []string) Governance {
	stakerAccounts := make([]sdk.AccAddress, 0, len(voterMnemonics))
	for _, stakerMnemonic := range voterMnemonics {
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
	govParams, err := g.getParams(ctx)
	if err != nil {
		return sdk.Coin{}, err
	}

	midDeposit := govParams.DepositParams.MinDeposit[0]
	initialGasPrice := NetworkConfig.Fee.FeeModel.Params().InitialGasPrice
	proposerInitialBalance := ComputeNeededBalance(
		initialGasPrice,
		g.chainContext.GasLimitByMsgs(&govtypes.MsgSubmitProposal{}),
		1,
		midDeposit.Amount,
	)

	return g.chainContext.NewCoin(proposerInitialBalance), nil
}

func (g Governance) Propose(ctx context.Context, proposer sdk.AccAddress, content govtypes.Content) (int, error) {
	govParams, err := g.getParams(ctx)
	if err != nil {
		return 0, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, govParams.DepositParams.MinDeposit, proposer)
	if err != nil {
		return 0, err
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
		return 0, err
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
	for _, voter := range g.stakerAccounts {
		msg := &govtypes.MsgVote{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Option:     option,
		}

		txf := g.chainContext.TxFactory()
		txf = txf.WithGas(g.chainContext.GasLimitByMsgs(msg))

		clientCtx := g.chainContext.ClientContext.
			WithBroadcastMode(flags.BroadcastSync)
		res, err := tx.BroadcastTx(
			ctx,
			clientCtx.
				WithFromAddress(voter),
			txf,
			msg,
		)
		if err != nil {
			return err
		}
		txHashes = append(txHashes, res.TxHash)
	}

	// await for the first error
	for _, txHash := range txHashes {
		res, err := tx.AwaitTx(ctx, g.chainContext.ClientContext, txHash)
		if err != nil {
			return err
		}
		if res.TxResult.Code != 0 {
			return errors.Errorf("the vote tx is failed, %s", string(res.TxResult.Data))
		}
	}

	return nil
}

// WaitForProposalStatus wait for the proposal status during the gov VotingPeriod.
func (g Governance) WaitForProposalStatus(ctx context.Context, status govtypes.ProposalStatus, proposalID uint64) (govtypes.Proposal, error) {
	var lastStatus govtypes.ProposalStatus

	govParams, err := g.getParams(ctx)
	if err != nil {
		return govtypes.Proposal{}, err
	}

	timeout := time.NewTimer(govParams.VotingParams.VotingPeriod + time.Second*10)

	ticker := time.NewTicker(time.Millisecond * 250)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return govtypes.Proposal{}, ctx.Err()
		case <-timeout.C:
			return govtypes.Proposal{}, errors.New(fmt.Sprintf("waiting for %s status is timed out for proposal %d and final status %s", status, proposalID, lastStatus))

		default:
			proposal, err := g.getProposal(ctx, proposalID)
			if err != nil {
				return govtypes.Proposal{}, err
			}

			if lastStatus = proposal.Status; lastStatus == status {
				return proposal, nil
			}
		}
	}
	return govtypes.Proposal{}, errors.New(fmt.Sprintf("waiting for %s status is timed out for proposal %d and final status %s", status, proposalID, lastStatus))
}

// GetVotersAccounts returns the configured voting accounts.
func (g Governance) GetVotersAccounts() []sdk.AccAddress {
	return g.stakerAccounts
}

func (g Governance) getParams(ctx context.Context) (govtypes.Params, error) {
	return queryGovParams(ctx, g.govClient)
}

func (g Governance) getProposal(ctx context.Context, proposalID uint64) (govtypes.Proposal, error) {
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
		return govtypes.Params{}, err
	}

	depositParams, err := govClient.Params(ctx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamDeposit,
	})
	if err != nil {
		return govtypes.Params{}, err
	}

	tailyParams, err := govClient.Params(ctx, &govtypes.QueryParamsRequest{
		ParamsType: govtypes.ParamTallying,
	})
	if err != nil {
		return govtypes.Params{}, err
	}

	return govtypes.Params{
		VotingParams:  votingParams.VotingParams,
		DepositParams: depositParams.DepositParams,
		TallyParams:   tailyParams.TallyParams,
	}, nil
}
