//go:build integrationtests

package modules

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestStakingProposalParamChange checks that staking param change proposal works correctly.
func TestStakingProposalParamChange(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	resp, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	stakingParams := resp.Params

	targetMaxValidators := 2 * stakingParams.MaxValidators

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	// Create proposition to change max validators value.
	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(ctx, proposer, paramproposal.NewParameterChangeProposal("Change MaxValidators", "Propose changing MaxValidators in the staking module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), strconv.Itoa(int(targetMaxValidators))),
		},
	))
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)
	t.Logf("Proposal has been submitted, proposalID: %d", proposalID)

	// Verify that voting period started.
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// Vote yes from all vote accounts.
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	t.Logf("Voters have voted successfully, waiting for voting period to be finished, votingEndTime:%s", proposal.VotingEndTime)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, finalStatus)

	// Check the proposed change is applied.
	resp, err = stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(targetMaxValidators, resp.Params.MaxValidators)
}

// TestStakingValidatorCRUDAndStaking checks validator creation, delegation and undelegation operations work correctly.
func TestStakingValidatorCRUDAndStaking(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// fastUnbondingTime is the coins unbonding time we use for the test only
	const fastUnbondingTime = time.Second * 10

	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	// we stake the minimum possible staking amount
	validatorStakingAmount := customStakingParams.Params.MinSelfDelegation.Mul(sdk.NewInt(2)) // we multiply not to conflict with the tests which increases the min amount
	// Setup delegator
	delegator := chain.GenAccount()
	delegateAmount := sdk.NewInt(100)
	chain.FundAccountWithOptions(ctx, t, delegator, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&stakingtypes.MsgUndelegate{},
			&stakingtypes.MsgBeginRedelegate{},
			&stakingtypes.MsgEditValidator{},
		},
		Amount: delegateAmount,
	})

	// Setup validator
	validatorAccAddress, validatorAddress, deactivateValidator, err := chain.CreateValidator(ctx, t, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer deactivateValidator()

	// Edit Validator
	updatedDetail := "updated detail"
	editValidatorMsg := &stakingtypes.MsgEditValidator{
		Description:      stakingtypes.Description{Details: updatedDetail},
		ValidatorAddress: validatorAddress.String(),
	}

	chain.FundAccountWithOptions(ctx, t, validatorAccAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{editValidatorMsg},
	})

	editValidatorRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(validatorAccAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(editValidatorMsg)),
		editValidatorMsg,
	)
	require.NoError(t, err)
	assert.EqualValues(t, int64(chain.GasLimitByMsgs(editValidatorMsg)), editValidatorRes.GasUsed)

	valResp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress.String(),
	})

	require.NoError(t, err)
	assert.EqualValues(t, updatedDetail, valResp.GetValidator().Description.Details)

	// Delegate coins
	delegateMsg := stakingtypes.NewMsgDelegate(delegator, validatorAddress, chain.NewCoin(delegateAmount))
	delegateResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	require.NoError(t, err)

	t.Logf("Delegation executed, txHash:%s", delegateResult.TxHash)

	// Make sure coins have been delegated
	ddResp, err := stakingClient.DelegatorDelegations(ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator.String(),
	})
	require.NoError(t, err)
	require.Equal(t, delegateAmount, ddResp.DelegationResponses[0].Balance.Amount)

	// Redelegate Coins
	_, validator2Address, deactivateValidator2, err := chain.CreateValidator(ctx, t, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer deactivateValidator2()

	redelegateMsg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    delegator.String(),
		ValidatorSrcAddress: validatorAddress.String(),
		ValidatorDstAddress: validator2Address.String(),
		Amount:              chain.NewCoin(delegateAmount),
	}

	redelegateResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(redelegateMsg)),
		redelegateMsg,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(chain.GasLimitByMsgs(redelegateMsg)), redelegateResult.GasUsed)
	t.Logf("Redelegation executed, txHash:%s", redelegateResult.TxHash)

	ddResp, err = stakingClient.DelegatorDelegations(ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator.String(),
	})

	require.NoError(t, err)
	assert.Equal(t, delegateAmount, ddResp.DelegationResponses[0].Balance.Amount)
	assert.Equal(t, validator2Address.String(), ddResp.DelegationResponses[0].GetDelegation().ValidatorAddress)

	stakingParams, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	require.NoError(t, err)
	initialUnbondingTime := stakingParams.Params.UnbondingTime

	// defer to restore the time to default after the test
	defer setUnbondingTimeViaGovernance(ctx, t, chain, initialUnbondingTime)
	// change the unbonding time to fast time, to pass the test
	setUnbondingTimeViaGovernance(ctx, t, chain, fastUnbondingTime)

	// Undelegate coins
	undelegateMsg := stakingtypes.NewMsgUndelegate(delegator, validator2Address, chain.NewCoin(delegateAmount))
	undelegateResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(undelegateMsg)),
		undelegateMsg,
	)
	require.NoError(t, err)

	t.Logf("Undelegation executed, txHash:%s ", undelegateResult.TxHash)

	// Wait for undelegation
	time.Sleep(fastUnbondingTime + time.Second*2)

	// Check delegator balance
	delegatorBalance := getBalance(ctx, t, chain, delegator)
	require.GreaterOrEqual(t, delegatorBalance.Amount.Int64(), delegateAmount.Int64())

	// Make sure coins have been undelegated
	valResp, err = stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress.String(),
	})
	require.NoError(t, err)
	require.Equal(t, validatorStakingAmount.String(), valResp.Validator.Tokens.String())
}

