package client

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// QueryProposalVotes queries for proposal votes info
func (c Client) QueryProposalVotes(ctx context.Context, proposalID uint64) (map[string][]govtypes.WeightedVoteOption, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// FIXME (wojtek): support pagination
	resp, err := c.govQueryClient.Votes(requestCtx, &govtypes.QueryVotesRequest{ProposalId: proposalID})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	votes := map[string][]govtypes.WeightedVoteOption{}
	for _, v := range resp.Votes {
		votes[v.Voter] = v.Options
	}
	return votes, nil
}

// TxSubmitProposalInput holds input data for PrepareTxSubmitProposal
type TxSubmitProposalInput struct {
	Proposer       types.Wallet
	InitialDeposit types.Coin
	Content        govtypes.Content

	Base tx.BaseInput
}

// PrepareTxSubmitProposal creates a transaction to submit a proposal
func (c Client) PrepareTxSubmitProposal(ctx context.Context, input TxSubmitProposalInput) ([]byte, error) {
	proposerAddress, err := sdk.AccAddressFromBech32(input.Proposer.Key.Address())
	must.OK(err)

	if err = input.InitialDeposit.Validate(); err != nil {
		return nil, errors.Wrap(err, "amount to deposit is invalid")
	}

	msg, err := govtypes.NewMsgSubmitProposal(input.Content, sdk.Coins{
		{
			Denom:  input.InitialDeposit.Denom,
			Amount: sdk.NewIntFromBigInt(input.InitialDeposit.Amount),
		},
	}, proposerAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proposal message")
	}

	signedTx, err := c.Sign(ctx, input.Base, msg)
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}

// TxSubmitProposalVoteInput holds input data for PrepareTxSubmitProposalVote
type TxSubmitProposalVoteInput struct {
	Voter      types.Wallet
	ProposalID uint64
	Option     govtypes.VoteOption

	Base tx.BaseInput
}

// PrepareTxSubmitProposalVote creates a transaction to submit a proposal vote
func (c Client) PrepareTxSubmitProposalVote(ctx context.Context, input TxSubmitProposalVoteInput) ([]byte, error) {
	voterAddress, err := sdk.AccAddressFromBech32(input.Voter.Key.Address())
	must.OK(err)

	msg := govtypes.NewMsgVote(voterAddress, input.ProposalID, input.Option)
	signedTx, err := c.Sign(ctx, input.Base, msg)
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}
