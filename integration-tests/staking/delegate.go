package staking

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestDelegate checks that delegation and undelegation works correctly
func TestDelegate(ctx context.Context, t testing.T, chain testing.Chain) {
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
		delegateAmount,
	))

	// Create random delegator wallet
	delegator := chain.RandomWallet()

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(delegator, chain.NewCoin(delegatorInitialBalance)),
	))

	// Fetch existing validator
	validatorsResp, err := stakingClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
	require.NoError(t, err)
	require.NotEmpty(t, validatorsResp.Validators)

	valAddress, err := sdk.ValAddressFromBech32(validatorsResp.Validators[0].OperatorAddress)
	require.NoError(t, err)

	// Delegate coins
	delegateMsg := stakingtypes.NewMsgDelegate(delegator, valAddress, chain.NewCoin(delegateAmount))
	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Delegation executed", zap.String("txHash", result.TxHash))

	// Check delegator address
	delegatorBalance := getBalance(ctx, t, chain, delegator)
	require.Equal(t, delegatorInitialBalance.Sub(delegateAmount), delegatorBalance.Amount)

	// Make sure coins have been delegated
	ddResp, err := stakingClient.DelegatorDelegations(ctx, &stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: delegator.String(),
	})
	require.NoError(t, err)
	require.Equal(t, validatorsResp.Validators[0].Tokens.Add(delegateAmount), ddResp.DelegationResponses[0].Balance.Amount)

	// Undelegate coins
	undelegateMsg := stakingtypes.NewMsgUndelegate(delegator, valAddress, chain.NewCoin(delegateAmount))
	result, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(undelegateMsg)),
		undelegateMsg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Undelegation executed", zap.String("txHash", result.TxHash))

	// Wait for undelegation
	unbondingTime, err := time.ParseDuration(chain.NetworkConfig.StakingConfig.UnbondingTime)
	require.NoError(t, err)
	time.Sleep(unbondingTime + time.Second*2)

	// Check delegator balance
	delegatorBalance = getBalance(ctx, t, chain, delegator)
	require.Equal(t, delegatorInitialBalance, delegatorBalance.Amount)

	// Make sure coins have been undelegated
	resp, err := stakingClient.Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: valAddress.String(),
	})
	require.NoError(t, err)
	require.Equal(t, validatorsResp.Validators[0].Tokens, resp.Validator.Tokens)
}

func getBalance(ctx context.Context, t testing.T, chain testing.Chain, addr sdk.AccAddress) sdk.Coin {
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: addr.String()})
	require.NoError(t, err)

	var balance sdk.Coin
	for _, b := range resp.Balances {
		if b.Denom == chain.NetworkConfig.TokenSymbol {
			balance = b
			break
		}
	}
	require.NotNil(t, balance)

	return balance
}
