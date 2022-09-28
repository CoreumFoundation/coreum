package bank

import (
	"context"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var maxMemo = strings.Repeat("-", 256) // cosmos sdk is configured to accept maximum memo of 256 characters by default

// TestTransferDeterministicGas checks that transfer takes the deterministic amount of gas
func TestTransferDeterministicGas(ctx context.Context, t testing.T, chain testing.Chain) {
	gasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{})

	amount := testing.MustNewIntFromString(t, "1000000000000")
	fees := testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&banktypes.MsgSend{}),
		1,
		sdk.NewInt(0),
	)

	sender := testing.RandomWallet()
	receiver := testing.RandomWallet()

	senderInitialBalance := chain.NewCoin(fees.Add(amount))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, senderInitialBalance)))

	client := chain.Client

	gasPrice := chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice)
	toSend := chain.NewCoin(amount)
	gasUsed, err := sendAndReturnGasUsed(ctx, client, sender, receiver, toSend, gasExpected, gasPrice)

	require.NoError(t, err)
	assert.Equal(t, gasExpected, gasUsed)
}

// TestTransferFailsIfNotEnoughGasIsProvided checks that transfer fails if not enough gas is provided
func TestTransferFailsIfNotEnoughGasIsProvided(ctx context.Context, t testing.T, chain testing.Chain) {
	gasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	sender := testing.RandomWallet()

	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&banktypes.MsgSend{}),
		1,
		sdk.NewInt(10),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	gasPrice := chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice)
	_, err := sendAndReturnGasUsed(ctx, chain.Client, sender, sender,
		chain.NewCoin(sdk.NewInt(1)),
		// declaring gas limit as gasExpected-1 means that tx must fail
		gasExpected-1, gasPrice)
	assert.True(t, client.IsErr(err, cosmoserrors.ErrOutOfGas), err)
}

// TestTransferGasEstimation checks that gas is correctly estimated for send message
func TestTransferGasEstimation(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := testing.RandomWallet()
	receiver := testing.RandomWallet()

	amount := testing.MustNewIntFromString(t, "1000000000000")
	initialBalance := chain.NewCoin(testing.ComputeNeededBalance(
		chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
		chain.GasLimitByMsgs(&banktypes.MsgSend{}),
		1,
		amount,
	))
	require.NoError(t, chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(sender, initialBalance)))

	gasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{})
	estimatedGas, err := chain.Client.EstimateGas(ctx, tx.BaseInput{
		Signer:   sender,
		GasPrice: chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice),
	}, &banktypes.MsgSend{
		FromAddress: sender.Address().String(),
		ToAddress:   receiver.Address().String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amount)),
	})
	require.NoError(t, err)

	assert.Equal(t, gasExpected, estimatedGas)
}

func sendAndReturnGasUsed(
	ctx context.Context,
	coredClient client.Client,
	sender, receiver types.Wallet,
	toSend sdk.Coin,
	gasLimit uint64,
	gasPrice sdk.DecCoin,
) (uint64, error) {
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
