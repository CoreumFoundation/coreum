package auth

import (
	"context"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TODO (wojtek): once we have other coins add test verifying that transaction offering fee in coin other then CORE is rejected

// TestTooLowGasPrice verifies that transaction fails if offered gas price is below minimum level
// specified by the fee model of the network
func TestTooLowGasPrice(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(100),
	).BigInt(), chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: initialBalance,
		},
	))

	coredClient := chain.Client

	gasPriceWithMaxDiscount := chain.NetworkConfig.Fee.FeeModel.InitialGasPrice.ToDec().Mul(sdk.OneDec().Sub(chain.NetworkConfig.Fee.FeeModel.MaxDiscount)).TruncateInt()
	gasPrice := gasPriceWithMaxDiscount.Sub(sdk.OneInt())
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
			GasPrice: types.Coin{Amount: gasPrice.BigInt(), Denom: chain.NetworkConfig.TokenSymbol},
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   types.Coin{Denom: chain.NetworkConfig.TokenSymbol, Amount: big.NewInt(10)},
	})
	require.NoError(t, err)

	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.True(t, client.IsInsufficientFeeError(err))
}

// TestNoFee verifies that transaction fails if sender does not offer fee at all
func TestNoFee(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(100),
	).BigInt(), chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: initialBalance,
		},
	))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   types.Coin{Denom: chain.NetworkConfig.TokenSymbol, Amount: big.NewInt(10)},
	})
	require.NoError(t, err)

	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.True(t, client.IsInsufficientFeeError(err))
}

// TestGasLimitHigherThanMaxBlockGas verifies that transaction requiring more gas than MaxBlockGas fails
func TestGasLimitHigherThanMaxBlockGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas+1),
		1,
		sdk.NewInt(100),
	).BigInt(), chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: initialBalance,
		},
	))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas + 1), // transaction requires more gas than block can fit
			GasPrice: types.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice.BigInt(), Denom: chain.NetworkConfig.TokenSymbol},
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   types.Coin{Denom: chain.NetworkConfig.TokenSymbol, Amount: big.NewInt(10)},
	})
	require.NoError(t, err)

	// Broadcast should fail because gas limit is higher than the block capacity
	_, err = coredClient.Broadcast(ctx, txBytes)
	require.Error(t, err)
}

// TestGasLimitEqualToMaxBlockGas verifies that transaction requiring MaxBlockGas gas succeeds
func TestGasLimitEqualToMaxBlockGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
		uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas),
		1,
		sdk.NewInt(100),
	).BigInt(), chain.NetworkConfig.TokenSymbol)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: initialBalance,
		},
	))

	coredClient := chain.Client

	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas),
			GasPrice: types.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice.BigInt(), Denom: chain.NetworkConfig.TokenSymbol},
		},
		Sender:   sender,
		Receiver: sender,
		Amount:   types.Coin{Denom: chain.NetworkConfig.TokenSymbol, Amount: big.NewInt(10)},
	})
	require.NoError(t, err)

	_, err = coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)
}
