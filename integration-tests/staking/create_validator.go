package staking

import (
	"context"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestCreateValidator checks that validator creation works correctly
func TestCreateValidator(ctx context.Context, t testing.T, chain testing.Chain) {
	validatorAmount := sdk.NewInt(100)
	validatorInitialBalance := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		validatorAmount,
	)

	// Create random validator wallet
	validator := chain.RandomWallet()
	validatorAddr := sdk.ValAddress(validator)

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			chain.AccAddressToLegacyWallet(validator),
			chain.NewCoin(validatorInitialBalance),
		),
	))

	// Create validator
	msg, err := stakingtypes.NewMsgCreateValidator(
		validatorAddr,
		cosmosed25519.GenPrivKey().PubKey(),
		chain.NewCoin(validatorAmount),
		stakingtypes.NewDescription("a", "b", "c", "d", "e"),
		stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		sdk.NewInt(1),
	)
	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromName(validator.String()).WithFromAddress(validator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Validator creation executed", zap.String("txHash", result.TxHash))

	// Make sure validator has been created
	resp, err := chain.Client.StakingQueryClient().Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: validatorAddr.String(),
	})
	require.NoError(t, err)
	require.Equal(t, validatorAmount, resp.Validator.Tokens)
}
