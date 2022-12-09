package staking

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestValidatorCRUDAndStaking checks validator creation, delegation and undelegation operations work correctly.
func TestValidatorCRUDAndStaking(ctx context.Context, t testing.T, chain testing.Chain) {
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
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, delegator, testing.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&stakingtypes.MsgUndelegate{},
			&stakingtypes.MsgBeginRedelegate{},
			&stakingtypes.MsgEditValidator{},
		},
		Amount: delegateAmount,
	}))

	// Setup validator
	validatorAccAddress, validatorAddress, deactivateValidator, err := testing.CreateValidator(ctx, chain, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator()
		require.NoError(t, err)
	}()

	// Edit Validator
	updatedDetail := "updated detail"
	editValidatorMsg := &stakingtypes.MsgEditValidator{
		Description:      stakingtypes.Description{Details: updatedDetail},
		ValidatorAddress: validatorAddress.String(),
	}

	err = chain.Faucet.FundAccountsWithOptions(ctx, validatorAccAddress, testing.BalancesOptions{
		Messages: []sdk.Msg{editValidatorMsg},
	})
	require.NoError(t, err)

	editValidatorRes, err := tx.BroadcastTx(
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
	delegateResult, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Delegation executed", zap.String("txHash", delegateResult.TxHash))

	// Make sure coins have been delegated
	ddResp, err := stakingClient.DelegatorDelegations(ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator.String(),
	})
	require.NoError(t, err)
	require.Equal(t, delegateAmount, ddResp.DelegationResponses[0].Balance.Amount)

	// Redelegate Coins
	_, validator2Address, deactivateValidator2, err := testing.CreateValidator(ctx, chain, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator2()
		require.NoError(t, err)
	}()
	redelegateMsg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    delegator.String(),
		ValidatorSrcAddress: validatorAddress.String(),
		ValidatorDstAddress: validator2Address.String(),
		Amount:              chain.NewCoin(delegateAmount),
	}

	redelegateResult, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(redelegateMsg)),
		redelegateMsg,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(chain.GasLimitByMsgs(redelegateMsg)), redelegateResult.GasUsed)
	logger.Get(ctx).Info("Redelegation executed", zap.String("txHash", redelegateResult.TxHash))

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
	undelegateResult, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(undelegateMsg)),
		undelegateMsg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Undelegation executed", zap.String("txHash", undelegateResult.TxHash))

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
func TestValidatorCreationWithLowMinSelfDelegation(ctx context.Context, t testing.T, chain testing.Chain) {
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)

	initialValidatorAmount := customStakingParams.Params.MinSelfDelegation

	notEnoughValidatorAmount := initialValidatorAmount.Quo(sdk.NewInt(2))

	// Try to create a validator with the amount less than the minimum
	_, _, _, err = testing.CreateValidator(ctx, chain, notEnoughValidatorAmount, notEnoughValidatorAmount) //nolint:dogsled // we await for the error only
	require.True(t, stakingtypes.ErrSelfDelegationBelowMinimum.Is(err))
}

// TestValidatorUpdateWithLowMinSelfDelegation checks validator can update its parameters even if the new min self
// delegation is higher than current validator self delegation.
func TestValidatorUpdateWithLowMinSelfDelegation(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	initialValidatorAmount := customStakingParams.Params.MinSelfDelegation

	// create new validator with min allowed self delegation
	validatorAccAddress, validatorAddress, deactivateValidator, err := testing.CreateValidator(ctx, chain, initialValidatorAmount, initialValidatorAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator()
		require.NoError(t, err)
	}()

	customStakingParams, err = customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	minSelfDelegation := customStakingParams.Params.MinSelfDelegation
	// we increase it here to test the update of the validators with the current min self delegation less than new param
	newMinSelfDelegation := minSelfDelegation.Add(sdk.NewInt(1))

	err = changeMinSelfDelegationCustomParam(ctx, requireT, chain, customParamsClient, newMinSelfDelegation)
	requireT.NoError(err)
	defer func() {
		// return the initial state back
		err = changeMinSelfDelegationCustomParam(ctx, requireT, chain, customParamsClient, initialValidatorAmount)
		require.NoError(t, err)
	}()

	// try to create a validator with the initial amount which we have increased
	_, _, _, err = testing.CreateValidator(ctx, chain, initialValidatorAmount, initialValidatorAmount) //nolint:dogsled // we await for the error only
	require.True(t, stakingtypes.ErrSelfDelegationBelowMinimum.Is(err))

	// edit validator
	editValidatorMsg := &stakingtypes.MsgEditValidator{
		Description: stakingtypes.Description{
			Details: "updated details",
		},
		ValidatorAddress: validatorAddress.String(),
	}
	err = chain.Faucet.FundAccountsWithOptions(ctx, validatorAccAddress, testing.BalancesOptions{
		Messages: []sdk.Msg{editValidatorMsg},
	})
	require.NoError(t, err)

	_, err = tx.BroadcastTx(
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
	requireT *require.Assertions,
	chain testing.Chain,
	customParamsClient customparamstypes.QueryClient,
	newMinSelfDelegation sdk.Int,
) error {
	// create new proposer
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	marshalledMinSelfDelegation, err := tmjson.Marshal(newMinSelfDelegation)
	requireT.NoError(err)
	// apply proposal
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
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(newMinSelfDelegation.String(), customStakingParams.Params.MinSelfDelegation.String())
	return err
}

func setUnbondingTimeViaGovernance(ctx context.Context, t testing.T, chain testing.Chain, unbondingTime time.Duration) {
	requireT := require.New(t)
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	// TODO(dhil) refactor other tests to use that func for the standard propose + vote action.
	// Create proposition to change max the unbonding time value.
	err = chain.Governance.ProposeAndVote(ctx, proposer,
		paramproposal.NewParameterChangeProposal(
			fmt.Sprintf("Change the unbnunbondingdig time to %s", unbondingTime.String()),
			"Changing unbonding time for the integration test",
			[]paramproposal.ParamChange{
				paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyUnbondingTime), fmt.Sprintf("\"%d\"", unbondingTime)),
			},
		),
		govtypes.OptionYes,
	)
	requireT.NoError(err)

	// Check the proposed change is applied.
	stakingParams, err := stakingClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(unbondingTime, stakingParams.Params.UnbondingTime)
}

func getBalance(ctx context.Context, t testing.T, chain testing.Chain, addr sdk.AccAddress) sdk.Coin {
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: addr.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	require.NoError(t, err)

	return *resp.Balance
}
