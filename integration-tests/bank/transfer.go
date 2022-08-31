package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestInitialBalance checks that initial balance is set by genesis block
func TestInitialBalance(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create new random wallet
	wallet := testing.RandomWallet()

	// Prefunding account required by test
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: wallet,
			Amount: testing.MustNewCoin(t, sdk.NewInt(100), chain.NetworkConfig.TokenSymbol),
		},
	))

	// Query for current balance available on the wallet
	balances, err := chain.Client.QueryBankBalances(ctx, wallet)
	require.NoError(t, err)

	// Test that wallet owns expected balance
	assert.Equal(t, "100", balances[chain.NetworkConfig.TokenSymbol].Amount.String())
}

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create two random wallets
	sender := testing.RandomWallet()
	receiver := testing.RandomWallet()

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: sender,
			Amount: testing.MustNewCoin(t, testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.InitialGasPrice,
				chain.NetworkConfig.Fee.DeterministicGas.BankSend,
				1,
				sdk.NewInt(100),
			), chain.NetworkConfig.TokenSymbol),
		},
		testing.FundedAccount{
			Wallet: receiver,
			Amount: testing.MustNewCoin(t, sdk.NewInt(10), chain.NetworkConfig.TokenSymbol),
		},
	))

	// Create client so we can send transactions and query state
	coredClient := chain.Client

	// Transfer 10 cores from sender to receiver
	txBytes, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
		Base: tx.BaseInput{
			Signer:   sender,
			GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
			GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, chain.NetworkConfig.TokenSymbol),
		},
		Sender:   sender,
		Receiver: receiver,
		Amount:   testing.MustNewCoin(t, sdk.NewInt(10), chain.NetworkConfig.TokenSymbol),
	})
	require.NoError(t, err)
	result, err := coredClient.Broadcast(ctx, txBytes)
	require.NoError(t, err)

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", result.TxHash))

	// Query wallets for current balance
	balancesSender, err := coredClient.QueryBankBalances(ctx, sender)
	require.NoError(t, err)

	balancesReceiver, err := coredClient.QueryBankBalances(ctx, receiver)
	require.NoError(t, err)

	// Test that tokens disappeared from sender's wallet
	// - 10core were transferred to receiver
	// - 180000000core were taken as fee
	assert.Equal(t, "90", balancesSender[chain.NetworkConfig.TokenSymbol].Amount.String())

	// Test that tokens reached receiver's wallet
	assert.Equal(t, "20", balancesReceiver[chain.NetworkConfig.TokenSymbol].Amount.String())
}
