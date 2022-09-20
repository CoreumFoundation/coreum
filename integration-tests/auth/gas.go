package auth

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TODO (wojtek): once we have other coins add test verifying that transaction offering fee in coin other then CORE is rejected

// TestTooLowGasPrice verifies that transaction fails if offered gas price is below minimum level
// specified by the fee model of the network
func TestTooLowGasPrice(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(100),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	coredClient := chain.Client

	gasPriceWithMaxDiscount := chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice.Mul(sdk.OneDec().Sub(chain.NetworkConfig.Fee.FeeModel.Params().MaxDiscount))
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
			GasPrice: chain.NewDecCoin(gasPriceWithMaxDiscount.Sub(sdk.OneDec())),
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   chain.NewCoin(sdk.NewInt(10)),
	})
	require.NoError(t, err)

	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.True(t, client.IsErr(err, cosmoserrors.ErrInsufficientFee))
}

// TestNoFee verifies that transaction fails if sender does not offer fee at all
func TestNoFee(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(100),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   chain.NewCoin(sdk.NewInt(10)),
	})
	require.NoError(t, err)

	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.True(t, client.IsErr(err, cosmoserrors.ErrInsufficientFee))
}

// TestGasLimitHigherThanMaxBlockGas verifies that transaction requiring more gas than MaxBlockGas fails
func TestGasLimitHigherThanMaxBlockGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(sender, chain.NewCoin(testing.ComputeNeededBalance(
			chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
			uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas+1),
			1,
			sdk.NewInt(100),
		))),
	))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas + 1), // transaction requires more gas than block can fit
			GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   chain.NewCoin(sdk.NewInt(10)),
	})
	require.NoError(t, err)

	// Broadcast should fail because gas limit is higher than the block capacity
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.Error(t, err)
}

// TestGasLimitEqualToMaxBlockGas verifies that transaction requiring MaxBlockGas gas succeeds
func TestGasLimitEqualToMaxBlockGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
		1,
		sdk.NewInt(100),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas),
			GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   chain.NewCoin(sdk.NewInt(10)),
	})
	require.NoError(t, err)

	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}
