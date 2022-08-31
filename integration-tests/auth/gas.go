package auth

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

	gasPriceWithMaxDiscount := chain.NetworkConfig.Fee.FeeModel.InitialGasPrice.ToDec().Mul(sdk.OneDec().Sub(chain.NetworkConfig.Fee.FeeModel.MaxDiscount)).TruncateInt()
	gasPrice := gasPriceWithMaxDiscount.Sub(sdk.OneInt())

	privateKey := secp256k1.PrivKey{Key: sender.Key}
	fromAddress := sdk.AccAddress(privateKey.PubKey().Bytes())
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   fromAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}

	signInput := tx.SignInput{
		PrivateKey: privateKey,
		GasLimit:   chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice:   sdk.Coin{Amount: gasPrice, Denom: chain.NetworkConfig.TokenSymbol},
	}
	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
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

	privateKey := secp256k1.PrivKey{Key: sender.Key}
	fromAddress := sdk.AccAddress(privateKey.PubKey().Bytes())
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   fromAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}

	signInput := tx.SignInput{
		PrivateKey: privateKey,
		GasLimit:   chain.NetworkConfig.Fee.DeterministicGas.BankSend,
	}
	// Broadcast should fail because gas price is too low for transaction to enter mempool
	_, err = tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
	require.True(t, client.IsInsufficientFeeError(err))
}

// TestGasLimitHigherThanMaxBlockGas verifies that transaction requiring more gas than MaxBlockGas fails
func TestGasLimitHigherThanMaxBlockGas(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: testing.MustNewCoin(t, testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
				uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas+1),
				1,
				sdk.NewInt(100),
			), chain.NetworkConfig.TokenSymbol),
		},
	))

	privateKey := secp256k1.PrivKey{Key: sender.Key}
	fromAddress := sdk.AccAddress(privateKey.PubKey().Bytes())
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   fromAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}

	signInput := tx.SignInput{
		PrivateKey: privateKey,
		GasLimit:   uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas + 1), // transaction requires more gas than block can fit
		GasPrice:   sdk.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, Denom: chain.NetworkConfig.TokenSymbol},
	}

	// Broadcast should fail because gas limit is higher than the block capacity
	_, err := tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
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

	privateKey := secp256k1.PrivKey{Key: sender.Key}
	fromAddress := sdk.AccAddress(privateKey.PubKey().Bytes())
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   fromAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}

	signInput := tx.SignInput{
		PrivateKey: privateKey,
		GasLimit:   uint64(chain.NetworkConfig.Fee.FeeModel.MaxBlockGas),
		GasPrice:   sdk.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, Denom: chain.NetworkConfig.TokenSymbol},
	}

	_, err = tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
	require.NoError(t, err)
}
