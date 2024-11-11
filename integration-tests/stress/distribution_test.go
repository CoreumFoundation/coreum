//go:build integrationtests

package stress

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	customparamstypes "github.com/CoreumFoundation/coreum/v5/x/customparams/types"
)

// TestDistributionWithdrawRewardWithDeterministicGas checks that withdraw reward works correctly and
// gas is deterministic.
func TestDistributionWithdrawRewardWithDeterministicGas(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	delegator := chain.GenAccount()
	delegatorRewardRecipient := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	requireT := require.New(t)
	// the amount of the delegation should be big enough to get at least some reward for the few blocks
	amountToDelegate := sdkmath.NewInt(1_000_000_000)
	chain.FundAccountWithOptions(ctx, t, delegator, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&distributiontypes.MsgDepositValidatorRewardsPool{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
			&distributiontypes.MsgSetWithdrawAddress{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
		},
		Amount: amountToDelegate.Add(sdkmath.NewInt(1_000)),
	})

	delegatedCoin := chain.NewCoin(amountToDelegate)

	// *** Create new validator to use it in the test and capture all required balances. ***
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	// we multiply not to conflict with the tests which increases the min amount
	validatorStakingAmount := customStakingParams.Params.MinSelfDelegation.Mul(sdkmath.NewInt(2))
	validatorStakerAddress, validatorAddress, deactivateValidator, err := chain.CreateValidator(
		ctx, t, validatorStakingAmount, validatorStakingAmount,
	)
	require.NoError(t, err)
	defer deactivateValidator()

	// delegate coins
	delegateMsg := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegatedCoin,
	}

	clientCtx := chain.ClientContext

	t.Log("Delegating some coins to validator to withdraw later")
	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	requireT.NoError(err)

	// deposit in validator rewards pool
	t.Log("Deposit some more amount in the validator rewards pool")
	// withdraw the normal reward
	depositValidatorRewardsPoolMsg := &distributiontypes.MsgDepositValidatorRewardsPool{
		Depositor:        delegator.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1000))),
	}
	txResult, err := client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositValidatorRewardsPoolMsg)),
		depositValidatorRewardsPoolMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(depositValidatorRewardsPoolMsg), uint64(txResult.GasUsed))

	// capture the normal staker balance
	delegatorBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegator.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	delegatorBalanceBeforeWithdrawal := delegatorBalanceRes.Balance

	// capture validator staker balance
	validatorStakerBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: validatorStakerAddress.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	validatorStakerBalanceBeforeWithdrawal := validatorStakerBalanceRes.Balance

	// await next 5 blocks
	requireT.NoError(client.AwaitNextBlocks(ctx, clientCtx, 5))

	// *** Withdraw and check the delegator reward. ***

	// normal delegator
	t.Log("Withdrawing the delegator reward")
	// withdraw the normal reward
	withdrawRewardMsg := &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
	}
	txResult, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawRewardMsg)),
		withdrawRewardMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawRewardMsg), uint64(txResult.GasUsed))

	delegatorBalanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegator.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	delegatorBalanceAfterWithdrawal := delegatorBalanceRes.Balance

	feeSpentOnWithdrawReward := chain.ComputeNeededBalanceFromOptions(integration.BalancesOptions{
		Messages: []sdk.Msg{withdrawRewardMsg},
	})

	delegatorReward := delegatorBalanceAfterWithdrawal.Amount.
		Sub(delegatorBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawReward))
	requireT.True(delegatorReward.IsPositive())
	t.Logf("Withdrawing of the delegator reward is done, amount:%s", delegatorReward.String())

	// *** Change the reward owner and withdraw the delegator reward. ***

	// Change the reward owner and withdraw the reward
	t.Log("Changing the reward recipient and windowing the reward")
	// change withdraw address
	setWithdrawAddressMsg := &distributiontypes.MsgSetWithdrawAddress{
		DelegatorAddress: delegator.String(),
		WithdrawAddress:  delegatorRewardRecipient.String(),
	}
	txResult, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(setWithdrawAddressMsg)),
		setWithdrawAddressMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(setWithdrawAddressMsg), uint64(txResult.GasUsed))
	// withdraw the reward second time
	txResult, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawRewardMsg)),
		withdrawRewardMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawRewardMsg), uint64(txResult.GasUsed))
	delegatorRewardRecipientBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegatorRewardRecipient.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	requireT.True(delegatorRewardRecipientBalanceRes.Balance.IsPositive())

	// *** Withdraw the validator commission. ***

	// validator commission
	t.Log("Withdrawing the validator commission")
	// withdraw the normal reward
	withdrawCommissionMsg := &distributiontypes.MsgWithdrawValidatorCommission{
		ValidatorAddress: validatorAddress.String(),
	}

	chain.FundAccountWithOptions(ctx, t, validatorStakerAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})

	txResult, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(validatorStakerAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawCommissionMsg)),
		withdrawCommissionMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawCommissionMsg), uint64(txResult.GasUsed))

	validatorStakerBalanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: validatorStakerAddress.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	validatorStakerBalanceAfterWithdrawal := validatorStakerBalanceRes.Balance

	feeSpentOnWithdrawCommission := chain.ComputeNeededBalanceFromOptions(integration.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})

	validatorStakerCommissionReward := validatorStakerBalanceAfterWithdrawal.Amount.
		Sub(validatorStakerBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawCommission))
	requireT.True(validatorStakerCommissionReward.IsPositive())
	t.Logf("Withdrawing of the validator commission is done, amount:%s", validatorStakerCommissionReward.String())
}
