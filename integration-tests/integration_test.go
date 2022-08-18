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
	cfg := configFromCLI()

	testSet := tests.Tests()

	network := app.NewNetwork(coreumtesting.NetworkConfig)
	network.SetupPrefixes()
	coredClient := client.New(network.ChainID(), cfg.CoredAddress)

	ctx := newContext(t, cfg)
	testCases, err := prepareTestCases(ctx, cfg.FundingPrivKey, coredClient, network, testSet)
	require.NoError(t, err)

	runTests(t, ctx, testCases)
}

type config struct {
	CoredAddress   string
	FundingPrivKey types.Secp256k1PrivateKey
	LogFormat      logger.Format
	LogVerbose     bool
}

func configFromCLI() config {
	// TODO: really read this from CLI
	return config{
		CoredAddress:   "localhost:26657",
		FundingPrivKey: types.Secp256k1PrivateKey{0x2c, 0xf2, 0xf, 0x71, 0x40, 0xd5, 0xa6, 0x9f, 0x2, 0x9f, 0xff, 0xe0, 0xf9, 0x83, 0x27, 0xb4, 0x6b, 0x6c, 0xf8, 0x37, 0xc3, 0x6d, 0xdd, 0xa0, 0x29, 0x37, 0x20, 0x52, 0xa, 0x92, 0x2d, 0xf6},
	}
}

func newContext(t *testing.T, cfg config) context.Context {
	loggerConfig := logger.Config{
		Format:  cfg.LogFormat,
		Verbose: cfg.LogVerbose,
	}

	ctx, cancel := context.WithCancel(logger.WithLogger(context.Background(), logger.New(loggerConfig)))
	t.Cleanup(cancel)

	return ctx
}

type walletToFund struct {
	Wallet types.Wallet
	Amount types.Coin
}

type testCase struct {
	Name        string
	PrepareFunc coreumtesting.PrepareFunc
	RunFunc     coreumtesting.RunFunc
}

func prepareTestCases(
	ctx context.Context,
	fundingPrivKey types.Secp256k1PrivateKey,
	coredClient client.Client,
	network app.Network,
	testSet coreumtesting.TestSet,
) ([]testCase, error) {
	// FIXME (wojtek): A lot of logic happens in this function due to how `walletsToFund` slice is built
	// once `crust` is switched to new framework it will be redone.

	var walletsToFund []walletToFund
	chain := coreumtesting.Chain{
		Network: &network,
		Client:  coredClient,
		Fund: func(wallet types.Wallet, amount types.Coin) {
			walletsToFund = append(walletsToFund, walletToFund{Wallet: wallet, Amount: amount})
		},
	}

	var testCases []testCase
	for _, testFunc := range testSet.SingleChain {
		prepFunc, runFunc := testFunc(chain)
		testCases = append(testCases, testCase{
			Name:        funcToName(testFunc),
			PrepareFunc: prepFunc,
			RunFunc:     runFunc,
		})
	}

	for _, tc := range testCases {
		ctx := logger.With(ctx, zap.String("test", tc.Name))
		if err := tc.PrepareFunc(ctx); err != nil {
			return nil, err
		}
	}

	var err error
	fundingWallet := types.Wallet{Key: fundingPrivKey}
	fundingWallet.AccountNumber, fundingWallet.AccountSequence, err = coredClient.GetNumberSequence(ctx, fundingPrivKey.Address())
	if err != nil {
		return nil, err
	}

	gasPrice, err := types.NewCoin(network.InitialGasPrice(), network.TokenSymbol())
	if err != nil {
		return nil, err
	}

	for _, toFund := range walletsToFund {
		// FIXME (wojtek): Fund all accounts in single tx once new "client" is ready
		encodedTx, err := coredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
			Base: tx.BaseInput{
				Signer:   fundingWallet,
				GasLimit: network.DeterministicGas().BankSend,
				GasPrice: gasPrice,
			},
			Sender:   fundingWallet,
			Receiver: toFund.Wallet,
			Amount:   toFund.Amount,
		})
		if err != nil {
			return nil, err
		}
		if _, err := coredClient.Broadcast(ctx, encodedTx); err != nil {
			return nil, err
		}
		fundingWallet.AccountSequence++
	}
	return testCases, nil
}

func runTests(t *testing.T, ctx context.Context, testCases []testCase) {
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(logger.With(ctx, zap.String("test", tc.Name)))
			t.Cleanup(cancel)

			log := logger.Get(ctx)
			log.Info("Test started")
			tc.RunFunc(ctx, t)
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
