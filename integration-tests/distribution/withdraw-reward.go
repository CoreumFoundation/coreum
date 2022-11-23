package distribution

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestWithdrawRewardWithDeterministicGas checks that withdraw reward works correctly and gas is deterministic.
func TestWithdrawRewardWithDeterministicGas(ctx context.Context, t testing.T, chain testing.Chain) {
	delegator := chain.GenAccount()
	delegatorRewardRecipient := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	requireT := require.New(t)
	// the amount of the delegation should be big enough to get at least some reward for the few blocks
	amountToDelegate := sdk.NewInt(1_000_000)
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, delegator, testing.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
			&distributiontypes.MsgSetWithdrawAddress{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
		},
		Amount: amountToDelegate,
	}))

	delegatedCoin := chain.NewCoin(amountToDelegate)

	// *** Create new validator to use it in the test and capture all required balances. ***
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	validatorStakingAmount := customStakingParams.Params.MinSelfDelegation
	validatorStakerAddress, validatorAddress, deactivateValidator, err := testing.CreateValidator(ctx, chain, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator()
		require.NoError(t, err)
	}()

	// delegate coins
	delegateMsg := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegatedCoin,
	}

	clientCtx := chain.ClientContext

	logger.Get(ctx).Info("Delegating some coins to validator to withdraw later")
	_, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	requireT.NoError(err)

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
	requireT.NoError(tx.AwaitNextBlocks(ctx, clientCtx, 5))

	// *** Withdraw and check the delegator reward. ***

	// normal delegator
	logger.Get(ctx).Info("Withdrawing the delegator reward")
	// withdraw the normal reward
	withdrawRewardMsg := &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
	}
	txResult, err := tx.BroadcastTx(
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

	feeSpentOnWithdrawReward := chain.ComputeNeededBalanceFromOptions(testing.BalancesOptions{
		Messages: []sdk.Msg{withdrawRewardMsg},
	})

	delegatorReward := delegatorBalanceAfterWithdrawal.Amount.Sub(delegatorBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawReward))
	requireT.True(delegatorReward.IsPositive())
	logger.Get(ctx).Info("Withdrawing of the delegator reward is done", zap.String("amount", delegatorReward.String()))

	// *** Change the reward owner and withdraw the delegator reward. ***

	// Change the reward owner and withdraw the reward
	logger.Get(ctx).Info("Changing the reward recipient and windowing the reward")
	// change withdraw address
	setWithdrawAddressMsg := &distributiontypes.MsgSetWithdrawAddress{
		DelegatorAddress: delegator.String(),
		WithdrawAddress:  delegatorRewardRecipient.String(),
	}
	txResult, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(setWithdrawAddressMsg)),
		setWithdrawAddressMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(setWithdrawAddressMsg), uint64(txResult.GasUsed))
	// withdraw the reward second time
	txResult, err = tx.BroadcastTx(
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
	logger.Get(ctx).Info("Withdrawing the validator commission")
	// withdraw the normal reward
	withdrawCommissionMsg := &distributiontypes.MsgWithdrawValidatorCommission{
		ValidatorAddress: validatorAddress.String(),
	}

	err = chain.Faucet.FundAccountsWithOptions(ctx, validatorStakerAddress, testing.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})
	requireT.NoError(err)

	txResult, err = tx.BroadcastTx(
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

	feeSpentOnWithdrawCommission := chain.ComputeNeededBalanceFromOptions(testing.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})

	validatorStakerCommissionReward := validatorStakerBalanceAfterWithdrawal.Amount.Sub(validatorStakerBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawCommission))
	requireT.True(validatorStakerCommissionReward.IsPositive())
	logger.Get(ctx).Info("Withdrawing of the validator commission is done", zap.String("amount", validatorStakerCommissionReward.String()))
}
