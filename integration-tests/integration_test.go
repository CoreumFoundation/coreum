//go:build integration
// +build integration

package tests_test

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/integration-tests"
	coreumtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

func Test(t *testing.T) {
	var (
		coredAddress        = "localhost:26657"
		hornOfPlentyPrivKey = types.Secp256k1PrivateKey{0x86, 0x2f, 0x17, 0xc3, 0x51, 0x83, 0xbe, 0x2a, 0x2b, 0x5d, 0x5b, 0x5b, 0xb5, 0x53, 0x86, 0xd8, 0xad, 0xc8, 0xde, 0x51, 0xa9, 0x73, 0x3f, 0xb7, 0x7d, 0x72, 0xb9, 0x29, 0x91, 0xb7, 0x2c, 0x60}
	)

	loggerConfig := logger.ToolDefaultConfig
	loggerConfig.Format = logger.FormatYAML
	integrationTests := tests.Tests()

	network := app.NewNetwork(coreumtesting.NetworkConfig)
	network.SetupPrefixes()
	coredClient := client.New(network.ChainID(), coredAddress)

	var walletsToFund []walletToFund
	chain := coreumtesting.Chain{
		Network: &network,
		Client:  coredClient,
		Fund: func(wallet types.Wallet, amount types.Coin) {
			walletsToFund = append(walletsToFund, walletToFund{Wallet: wallet, Amount: amount})
		},
	}

	var testList []test
	for _, testFunc := range integrationTests.SingleChain {
		prepFunc, runFunc := testFunc(chain)
		testList = append(testList, test{
			Name:        funcToName(testFunc),
			PrepareFunc: prepFunc,
			RunFunc:     runFunc,
		})

	}

	ctx, cancel := context.WithCancel(logger.WithLogger(context.Background(), logger.New(loggerConfig)))
	t.Cleanup(cancel)

	for _, test := range testList {
		ctx := logger.With(ctx, zap.String("test", test.Name))
		require.NoError(t, test.PrepareFunc(ctx))
	}

	var err error
	hornOfPlentyWallet := types.Wallet{Key: hornOfPlentyPrivKey}
	hornOfPlentyWallet.AccountNumber, hornOfPlentyWallet.AccountSequence, err = coredClient.GetNumberSequence(ctx, hornOfPlentyPrivKey.Address())
	require.NoError(t, err)

	gasPrice, err := types.NewCoin(network.InitialGasPrice(), network.TokenSymbol())
	require.NoError(t, err)
	for _, toFund := range walletsToFund {
		// FIXME (wojtek): Fund all accounts in single tx once new "client" is ready
		encodedTx, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
			Base: tx.BaseInput{
				Signer:   hornOfPlentyWallet,
				GasLimit: network.DeterministicGas().BankSend,
				GasPrice: gasPrice,
			},
			Sender:   hornOfPlentyWallet,
			Receiver: toFund.Wallet,
			Amount:   toFund.Amount,
		})
		require.NoError(t, err)
		_, err = coredClient.Broadcast(ctx, encodedTx)
		require.NoError(t, err)
		hornOfPlentyWallet.AccountSequence++
	}

	for _, test := range testList {
		test := test
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(logger.With(ctx, zap.String("test", test.Name)))
			t.Cleanup(cancel)

			log := logger.Get(ctx)
			log.Info("Test started")
			test.RunFunc(ctx, t)
			if t.Failed() {
				log.Error("Test failed")
			} else {
				log.Info("Test succeeded")
			}
		})
	}
}

func funcToName(f interface{}) string {
	parts := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "/")
	repoName := parts[2]
	funcName := strings.TrimSuffix(parts[len(parts)-1], ".func1")

	return repoName + "." + funcName
}

type walletToFund struct {
	Wallet types.Wallet
	Amount types.Coin
}

type test struct {
	Name        string
	PrepareFunc coreumtesting.PrepareFunc
	RunFunc     coreumtesting.RunFunc
}
