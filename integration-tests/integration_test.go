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
	"time"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	tests "github.com/CoreumFoundation/coreum/integration-tests"
	coreumtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var cfg = testingConfig{
	NetworkConfig: coreumtesting.NetworkConfig,
}

func TestMain(m *testing.M) {
	var fundingPrivKey, coredAddress, logFormat, filter string

	flag.StringVar(&coredAddress, "cored-address", "tcp://localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&fundingPrivKey, "priv-key", "LPIPcUDVpp8Cn__g-YMntGts-DfDbd2gKTcgUgqSLfY", "Base64-encoded private key used to fund accounts required by tests")
	flag.StringVar(&filter, "filter", "", "Regular expression used to run only a subset of tests")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.Parse()

	decodedFundingPrivKey, err := base64.RawURLEncoding.DecodeString(fundingPrivKey)
	if err != nil {
		panic(err)
	}
	cfg.FundingPrivKey = decodedFundingPrivKey
	cfg.CoredAddress = coredAddress
	cfg.Filter = regexp.MustCompile(filter)
	cfg.LogFormat = logger.Format(logFormat)
	cfg.LogVerbose = flag.Lookup("test.v").Value.String() == "true"

	// FIXME (wojtek): remove this once we have our own address encoder
	config.NewNetwork(cfg.NetworkConfig).SetupPrefixes()

	m.Run()
}

func Test(t *testing.T) {
	t.Parallel()

	testSet := tests.Tests()
	ctx := newContext(t, cfg)

	chain, err := newChain(ctx, cfg)
	require.NoError(t, err)

	testCases := collectTestCases(chain, testSet, cfg.Filter)
	if len(testCases) == 0 {
		logger.Get(ctx).Warn("No tests to run")
		return
	}

	runTests(ctx, t, testCases)
}

type testingConfig struct {
	CoredAddress   string
	NetworkConfig  config.NetworkConfig
	FundingPrivKey types.Secp256k1PrivateKey
	Filter         *regexp.Regexp
	LogFormat      logger.Format
	LogVerbose     bool
}

func newChain(ctx context.Context, cfg testingConfig) (coreumtesting.Chain, error) {
	//nolint:contextcheck // `New->New->NewWithClient->New$1` should pass the context parameter
	coredClient := client.New(cfg.NetworkConfig.ChainID, cfg.CoredAddress)
	//nolint:contextcheck // `New->NewWithClient` should pass the context parameter
	rpcClient, err := cosmosclient.NewClientFromNode(cfg.CoredAddress)
	if err != nil {
		panic(err)
	}
	clientContext := config.NewClientContext(app.ModuleBasics).
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient).
		WithBroadcastMode(flags.BroadcastBlock)

	fundingWallet := types.Wallet{Key: cfg.FundingPrivKey}
	fundingWallet.AccountNumber, fundingWallet.AccountSequence, err = coredClient.GetNumberSequence(ctx, cfg.FundingPrivKey.Address())
	if err != nil {
		return coreumtesting.Chain{}, errors.Wrapf(err, "failed to get funding wallet sequence")
	}

	chain := coreumtesting.Chain{
		Client:        coredClient,
		ClientContext: clientContext,
		NetworkConfig: cfg.NetworkConfig,
		Keyring:       keyring.NewInMemory(),
	}

	faucet, err := newTestingFaucet(cfg, chain)
	if err != nil {
		return coreumtesting.Chain{}, err
	}

	chain.Faucet = faucet

	return chain, nil
}

func newContext(t *testing.T, cfg testingConfig) context.Context {
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

func collectTestCases(chain coreumtesting.Chain, testSet coreumtesting.TestSet, testFilter *regexp.Regexp) []testCase {
	var testCases []testCase
	for _, testFunc := range testSet.SingleChain {
		testFunc := testFunc
		name := funcToName(testFunc)
		if !testFilter.MatchString(name) {
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

type fundingRequest struct {
	AccountsToFund []coreumtesting.FundedAccount
	FundedCh       chan error
}

func newTestingFaucet(cfg testingConfig, chain coreumtesting.Chain) (*testingFaucet, error) {
	privKey := &cosmossecp256k1.PrivKey{Key: cfg.FundingPrivKey}
	keyringDB := keyring.NewInMemory()
	err := keyringDB.ImportPrivKey("faucet", crypto.EncryptArmorPrivKey(privKey, "dummy", privKey.Type()), "dummy")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	faucet := &testingFaucet{
		chain:         chain,
		keyring:       keyringDB,
		address:       sdk.AccAddress(privKey.PubKey().Address()),
		networkConfig: cfg.NetworkConfig,
		queue:         make(chan fundingRequest),
		muCh:          make(chan struct{}, 1),
	}
	faucet.muCh <- struct{}{}
	return faucet, err
}

type testingFaucet struct {
	chain         coreumtesting.Chain
	keyring       keyring.Keyring
	address       sdk.AccAddress
	networkConfig config.NetworkConfig
	queue         chan fundingRequest

	// muCh is used to serve the same purpose as `sync.Mutex` to protect `fundingWallet` against being used
	// to broadcast many transactions in parallel by different integration tests. The difference between this and `sync.Mutex`
	// is that test may exit immediately when `ctx` is canceled, without waiting for mutex to be unlocked.
	muCh chan struct{}
}

func (tf *testingFaucet) FundAccounts(ctx context.Context, accountsToFund ...coreumtesting.FundedAccount) (retErr error) {
	req := fundingRequest{
		AccountsToFund: accountsToFund,
		FundedCh:       make(chan error, 1),
	}

	requests := make([]fundingRequest, 0, 20)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case tf.queue <- req:
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-req.FundedCh:
			return err
		}
	case <-tf.muCh:
		defer func() {
			tf.muCh <- struct{}{}
			for _, req := range requests {
				req.FundedCh <- retErr
			}
		}()
	}

	requests = append(requests, req)
	numOfAccounts := len(req.AccountsToFund)
	timeout := time.After(100 * time.Millisecond)
loop:
	for numOfAccounts < cap(requests) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			break loop
		case req := <-tf.queue:
			requests = append(requests, req)
			numOfAccounts += len(req.AccountsToFund)
		}
	}

	messages := make([]sdk.Msg, 0, numOfAccounts)
	for _, req := range requests {
		for _, acc := range req.AccountsToFund {
			messages = append(messages, &banktypes.MsgSend{
				FromAddress: tf.address.String(),
				ToAddress:   acc.Wallet.Key.Address(),
				Amount:      sdk.NewCoins(acc.Amount),
			})
		}
	}

	log := logger.Get(ctx)
	log.Info("Funding accounts for test, it might take a while...")

	clientCtx := tf.chain.ClientContext.WithKeyring(tf.keyring).WithFromName("faucet").WithFromAddress(tf.address)
	resp, err := tx.BroadcastTx(
		ctx,
		clientCtx,
		tf.chain.TxFactory().WithKeybase(tf.keyring).WithGas(uint64(numOfAccounts)*tf.chain.GasLimitByMsgs(&banktypes.MsgSend{})),
		messages...,
	)
	if err != nil {
		return err
	}
	if _, err := tx.AwaitTx(ctx, clientCtx, resp.TxHash); err != nil {
		return err
	}
	log.Info("Test accounts funded")

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
