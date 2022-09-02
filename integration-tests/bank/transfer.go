package bank

import (
	"context"
	"encoding/hex"
	"time"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
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
	balances, err := chain.Client.QueryBankBalances(ctx, wallet)
	require.NoError(t, err)

	// Test that wallet owns expected balance
	assert.Equal(t, "100", balances[chain.NetworkConfig.TokenSymbol].Amount.String())
}

// TestCoreTransfer checks that core is transferred correctly between wallets
func TestCoreTransfer(ctx context.Context, t testing.T, chain testing.Chain) {
	// Create two random wallets
	clientCtx := chain.ClientContext
	txf, walletGen := setup(clientCtx)
	sender, senderAddress := walletGen()
	receiver, receiverAddress := walletGen()

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
	msg := &banktypes.MsgSend{
		FromAddress: senderAddress.String(),
		ToAddress:   receiverAddress.String(),
		Amount: []sdk.Coin{
			{Denom: chain.NetworkConfig.TokenSymbol, Amount: sdk.NewInt(10)},
		},
	}
	clCtx := clientCtx.
		WithFromName(sender.Name).
		WithFromAddress(senderAddress).
		WithBroadcastMode(flags.BroadcastBlock)
	txf = txf.
		WithGas(chain.NetworkConfig.Fee.DeterministicGas.BankSend).
		WithGasPrices(sdk.NewCoin(chain.NetworkConfig.TokenSymbol, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice).String())
	result, err := tx.BroadcastTx(ctx, clCtx, txf, msg)
	require.NoError(t, err)
	time.Sleep(4 * time.Second)

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

type walletGen = func() (types.Wallet, sdk.AccAddress)

func gen(kr keyring.UnsafeKeyring) walletGen {
	return func() (types.Wallet, sdk.AccAddress) {
		name := uuid.New().String()
		_, _, err := kr.NewMnemonic(name, keyring.English, "", "", hd.Secp256k1)
		if err != nil {
			// we are using panic here, since we are sure it will not error out, and handling error
			// upstream is a waste of time.
			panic(err)
		}
		privKeyHex, err := kr.UnsafeExportPrivKeyHex(name)
		if err != nil {
			panic(err)
		}

		privKeyBytes, err := hex.DecodeString(privKeyHex)
		if err != nil {
			panic(err)
		}

		privKey := secp256k1.PrivKey{Key: privKeyBytes}
		address := sdk.AccAddress(privKey.PubKey().Address())

		return types.Wallet{Name: name, Key: privKeyBytes}, address
	}
}

func setup(clientCtx cosmosclient.Context) (tx.Factory, walletGen) {
	kr := keyring.NewUnsafe(keyring.NewInMemory())
	txf := tx.Factory{}.
		WithKeybase(kr).
		WithChainID(clientCtx.ChainID).
		WithTxConfig(clientCtx.TxConfig)
	return txf, gen(kr)
}
