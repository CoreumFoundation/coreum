package staking

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestDelegate checks that delegation works correctly
func TestDelegate(ctx context.Context, t testing.T, chain testing.Chain) {
	delegateAmount := sdk.NewInt(100)
	delegatorInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		delegateAmount,
	)

	// Create random delegator wallet
	delegator := chain.RandomWallet()

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			chain.AccAddressToLegacyWallet(delegator),
			chain.NewCoin(delegatorInitialBalance)),
	))

	// Fetch existing validator
	validators, err := chain.Client.GetValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)

	valAddress, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)

	// Delegate coins
	msg := stakingtypes.NewMsgDelegate(delegator, valAddress, chain.NewCoin(delegateAmount))
	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromName(delegator.String()).WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Delegation executed", zap.String("txHash", result.TxHash))

	// Check delegator address
	delegatorBalance, err := chain.Client.QueryBankBalances(ctx, chain.AccAddressToLegacyWallet(delegator))
	require.NoError(t, err)
	require.Equal(t, delegatorInitialBalance.Sub(delegateAmount), delegatorBalance[chain.NetworkConfig.TokenSymbol].Amount)

	// Make sure coins have been delegated
	resp, err := chain.Client.StakingQueryClient().Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: valAddress.String(),
	})
	require.NoError(t, err)
	require.Equal(t, validators[0].Tokens.Add(delegateAmount), resp.Validator.Tokens)
}
