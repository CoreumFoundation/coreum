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

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

var cfg = config{
	NetworkConfig: coreumtesting.NetworkConfig,
}

func TestMain(m *testing.M) {
	var coredAddress, fundingPrivKey, logFormat, filter string

	flag.StringVar(&coredAddress, "cored-address", "localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&fundingPrivKey, "priv-key", "LPIPcUDVpp8Cn__g-YMntGts-DfDbd2gKTcgUgqSLfY", "Base64-encoded private key used to fund accounts required by tests")
	flag.StringVar(&filter, "filter", "", "Regular expression used to run only a subset of tests")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.Parse()

	cfg.CoredClient = client.New(cfg.NetworkConfig.ChainID, coredAddress)

	rpcClient, err := cosmosclient.NewClientFromNode("tcp://" + coredAddress)
	if err != nil {
		panic(err)
	}
	cfg.ClientContext = app.
		NewDefaultClientContext().
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient)

	cfg.FundingPrivKey, err = base64.RawURLEncoding.DecodeString(fundingPrivKey)
	if err != nil {
		panic(err)
	}

	cfg.Filter = regexp.MustCompile(filter)
	cfg.LogFormat = logger.Format(logFormat)
	cfg.LogVerbose = flag.Lookup("test.v").Value.String() == "true"

	// FIXME (wojtek): remove this once we have our own address encoder
	app.NewNetwork(cfg.NetworkConfig).SetupPrefixes()

	m.Run()
}

func Test(t *testing.T) {
	t.Parallel()

	testSet := tests.Tests()
	ctx := newContext(t, cfg)

	var err error
	fundingWallet := types.Wallet{Key: cfg.FundingPrivKey}
	fundingWallet.AccountNumber, fundingWallet.AccountSequence, err = cfg.CoredClient.GetNumberSequence(ctx, cfg.FundingPrivKey.Address())
	require.NoError(t, err)

	testCases := collectTestCases(cfg, fundingWallet, testSet)

	if len(testCases) == 0 {
		logger.Get(ctx).Warn("No tests to run")
		return
	}

	runTests(ctx, t, testCases)
}

type config struct {
	CoredClient    client.Client
	ClientContext  cosmosclient.Context
	NetworkConfig  app.NetworkConfig
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

type testCase struct {
	Name    string
	RunFunc func(ctx context.Context, t *testing.T)
}

func collectTestCases(cfg config, fundingWallet types.Wallet, testSet coreumtesting.TestSet) []testCase {
	faucet := &testingFaucet{
		client:        cfg.CoredClient,
		networkConfig: cfg.NetworkConfig,
		muCh:          make(chan struct{}, 1),
		fundingWallet: fundingWallet,
	}
	faucet.muCh <- struct{}{}

	chain := coreumtesting.Chain{
		ClientCtx:     cfg.ClientContext,
		NetworkConfig: cfg.NetworkConfig,
		Client:        cfg.CoredClient,
		Faucet:        faucet,
	}

	var testCases []testCase
	for _, testFunc := range testSet.SingleChain {
		testFunc := testFunc
		name := funcToName(testFunc)
		if !cfg.Filter.MatchString(name) {
			continue
		}

		testCases = append(testCases, testCase{
			Name: name,
			RunFunc: func(ctx context.Context, t *testing.T) {
				testFunc(ctx, t, chain)
			},
		})
	}
	return testCases
}

type testingFaucet struct {
	client        client.Client
	networkConfig app.NetworkConfig

	// muCh is used to serve the same purpose as `sync.Mutex` to protect `fundingWallet` against being used
	// to broadcast many transactions in parallel by different integration tests. The difference between this and `sync.Mutex`
	// is that test may exit immediately when `ctx` is canceled, without waiting for mutex to be unlocked.
	muCh          chan struct{}
	fundingWallet types.Wallet
}

func coinTypeToSDK(c types.Coin) sdk.Coin {
	return sdk.NewCoin(c.Denom, sdk.NewIntFromBigInt(c.Amount))
}

func (tf *testingFaucet) FundAccounts(ctx context.Context, accountsToFund ...coreumtesting.FundedAccount) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-tf.muCh:
		defer func() {
			tf.muCh <- struct{}{}
		}()
	}

	gasPrice := sdk.NewCoin(tf.networkConfig.TokenSymbol, tf.networkConfig.Fee.FeeModel.InitialGasPrice)

	log := logger.Get(ctx)
	log.Info("Funding accounts for test, it might take a while...")
	fundingPrivateKey := secp256k1.PrivKey{Key: cfg.FundingPrivKey}
	fundingAddress := sdk.AccAddress(fundingPrivateKey.PubKey().Address())

	var msgList []sdk.Msg
	for _, toFund := range accountsToFund {
		// FIXME (wojtek): Fund all accounts in single tx once new "client" is ready
		toPrivateKey := secp256k1.PrivKey{Key: toFund.Wallet.Key}
		toAddress := sdk.AccAddress(toPrivateKey.PubKey().Address())
		msg := &banktypes.MsgSend{
			FromAddress: fundingAddress.String(),
			ToAddress:   toAddress.String(),
			Amount: []sdk.Coin{
				coinTypeToSDK(toFund.Amount),
			},
		}
		msgList = append(msgList, msg)
	}

	signInput := tx.SignInput{
		PrivateKey: fundingPrivateKey,
		GasLimit:   cfg.NetworkConfig.Fee.DeterministicGas.BankSend * uint64(len(accountsToFund)),
		GasPrice:   gasPrice,
	}

	_, err := tx.BroadcastSync(ctx, cfg.ClientContext, signInput, msgList...)
	if err != nil {
		return err
	}

	logger.Get(ctx).Info("Test accounts funded")
	return nil
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
