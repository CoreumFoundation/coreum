package bank

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestTransferMaximumGas checks that transfer does not take more gas than assumed
func TestTransferMaximumGas(numOfTransactions int) testing.SingleChainSignature {
	return func(ctx context.Context, t testing.T, chain *testing.Chain) {
		const margin = 1.5
		maxGasAssumed := chain.NetworkConfig.Fee.DeterministicGas.BankSend // set it to 50%+ higher than maximum observed value

		amount := testing.MustNewIntFromString(t, "1000000000000")
		fees := testing.ComputeNeededBalance(
			chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
			chain.NetworkConfig.Fee.DeterministicGas.BankSend,
			numOfTransactions,
			sdk.NewInt(0),
		)

		wallet1 := testing.RandomWallet()
		wallet2 := testing.RandomWallet()

		wallet1InitialBalance := chain.NewCoin(fees.Add(amount))
		wallet2InitialBalance := chain.NewCoin(fees)

		require.NoError(t, chain.Faucet.FundAccounts(ctx,
			testing.NewFundedAccount(wallet1, wallet1InitialBalance),
			testing.NewFundedAccount(wallet2, wallet2InitialBalance),
		))

		client := chain.Client

		var err error
		wallet1.AccountNumber, wallet1.AccountSequence, err = client.GetNumberSequence(ctx, wallet1.Key.Address())
		require.NoError(t, err)
		wallet2.AccountNumber, wallet2.AccountSequence, err = client.GetNumberSequence(ctx, wallet2.Key.Address())
		require.NoError(t, err)

		var maxGasUsed int64
		gasPrice := chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice)
		toSend := chain.NewCoin(amount)
		for i, sender, receiver := numOfTransactions, wallet1, wallet2; i >= 0; i, sender, receiver = i-1, receiver, sender {
			gasUsed, err := sendAndReturnGasUsed(ctx, client, sender, receiver, toSend, maxGasAssumed, gasPrice)
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

// TestTransferFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided
func TestTransferFailsIfNotEnoughGasIsProvided(ctx context.Context, t testing.T, chain *testing.Chain) {
	maxGasAssumed := chain.NetworkConfig.Fee.DeterministicGas.BankSend
	sender := testing.RandomWallet()

	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		1,
		sdk.NewInt(10),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	gasPrice := chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice)
	_, err := sendAndReturnGasUsed(ctx, chain.Client, sender, sender,
		chain.NewCoin(sdk.NewInt(1)),
		// declaring gas limit as maxGasAssumed-1 means that tx must fail
		maxGasAssumed-1, gasPrice)
	assert.True(t, client.IsErr(err, cosmoserrors.ErrInsufficientFee))
}

func sendAndReturnGasUsed(
	ctx context.Context,
	coredClient client.Client,
	sender, receiver types.Wallet,
	toSend sdk.Coin,
	gasLimit uint64,
	gasPrice sdk.DecCoin,
) (int64, error) {
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: gasLimit,
			GasPrice: gasPrice,
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
