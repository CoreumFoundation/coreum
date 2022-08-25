//go:build integration
// +build integration

package tests_test

import (
	"context"
	"encoding/base64"
	"flag"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	tests "github.com/CoreumFoundation/coreum/integration-tests"
	coreumtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var cfg config

func TestMain(m *testing.M) {
	var coredAddress, fundingPrivKey, logFormat, filter string

	flag.StringVar(&coredAddress, "cored-address", "localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&fundingPrivKey, "priv-key", "LPIPcUDVpp8Cn__g-YMntGts-DfDbd2gKTcgUgqSLfY", "Base64-encoded private key used to fund accounts required by tests")
	flag.StringVar(&filter, "filter", "", "Regular expression used to run only a subset of tests")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.Parse()

	cfg.Network = app.NewNetwork(coreumtesting.NetworkConfig)
	cfg.Network.SetupPrefixes()
	cfg.CoredClient = client.New(cfg.Network.ChainID(), coredAddress)

	var err error
	cfg.FundingPrivKey, err = base64.RawURLEncoding.DecodeString(fundingPrivKey)
	if err != nil {
		panic(err)
	}

	cfg.Filter = regexp.MustCompile(filter)
	cfg.LogFormat = logger.Format(logFormat)
	cfg.LogVerbose = flag.Lookup("test.v").Value.String() == "true"

	m.Run()
}

func Test(t *testing.T) {
	t.Parallel()

	testSet := tests.Tests()

	ctx := newContext(t, cfg)
	testCases, err := prepareTestCases(ctx, cfg, testSet)
	require.NoError(t, err)

	if len(testCases) == 0 {
		logger.Get(ctx).Warn("No tests to run")
		return
	}

	runTests(ctx, t, testCases)
}

type config struct {
	CoredClient    client.Client
	Network        app.Network
	FundingPrivKey types.Secp256k1PrivateKey
	Filter         *regexp.Regexp
	LogFormat      logger.Format
	LogVerbose     bool
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
	cfg config,
	testSet coreumtesting.TestSet,
) ([]testCase, error) {
	// FIXME (wojtek): A lot of logic happens in this function due to how `walletsToFund` slice is built
	// once `crust` is switched to new framework it will be redone.

	var walletsToFund []walletToFund
	chain := coreumtesting.Chain{
		Network: &cfg.Network,
		Client:  cfg.CoredClient,
		Fund: func(wallet types.Wallet, amount types.Coin) {
			walletsToFund = append(walletsToFund, walletToFund{Wallet: wallet, Amount: amount})
		},
	}

	var testCases []testCase
	for _, testFunc := range testSet.SingleChain {
		name := funcToName(testFunc)
		if !cfg.Filter.MatchString(name) {
			continue
		}

		prepFunc, runFunc := testFunc(chain)
		testCases = append(testCases, testCase{
			Name:        name,
			PrepareFunc: prepFunc,
			RunFunc:     runFunc,
		})
	}

	if len(testCases) == 0 {
		return nil, nil
	}

	for _, tc := range testCases {
		ctx := logger.With(ctx, zap.String("test", tc.Name))
		if err := tc.PrepareFunc(ctx); err != nil {
			return nil, err
		}
	}

	logger.Get(ctx).Info("Funding accounts for tests, it might take a while...")

	var err error
	fundingWallet := types.Wallet{Key: cfg.FundingPrivKey}
	fundingWallet.AccountNumber, fundingWallet.AccountSequence, err = cfg.CoredClient.GetNumberSequence(ctx, cfg.FundingPrivKey.Address())
	if err != nil {
		return nil, err
	}

	gasPrice, err := types.NewCoin(cfg.Network.InitialGasPrice(), cfg.Network.TokenSymbol())
	if err != nil {
		return nil, err
	}

	for _, toFund := range walletsToFund {
		// FIXME (wojtek): Fund all accounts in single tx once new "client" is ready
		encodedTx, err := cfg.CoredClient.PrepareTxBankSend(ctx, client.TxBankSendInput{
			Base: tx.BaseInput{
				Signer:   fundingWallet,
				GasLimit: cfg.Network.DeterministicGas().BankSend,
				GasPrice: gasPrice,
			},
			Sender:   fundingWallet,
			Receiver: toFund.Wallet,
			Amount:   toFund.Amount,
		})
		if err != nil {
			return nil, err
		}
		if _, err := cfg.CoredClient.Broadcast(ctx, encodedTx); err != nil {
			return nil, err
		}
		fundingWallet.AccountSequence++
	}

	logger.Get(ctx).Info("Test accounts funded")

	return testCases, nil
}

func runTests(ctx context.Context, t *testing.T, testCases []testCase) {
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
