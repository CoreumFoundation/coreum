package staking

import (
	"context"
	"time"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

const (
	initialValidatorAmount = 1000000
)

// TestStaking checks validator creation, delegation and undelegation operations work correctly
func TestStaking(ctx context.Context, t testing.T, chain testing.Chain) {
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	delegateAmount := sdk.NewInt(100)
	delegatorInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&stakingtypes.MsgDelegate{}),
		1,
		delegateAmount,
	).Add(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&stakingtypes.MsgUndelegate{}),
		1,
		sdk.NewInt(0),
	))

	// Setup validator and delegator
	delegator := chain.GenAccount()
	validator, deactivateValidator := createValidator(ctx, t, chain)
	defer deactivateValidator()

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(delegator, chain.NewCoin(delegatorInitialBalance)),
	))

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

	// Undelegate coins
	undelegateMsg := stakingtypes.NewMsgUndelegate(delegator, validator, chain.NewCoin(delegateAmount))
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
	resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validator.String(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(initialValidatorAmount), resp.Validator.Tokens.Int64())
}

func createValidator(ctx context.Context, t testing.T, chain testing.Chain) (sdk.ValAddress, func()) {
	stakingClient := stakingtypes.NewQueryClient(chain.ClientContext)

	validatorAmount := sdk.NewInt(initialValidatorAmount)
	validatorInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&stakingtypes.MsgCreateValidator{}),
		1,
		validatorAmount,
	).Add(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&stakingtypes.MsgUndelegate{}),
		1,
		validatorAmount,
	))

	// Create random validator wallet
	validator := chain.GenAccount()
	validatorAddr := sdk.ValAddress(validator)

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(validator, chain.NewCoin(validatorInitialBalance)),
	))

	// Create validator
	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddr,
		cosmosed25519.GenPrivKey().PubKey(),
		chain.NewCoin(validatorAmount),
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
	require.Equal(t, validatorAmount, resp.Validator.Tokens)
	require.Equal(t, stakingtypes.Bonded, resp.Validator.Status)

	return validatorAddr, func() {
		// Undelegate coins, i.e. deactivate validator
		undelegateMsg := stakingtypes.NewMsgUndelegate(validator, validatorAddr, chain.NewCoin(validatorAmount))
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
		Denom:   chain.NetworkConfig.TokenSymbol,
	})
	require.NoError(t, err)

	return *resp.Balance
}
