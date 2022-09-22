package testing

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// Governance keep the test chain predefined account for the governance operations.
type Governance struct {
	chainContext  *ChainContext
	govClient     govtypes.QueryClient
	faucet        *Faucet
	voterAccounts []sdk.AccAddress
}

// NewGovernance initializes the voter accounts to have enough voting power for the voting.
func NewGovernance( //nolint:funlen // The test covers step-by step use case
	ctx context.Context,
	chainContext *ChainContext,
	faucet *Faucet,
) (*Governance, error) {
	const (
		govVotersNumber      = 2
		delegationMultiplier = "1.02"
	)

	log := logger.Get(ctx)
	log.Info("Initialising the governance accounts")

	delegationMultiplierDec, err := sdk.NewDecFromStr(delegationMultiplier)
	if err != nil {
		return nil, err
	}

	clientCtx := chainContext.ClientContext
	networkCfg := chainContext.NetworkConfig

	govClient := govtypes.NewQueryClient(clientCtx)
	stakingClient := stakingtypes.NewQueryClient(clientCtx)

	// add voters to keyring
	votersAccounts := make([]sdk.AccAddress, 0, govVotersNumber)
	for i := 0; i < govVotersNumber; i++ {
		votersAccounts = append(votersAccounts, chainContext.RandomWallet())
	}

	govParams, err := queryGovParams(ctx, govClient)
	if err != nil {
		return nil, err
	}

	stakingPool, err := stakingClient.Pool(ctx, &stakingtypes.QueryPoolRequest{})
	if err != nil {
		return nil, err
	}

	// compute needed balance for voters and add fund them

	voterDelegateAmount := stakingPool.Pool.BondedTokens.ToDec().
		Mul(govParams.TallyParams.Threshold.Mul(delegationMultiplierDec)).
		QuoInt64(int64(len(votersAccounts))).RoundInt()

	voterInitialBalance := ComputeNeededBalance(
		networkCfg.Fee.FeeModel.Params().InitialGasPrice,
		uint64(networkCfg.Fee.FeeModel.Params().MaxBlockGas),
		3,
		voterDelegateAmount,
	)

	fundedAccounts := make([]FundedAccount, 0, len(votersAccounts))
	for _, voter := range votersAccounts {
		wallet := chainContext.AccAddressToLegacyWallet(voter)
		fundedAccounts = append(fundedAccounts, NewFundedAccount(wallet, sdk.NewCoin(networkCfg.TokenSymbol, voterInitialBalance)))
	}

	err = faucet.FundAccounts(ctx, fundedAccounts...)
	if err != nil {
		return nil, err
	}

	// Delegate voter coins for the voters

	validators, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{})
	if err != nil {
		return nil, err
	}
	valAddress, err := sdk.ValAddressFromBech32(validators.Validators[0].OperatorAddress)
	if err != nil {
		return nil, err
	}

	delegateCoin := chainContext.NewCoin(voterDelegateAmount)

	txf := chainContext.TxFactory()
	txf = txf.WithGas(uint64(networkCfg.Fee.FeeModel.Params().MaxBlockGas))
	for _, voter := range votersAccounts {
		msg := &stakingtypes.MsgDelegate{
			DelegatorAddress: voter.String(),
			ValidatorAddress: valAddress.String(),
			Amount:           delegateCoin,
		}

		_, err := tx.BroadcastTx(
			ctx,
			clientCtx.
				WithFromName(voter.String()).
				WithFromAddress(voter),
			txf,
			msg,
		)
		if err != nil {
			return nil, err
		}
	}

	log.Info("Initialisation of the governance accounts is done")

	return &Governance{
		chainContext:  chainContext,
		faucet:        faucet,
		voterAccounts: votersAccounts,
		govClient:     govClient,
	}, nil
}

// CreateProposer creates a new proposed and funds it with enough tokens for the proposal.
func (g *Governance) CreateProposer(ctx context.Context) (sdk.AccAddress, error) {
	proposer := g.chainContext.RandomWallet()
	govParams, err := g.getParams(ctx)
	if err != nil {
		return nil, err
	}

	proposerInitialBalance := ComputeNeededBalance(
		g.chainContext.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(g.chainContext.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		govParams.DepositParams.MinDeposit[0].Amount,
	)

	err = g.faucet.FundAccounts(ctx, NewFundedAccount(g.chainContext.AccAddressToLegacyWallet(proposer), g.chainContext.NewCoin(proposerInitialBalance)))
	if err != nil {
		return nil, err
	}

	return proposer, nil
}

func (g *Governance) Propose(ctx context.Context, proposer sdk.AccAddress, content govtypes.Content) (int, error) {
	govParams, err := g.getParams(ctx)
	if err != nil {
		return 0, err
	}

	msg, err := govtypes.NewMsgSubmitProposal(content, govParams.DepositParams.MinDeposit, proposer)
	if err != nil {
		return 0, err
	}

	txf := g.chainContext.TxFactory()
	txf = txf.WithGas(uint64(g.chainContext.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas))
	result, err := tx.BroadcastTx(
		ctx,
		g.chainContext.ClientContext.
			WithFromName(proposer.String()).
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
func (g *Governance) VoteAll(ctx context.Context, option govtypes.VoteOption, proposalID uint64) error {
	txf := g.chainContext.TxFactory()
	txf = txf.WithGas(uint64(g.chainContext.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas))
	for _, voter := range g.voterAccounts {
		msg := &govtypes.MsgVote{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Option:     option,
		}

		_, err := tx.BroadcastTx(
			ctx,
			g.chainContext.ClientContext.
				WithFromName(voter.String()).
				WithFromAddress(voter),
			txf,
			msg,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForProposalStatus wait for the proposal status during the gov VotingPeriod.
func (g *Governance) WaitForProposalStatus(ctx context.Context, status govtypes.ProposalStatus, proposalID uint64) (govtypes.Proposal, error) {
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
func (g *Governance) GetVotersAccounts() []sdk.AccAddress {
	return g.voterAccounts
}

func (g *Governance) getParams(ctx context.Context) (govtypes.Params, error) {
	return queryGovParams(ctx, g.govClient)
}

func (g *Governance) getProposal(ctx context.Context, proposalID uint64) (govtypes.Proposal, error) {
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
