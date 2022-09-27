package staking

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestDelegate checks that delegation works correctly
func TestDelegate(ctx context.Context, t testing.T, chain testing.Chain) {
	delegateAmount := sdk.NewInt(100)

	// Create random delegator wallet
	delegator := testing.RandomWallet()

	// Fund wallets
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(delegator, chain.NewCoin(delegateAmount)),
	))

	// Fetch existing validator
	validators, err := chain.Client.GetValidators(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, validators)

	valAddress, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)

	// Delegate coins
	txBytes, err := chain.Client.PrepareTxSubmitDelegation(ctx, client.TxSubmitDelegationInput{
		Base: tx.BaseInput{
			Signer:   delegator,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
			GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
		},
		Delegator: delegator,
		Validator: valAddress,
		Amount:    chain.NewCoin(delegateAmount),
	})
	require.NoError(t, err)
	_, err = chain.Client.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	// Make sure coins have been delegated
	validator, err := chain.Client.GetValidator(ctx, valAddress)
	require.NoError(t, err)
	require.Equal(t, validators[0].Tokens.Add(delegateAmount), validator.Tokens)
}
