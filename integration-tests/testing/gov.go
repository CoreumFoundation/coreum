package testing

import (
	"context"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Governance keep the test chain predefined account for the governance operations.
type Governance struct {
	chainCtx       ChainContext
	govClient      govtypes.QueryClient
	votersAccounts []sdk.AccAddress
}

// NewGovernance initializes the voter accounts to have enough voting power for the voting.
func NewGovernance(
	ctx context.Context,
	chainCtx ChainContext,
	faucet *Faucet,
) (*Governance, error) {
	const (
		govVotersNumber      = 2
		delegationMultiplier = "1.02"
	)
	delegationMultiplierDec, err := sdk.NewDecFromStr(delegationMultiplier)
	if err != nil {
		return nil, err
	}

	clientCtx := chainCtx.ClientContext
	networkCfg := chainCtx.NetworkConfig

	govClient := govtypes.NewQueryClient(clientCtx)
	stakingClient := stakingtypes.NewQueryClient(clientCtx)

	// add voters to keyring
	votersAccounts := make([]sdk.AccAddress, 0, govVotersNumber)
	for i := 0; i < govVotersNumber; i++ {
		votersAccounts = append(votersAccounts, chainCtx.RandomWallet())
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
		wallet := chainCtx.AccAddressToLegacyWallet(voter)
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
	delegateCoin := chainCtx.NewCoin(voterDelegateAmount)

	txf := chainCtx.TxFactory()
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

	return &Governance{
		chainCtx:       chainCtx,
		votersAccounts: votersAccounts,
		govClient:      govClient,
	}, nil
}

func (g *Governance) VoteAll(ctx context.Context, option govtypes.VoteOption, proposalID uint64) error {
	txf := g.chainCtx.TxFactory()
	txf = txf.WithGas(uint64(g.chainCtx.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas))
	for _, voter := range g.votersAccounts {
		msg := &govtypes.MsgVote{
			ProposalId: proposalID,
			Voter:      voter.String(),
			Option:     option,
		}

		_, err := tx.BroadcastTx(
			ctx,
			g.chainCtx.ClientContext.
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

func (g *Governance) GetParams(ctx context.Context) (govtypes.Params, error) {
	return queryGovParams(ctx, g.govClient)
}

func (g *Governance) GetVotersAccounts() []sdk.AccAddress {
	return g.votersAccounts
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
