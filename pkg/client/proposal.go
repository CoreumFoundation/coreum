package client

import (
	"context"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// GetProposal returns proposal by the given ID
func (c Client) GetProposal(ctx context.Context, proposalID uint64) (*govtypes.Proposal, error) {
	resp, err := c.govQueryClient.Proposal(ctx, &govtypes.QueryProposalRequest{
		ProposalId: proposalID,
	})
	must.OK(err)

	return &resp.Proposal, nil
}

// GetProposalByTx returns proposal ID by the given transaction hash
func (c Client) GetProposalByTx(ctx context.Context, tx string) (*govtypes.Proposal, error) {
	txHashBytes, err := hex.DecodeString(tx)
	must.OK(err)

	txData, err := c.clientCtx.Client.Tx(ctx, txHashBytes, false)
	must.OK(err)

	var proposalID uint64
	for _, event := range txData.TxResult.Events {
		if event.Type != "submit_proposal" {
			continue
		}

		if len(event.Attributes) == 0 {
			continue
		}

		if string(event.Attributes[0].GetKey()) != "proposal_id" {
			continue
		}

		id, err := strconv.Atoi(string(event.Attributes[0].GetValue()))
		must.OK(err)

		proposalID = uint64(id)
	}

	if proposalID == 0 {
		return nil, errors.New("no proposal event found for the given transaction")
	}

	resp, err := c.govQueryClient.Proposal(ctx, &govtypes.QueryProposalRequest{
		ProposalId: proposalID,
	})
	must.OK(err)

	return &resp.Proposal, nil
}

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

// TxSubmitProposalDepositInput holds input data for PrepareTxSubmitProposalDeposit
type TxSubmitProposalDepositInput struct {
	Depositor  types.Wallet
	ProposalID uint64
	Amount     types.Coin

	Base tx.BaseInput
}

// PrepareTxSubmitProposalDeposit creates a transaction to submit a proposal deposit
func (c Client) PrepareTxSubmitProposalDeposit(ctx context.Context, input TxSubmitProposalDepositInput) ([]byte, error) {
	depositorAddress, err := sdk.AccAddressFromBech32(input.Depositor.Key.Address())
	must.OK(err)

	msg := govtypes.NewMsgDeposit(depositorAddress, input.ProposalID, sdk.Coins{
		{
			Denom:  input.Amount.Denom,
			Amount: sdk.NewIntFromBigInt(input.Amount.Amount),
		},
	})
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
