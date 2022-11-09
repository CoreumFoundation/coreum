package staking

import (
	"context"
	"time"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

// TestStaking checks validator creation, delegation and undelegation operations work correctly.
//
//nolint:funlen // this function is a long test scenario and breaking it down might not be that beneficial
func TestStaking(ctx context.Context, t testing.T, chain testing.Chain) {
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
	validator, deactivateValidator := createValidator(ctx, t, chain, sdk.NewInt(initialValidatorAmount))
	defer deactivateValidator()

	// Edit Validator
	updatedDetail := "updated detail"
	editValidatorMsg := &stakingtypes.MsgEditValidator{
		Description:      stakingtypes.Description{Details: updatedDetail},
		ValidatorAddress: validator.String(),
	}

	editValidatorRes, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sdk.AccAddress(validator)),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(editValidatorMsg)),
		editValidatorMsg,
	)
	require.NoError(t, err)
	assert.EqualValues(t, int64(chain.GasLimitByMsgs(editValidatorMsg)), editValidatorRes.GasUsed)

	valResp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator.String(),
	})

	require.NoError(t, err)
	assert.EqualValues(t, updatedDetail, valResp.GetValidator().Description.Details)

	// Delegate coins
	delegateMsg := stakingtypes.NewMsgDelegate(delegator, validator, chain.NewCoin(delegateAmount))
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
	validator2, deactivateValidator2 := createValidator(ctx, t, chain, sdk.NewInt(initialValidatorAmount))
	defer deactivateValidator2()
	redelegateMsg := &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    delegator.String(),
		ValidatorSrcAddress: validator.String(),
		ValidatorDstAddress: validator2.String(),
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
	assert.Equal(t, validator2.String(), ddResp.DelegationResponses[0].GetDelegation().ValidatorAddress)

	// Undelegate coins
	undelegateMsg := stakingtypes.NewMsgUndelegate(delegator, validator2, chain.NewCoin(delegateAmount))
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
		ValidatorAddr: validator.String(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(initialValidatorAmount), valResp.Validator.Tokens.Int64())
}

func createValidator(ctx context.Context, t testing.T, chain testing.Chain, initialAmount sdk.Int) (sdk.ValAddress, func()) {
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)
	validator := chain.GenAccount()

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, validator, testing.BalancesOptions{
		Messages: []sdk.Msg{&stakingtypes.MsgCreateValidator{}, &stakingtypes.MsgUndelegate{}},
		Amount:   initialAmount.MulRaw(2),
	}))

	// Create validator
	validatorAddr := sdk.ValAddress(validator)
	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddr,
		cosmosed25519.GenPrivKey().PubKey(),
		chain.NewCoin(initialAmount),
		stakingtypes.Description{Moniker: "TestCreateValidator"},
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.OneInt(),
	)
	require.NoError(t, err)
	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(validator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Validator creation executed", zap.String("txHash", result.TxHash))

	// Make sure validator has been created
	resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddr.String(),
	})
	require.NoError(t, err)
	require.Equal(t, initialAmount, resp.Validator.Tokens)
	require.Equal(t, stakingtypes.Bonded, resp.Validator.Status)

	return validatorAddr, func() {
		// Undelegate coins, i.e. deactivate validator
		undelegateMsg := stakingtypes.NewMsgUndelegate(validator, validatorAddr, chain.NewCoin(initialAmount))
		_, err = tx.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(validator),
			chain.TxFactory().WithGas(chain.GasLimitByMsgs(undelegateMsg)),
			undelegateMsg,
		)
		require.NoError(t, err)
	}
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
