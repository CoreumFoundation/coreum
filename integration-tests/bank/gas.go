package bank

import (
	"context"
	"math/big"
	"strings"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestTransferMaximumGas checks that transfer does not take more gas than assumed
func TestTransferMaximumGas(numOfTransactions int) testing.SingleChainSignature {
	return func(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
		const margin = 1.5
		maxGasAssumed := chain.Network.DeterministicGas().BankSend // set it to 50%+ higher than maximum observed value

		amount, ok := big.NewInt(0).SetString("100000000000000000000000000000000000", 10)
		if !ok {
			panic("invalid amount")
		}

		fees := testing.ComputeNeededBalance(
			chain.Network.InitialGasPrice(),
			chain.Network.DeterministicGas().BankSend,
			numOfTransactions,
			big.NewInt(0),
		)

		wallet1 := testing.RandomWallet()
		wallet2 := testing.RandomWallet()

		return func(ctx context.Context) error {
				wallet1InitialBalance, err := types.NewCoin(new(big.Int).Add(fees, amount), chain.Network.TokenSymbol())
				if err != nil {
					return err
				}
				wallet2InitialBalance, err := types.NewCoin(fees, chain.Network.TokenSymbol())
				if err != nil {
					return err
				}

				if err := chain.Network.FundAccount(wallet1.Key.PubKey(), wallet1InitialBalance.String()); err != nil {
					return err
				}
				return chain.Network.FundAccount(wallet2.Key.PubKey(), wallet2InitialBalance.String())
			},
			func(ctx context.Context, t testing.T) {
				client := chain.Client

				var err error
				wallet1.AccountNumber, wallet1.AccountSequence, err = client.GetNumberSequence(ctx, wallet1.Key.Address())
				require.NoError(t, err)
				wallet2.AccountNumber, wallet2.AccountSequence, err = client.GetNumberSequence(ctx, wallet2.Key.Address())
				require.NoError(t, err)

				var maxGasUsed int64
				toSend := types.Coin{Denom: chain.Network.TokenSymbol(), Amount: amount}
				for i, sender, receiver := numOfTransactions, wallet1, wallet2; i >= 0; i, sender, receiver = i-1, receiver, sender {
					gasUsed, err := sendAndReturnGasUsed(ctx, client, sender, receiver, toSend, maxGasAssumed, *chain.Network)
					if !assert.NoError(t, err) {
						break
					}

					if gasUsed > maxGasUsed {
						maxGasUsed = gasUsed
					}
					sender.AccountSequence++
				}
				assert.LessOrEqual(t, margin*float64(maxGasUsed), float64(maxGasAssumed))
				logger.Get(ctx).Info("Maximum gas used", zap.Int64("maxGasUsed", maxGasUsed))
			}
	}
}

// TestTransferFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided
func TestTransferFailsIfNotEnoughGasIsProvided(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	maxGasAssumed := chain.Network.DeterministicGas().BankSend
	sender := testing.RandomWallet()

	return func(ctx context.Context) error {
			initialBalance, err := types.NewCoin(testing.ComputeNeededBalance(
				chain.Network.InitialGasPrice(),
				chain.Network.DeterministicGas().BankSend,
				1,
				big.NewInt(10),
			), chain.Network.TokenSymbol())
			if err != nil {
				return err
			}
			return chain.Network.FundAccount(sender.Key.PubKey(), initialBalance.String())
		},
		func(ctx context.Context, t testing.T) {
			_, err := sendAndReturnGasUsed(ctx, chain.Client, sender, sender,
				types.Coin{Amount: big.NewInt(1), Denom: chain.Network.TokenSymbol()},
				// declaring gas limit as maxGasAssumed-1 means that tx must fail
				maxGasAssumed-1, *chain.Network)
			assert.True(t, client.IsInsufficientFeeError(err))
		}
}

func sendAndReturnGasUsed(ctx context.Context, coredClient client.Client, sender, receiver types.Wallet, toSend types.Coin, gasLimit uint64, network app.Network) (int64, error) {
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: gasLimit,
			GasPrice: types.Coin{Amount: network.InitialGasPrice(), Denom: network.TokenSymbol()},
			Memo:     maxMemo, // memo is set to max length here to charge as much gas as possible
		},
		Sender:   sender,
		Receiver: receiver,
		Amount:   toSend,
	})
	if err != nil {
		return 0, err
	}
	result, err := coredClient.Broadcast(ctx, txBytes)
	if err != nil {
		return 0, err
	}
	return result.GasUsed, nil
}
