package staking

import (
	"context"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
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
	validator := testing.RandomWallet()
	validatorAddr := sdk.ValAddress(validator.Address())

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(validator, chain.NewCoin(validatorInitialBalance)),
	))

	// Create validator
	txBytes, err := chain.Client.PrepareTxCreateValidator(ctx, client.TxCreateValidatorInput{
		Validator:         validatorAddr,
		PubKey:            (&cosmossecp256k1.PrivKey{Key: validator.Key}).PubKey(),
		Amount:            chain.NewCoin(validatorAmount),
		Description:       stakingtypes.NewDescription("a", "b", "c", "d", "e"),
		CommissionRates:   stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		MinSelfDelegation: sdk.NewInt(1),
		Base: tx.BaseInput{
			Signer:   validator,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
			GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
		},
	})
	require.NoError(t, err)
	_, err = chain.Client.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	// Make sure validator has been created
	validatorModel, err := chain.Client.GetValidator(ctx, validatorAddr)
	require.NoError(t, err)
	require.Equal(t, validatorAmount, validatorModel.Tokens)
}
