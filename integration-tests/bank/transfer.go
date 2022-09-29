package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// FIXME (wojtek): add test verifying that transfer fails if sender is out of balance.

// TestInitialBalance checks that initial balance is set by genesis block
func TestInitialBalance(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create new random wallet
	wallet := testing.RandomWallet()

	// Prefunding account required by test
	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(wallet, chain.NewCoin(sdk.NewInt(100))),
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
	sender := chain.RandomWallet()
	receiver := chain.RandomWallet()

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			chain.AccAddressToLegacyWallet(sender),
			chain.NewCoin(testing.ComputeNeededBalance(
				chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice,
				chain.GasLimitByMsgs(&banktypes.MsgSend{}),
				1,
				sdk.NewInt(100),
			)),
		),
	))

	require.NoError(t, chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(
			chain.AccAddressToLegacyWallet(receiver),
			chain.NewCoin(sdk.NewInt(10)),
		),
	))

	// Transfer 10 cores from sender to receiver
	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   receiver.String(),
		Amount: []sdk.Coin{
			chain.NewCoin(sdk.NewInt(10)),
		},
	}

	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)
	require.NoError(t, err)

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", result.TxHash))

	// Query wallets for current balance
	balancesSender, err := chain.Client.QueryBankBalances(ctx, chain.AccAddressToLegacyWallet(sender))
	require.NoError(t, err)

	balancesReceiver, err := chain.Client.QueryBankBalances(ctx, chain.AccAddressToLegacyWallet(receiver))
	require.NoError(t, err)

	// Test that tokens disappeared from sender's wallet
	// - 10core were transferred to receiver
	// - 180000000core were taken as fee
	assert.Equal(t, "90", balancesSender[chain.NetworkConfig.TokenSymbol].Amount.String())

	// Test that tokens reached receiver's wallet
	assert.Equal(t, "20", balancesReceiver[chain.NetworkConfig.TokenSymbol].Amount.String())
}
