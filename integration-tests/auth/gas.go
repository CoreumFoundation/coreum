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

// TestTooLowGasPrice verifies that transaction fails if offered gas price is below minimum level
// specified by the fee model of the network
func TestTooLowGasPrice(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.FeeModel().InitialGasPrice,
				chain.Network.DeterministicGas().BankSend,
				1,
				sdk.NewInt(100),
			).BigInt(), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			gasPriceWithMaxDiscountFloat := new(big.Float).SetInt(chain.Network.FeeModel().InitialGasPrice.BigInt())
			gasPriceWithMaxDiscountFloat.Mul(gasPriceWithMaxDiscountFloat, big.NewFloat(1.-chain.Network.FeeModel().MaxDiscount))
			gasPriceWithMaxDiscount, _ := gasPriceWithMaxDiscountFloat.Int(nil)

			gasPrice := new(big.Int).Sub(gasPriceWithMaxDiscount, big.NewInt(1))
			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
					GasPrice: types.Coin{Amount: gasPrice, Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)

			// Broadcast should fail because gas price is too low for transaction to enter mempool
			_, err = coredClient.Broadcast(ctx, txBytes)
			require.True(t, client.IsInsufficientFeeError(err))
		}
}

// TestNoFee verifies that transaction fails if sender does not offer fee at all
func TestNoFee(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.FeeModel().InitialGasPrice,
				chain.Network.DeterministicGas().BankSend,
				1,
				sdk.NewInt(100),
			).BigInt(), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: chain.Network.DeterministicGas().BankSend,
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)

			// Broadcast should fail because gas price is too low for transaction to enter mempool
			_, err = coredClient.Broadcast(ctx, txBytes)
			require.True(t, client.IsInsufficientFeeError(err))
		}
}

// TestGasLimitHigherThanMaxBlockGas verifies that transaction requiring more gas than MaxBlockGas fails
func TestGasLimitHigherThanMaxBlockGas(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.FeeModel().MaxGasPrice,
				uint64(chain.Network.FeeModel().MaxBlockGas+1),
				1,
				sdk.NewInt(100),
			).BigInt(), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: uint64(chain.Network.FeeModel().MaxBlockGas + 1), // transaction requires more gas than block can fit
					GasPrice: types.Coin{Amount: chain.Network.FeeModel().InitialGasPrice.BigInt(), Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)

			// Broadcast should fail because gas limit is higher than the block capacity
			_, err = coredClient.Broadcast(ctx, txBytes)
			require.Error(t, err)
		}
}

// TestGasLimitEqualToMaxBlockGas verifies that transaction requiring MaxBlockGas gas succeeds
func TestGasLimitEqualToMaxBlockGas(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.FeeModel().MaxGasPrice,
				uint64(chain.Network.FeeModel().MaxBlockGas),
				1,
				sdk.NewInt(100),
			).BigInt(), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			coredClient := chain.Client

			txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
				Base: tx.BaseInput{
					Signer:   sender,
					GasLimit: uint64(chain.Network.FeeModel().MaxBlockGas),
					GasPrice: types.Coin{Amount: chain.Network.FeeModel().InitialGasPrice.BigInt(), Denom: chain.Network.TokenSymbol()},
				},
				Sender:   sender,
				Receiver: sender,
				Amount:   types.Coin{Denom: chain.Network.TokenSymbol(), Amount: big.NewInt(10)},
			})
			require.NoError(t, err)

			_, err = coredClient.Broadcast(ctx, txBytes)
			require.NoError(t, err)
		}
}
