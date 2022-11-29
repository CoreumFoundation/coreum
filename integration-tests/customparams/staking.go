package customparams

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestStakingProposalParamChange checks that customparams param change proposal works correctly for staking params.
func TestStakingProposalParamChange(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	// create new proposer
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	minSelfDelegation := customStakingParams.Params.MinSelfDelegation
	// we decrease it here in order not to conflict with the tests which create the validators
	newMinSelfDelegation := minSelfDelegation.Sub(sdk.NewInt(1))

	marshalledMinSelfDelegation, err := tmjson.Marshal(newMinSelfDelegation)
	requireT.NoError(err)
	err = chain.Governance.ProposeAndVote(ctx, proposer,
		paramproposal.NewParameterChangeProposal(
			"Custom staking params change proposal", "-",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange(
					customparamstypes.CustomParamsStaking, string(customparamstypes.ParamStoreKeyMinSelfDelegation), string(marshalledMinSelfDelegation),
				),
			},
		),
		govtypes.OptionYes,
	)
	requireT.NoError(err)

	// check the proposed change is applied
	customStakingParams, err = customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(newMinSelfDelegation.String(), customStakingParams.Params.MinSelfDelegation.String())
}