// TestValidatorCreationWithLowMinSelfDelegation checks validator can't set the self delegation less than min limit.
func TestValidatorCreationWithLowMinSelfDelegation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)

	initialValidatorAmount := customStakingParams.Params.MinSelfDelegation

	notEnoughValidatorAmount := initialValidatorAmount.Quo(sdk.NewInt(2))

	// Try to create a validator with the amount less than the minimum
	_, _, _, err = chain.CreateValidator(ctx, t, notEnoughValidatorAmount, notEnoughValidatorAmount) //nolint:dogsled // we await for the error only
	require.True(t, stakingtypes.ErrSelfDelegationBelowMinimum.Is(err))
}

// TestValidatorUpdateWithLowMinSelfDelegation checks validator can update its parameters even if the new min self
// delegation is higher than current validator self delegation.
func TestValidatorUpdateWithLowMinSelfDelegation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	initialValidatorAmount := customStakingParams.Params.MinSelfDelegation

	// create new validator with min allowed self delegation
	validatorAccAddress, validatorAddress, deactivateValidator, err := chain.CreateValidator(ctx, t, initialValidatorAmount, initialValidatorAmount)
	require.NoError(t, err)
	defer deactivateValidator()

	customStakingParams, err = customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	minSelfDelegation := customStakingParams.Params.MinSelfDelegation
	// we increase it here to test the update of the validators with the current min self delegation less than new param
	newMinSelfDelegation := minSelfDelegation.Add(sdk.NewInt(1))

	changeMinSelfDelegationCustomParam(ctx, t, chain, customParamsClient, newMinSelfDelegation)
	defer changeMinSelfDelegationCustomParam(ctx, t, chain, customParamsClient, initialValidatorAmount)

	// try to create a validator with the initial amount which we have increased
	_, _, _, err = chain.CreateValidator(ctx, t, initialValidatorAmount, initialValidatorAmount) //nolint:dogsled // we await for the error only
	requireT.ErrorIs(err, stakingtypes.ErrSelfDelegationBelowMinimum)

	// edit validator
	editValidatorMsg := &stakingtypes.MsgEditValidator{
		Description: stakingtypes.Description{
			Details: "updated details",
		},
		ValidatorAddress: validatorAddress.String(),
	}
	chain.FundAccountWithOptions(ctx, t, validatorAccAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{editValidatorMsg},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(validatorAccAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(editValidatorMsg)),
		editValidatorMsg,
	)
	require.NoError(t, err)

	valResp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress.String(),
	})

	require.NoError(t, err)
	assert.EqualValues(t, editValidatorMsg.Description.Details, valResp.GetValidator().Description.Details)
}

func changeMinSelfDelegationCustomParam(
	ctx context.Context,
	t *testing.T,
	chain integrationtests.CoreumChain,
	customParamsClient customparamstypes.QueryClient,
	newMinSelfDelegation sdk.Int,
) {
	requireT := require.New(t)
	// create new proposer
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	marshalledMinSelfDelegation, err := tmjson.Marshal(newMinSelfDelegation)
	requireT.NoError(err)
	// apply proposal
	chain.Governance.ProposeAndVote(ctx, t, proposer,
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

	// check the proposed change is applied
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(newMinSelfDelegation.String(), customStakingParams.Params.MinSelfDelegation.String())
}

func setUnbondingTimeViaGovernance(ctx context.Context, t *testing.T, chain integrationtests.CoreumChain, unbondingTime time.Duration) {
	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	// Create proposition to change max the unbonding time value.
	chain.Governance.ProposeAndVote(ctx, t, proposer,
		paramproposal.NewParameterChangeProposal(
			fmt.Sprintf("Change the unbnunbondingdig time to %s", unbondingTime.String()),
			"Changing unbonding time for the integration test",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyUnbondingTime), fmt.Sprintf("\"%d\"", unbondingTime)),
			},
		),
		govtypes.OptionYes,
	)

	// Check the proposed change is applied.
	stakingParams, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(unbondingTime, stakingParams.Params.UnbondingTime)
}

func getBalance(ctx context.Context, t *testing.T, chain integrationtests.CoreumChain, addr sdk.AccAddress) sdk.Coin {
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	require.NoError(t, err)

	return *resp.Balance
}
