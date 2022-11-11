package staking

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestValidatorCrudAndStaking checks validator creation, delegation and undelegation operations work correctly.
//
//nolint:funlen // this function is a long test scenario and breaking it down might not be that beneficial
func TestValidatorCrudAndStaking(ctx context.Context, t testing.T, chain testing.Chain) {
	const initialValidatorAmount = 1000000

	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

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
	validatorAccAddress, validatorAddress, deactivateValidator, err := testing.CreateValidator(ctx, chain, sdk.NewInt(initialValidatorAmount))
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
	_, validator2Address, deactivateValidator2, err := testing.CreateValidator(ctx, chain, sdk.NewInt(initialValidatorAmount))
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
	unbondingTime, err := time.ParseDuration(chain.NetworkConfig.StakingConfig.UnbondingTime)
	require.NoError(t, err)
	time.Sleep(unbondingTime + time.Second*2)

	// Check delegator balance
	delegatorBalance := getBalance(ctx, t, chain, delegator)
	require.GreaterOrEqual(t, delegatorBalance.Amount.Int64(), delegateAmount.Int64())

	// Make sure coins have been undelegated
	valResp, err = stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddress.String(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(initialValidatorAmount), valResp.Validator.Tokens.Int64())
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
