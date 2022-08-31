package bank

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
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
	balances, err := queryBankBalances(ctx, chain.ClientCtx, wallet)
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

	// Transfer 10 cores from sender to receiver
	senderPrivateKey := secp256k1.PrivKey{Key: sender.Key}
	fromAddress := sdk.AccAddress(senderPrivateKey.PubKey().Address())
	receiverPrivateKey := secp256k1.PrivKey{Key: receiver.Key}
	toAddress := sdk.AccAddress(receiverPrivateKey.PubKey().Address())
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress.String(),
		ToAddress:   toAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}

	signInput := tx.SignInput{
		PrivateKey: senderPrivateKey,
		GasLimit:   chain.NetworkConfig.Fee.DeterministicGas.BankSend,
		GasPrice:   sdk.Coin{Amount: chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, Denom: chain.NetworkConfig.TokenSymbol},
		Memo:       maxMemo, // memo is set to max length here to charge as much gas as possible
	}
	txHash, err := tx.BroadcastAsync(ctx, chain.ClientCtx, signInput, msg)
	require.NoError(t, err)
	_, err = tx.AwaitTx(ctx, chain.ClientCtx, txHash)
	require.NoError(t, err)

	logger.Get(ctx).Info("Transfer executed", zap.String("txHash", txHash))

	// Query wallets for current balance
	balancesSender, err := queryBankBalances(ctx, chain.ClientCtx, sender)
	require.NoError(t, err)

	balancesReceiver, err := queryBankBalances(ctx, chain.ClientCtx, receiver)
	require.NoError(t, err)

	// Test that tokens disappeared from sender's wallet
	// - 10core were transferred to receiver
	// - 180000000core were taken as fee
	assert.Equal(t, "90", balancesSender[chain.NetworkConfig.TokenSymbol].Amount.String())

	// Test that tokens reached receiver's wallet
	assert.Equal(t, "20", balancesReceiver[chain.NetworkConfig.TokenSymbol].Amount.String())
}

func queryBankBalances(ctx context.Context, clientCtx client.Context, wallet types.Wallet) (map[string]sdk.Coin, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	bankQueryClient := banktypes.NewQueryClient(clientCtx)

	resp, err := bankQueryClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	balances := map[string]sdk.Coin{}
	for _, b := range resp.Balances {
		coin := sdk.NewCoin(b.Denom, b.Amount)
		balances[b.Denom] = coin
	}
	return balances, nil
}
