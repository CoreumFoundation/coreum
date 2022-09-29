//go:build integration
// +build integration

package tests_test

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
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

// stringsFlag allows setting a value multiple times to collect a list, as in -I=val1 -I=val2.
type stringsFlag []string

func (m *stringsFlag) String() string {
	if len(*m) == 0 {
		return ""
	}
	return fmt.Sprint(*m)
}

func (m *stringsFlag) Set(val string) error {
	*m = append(*m, val)
	return nil
}

var cfg = testingConfig{
	NetworkConfig: coreumtesting.NetworkConfig,
}

func TestMain(m *testing.M) {
	var fundingPrivKey, fundingMnemonic, coredAddress, logFormat, filter string
	var stakerMnemonics stringsFlag

	flag.StringVar(&coredAddress, "cored-address", "tcp://localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&fundingPrivKey, "priv-key", "LPIPcUDVpp8Cn__g-YMntGts-DfDbd2gKTcgUgqSLfY", "Base64-encoded private key used to fund accounts required by tests")
	// TODO (dhil) those values are needed here for the backward compatibility of the crust, during the migration from priv keys to mnemonics
	flag.StringVar(&fundingMnemonic, "funding-mnemonic", "", "Funding account mnemonic required by tests")
	flag.Var(&stakerMnemonics, "staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
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

	chain, err := newChain(cfg)
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

func newChain(cfg testingConfig) (coreumtesting.Chain, error) {
	coredClient := client.New(cfg.NetworkConfig.ChainID, cfg.CoredAddress)
	rpcClient, err := cosmosclient.NewClientFromNode(cfg.CoredAddress)
	if err != nil {
		panic(err)
	}

	clientContext := config.NewClientContext(app.ModuleBasics).
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient).
		WithBroadcastMode(flags.BroadcastBlock)

	chain := coreumtesting.Chain{
		Client:        coredClient,
		ClientContext: clientContext,
		NetworkConfig: cfg.NetworkConfig,
		Keyring:       keyring.NewInMemory(),
	}

	faucetPrivKey := &cosmossecp256k1.PrivKey{Key: cfg.FundingPrivKey}
	err = chain.Keyring.ImportPrivKey("faucet", crypto.EncryptArmorPrivKey(faucetPrivKey, "dummy", faucetPrivKey.Type()), "dummy")
	if err != nil {
		return coreumtesting.Chain{}, errors.WithStack(err)
	}

	faucet := newTestingFaucet(
		clientContext.WithFromName("faucet").WithFromAddress(sdk.AccAddress(faucetPrivKey.PubKey().Address())),
		chain,
	)
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

func newTestingFaucet(clientCtx cosmosclient.Context, chain coreumtesting.Chain) *testingFaucet {
	faucet := &testingFaucet{
		clientCtx: clientCtx,
		chain:     chain,
		queue:     make(chan fundingRequest),
		muCh:      make(chan struct{}, 1),
	}
	faucet.muCh <- struct{}{}
	return faucet
}

type testingFaucet struct {
	clientCtx cosmosclient.Context
	chain     coreumtesting.Chain
	queue     chan fundingRequest

	// muCh is used to serve the same purpose as `sync.Mutex` to protect `fundingWallet` against being used
	// to broadcast many transactions in parallel by different integration tests. The difference between this and `sync.Mutex`
	// is that test may exit immediately when `ctx` is canceled, without waiting for mutex to be unlocked.
	muCh chan struct{}
}

func (tf *testingFaucet) FundAccounts(ctx context.Context, accountsToFund ...coreumtesting.FundedAccount) (retErr error) {
	const maxAccountsPerRequest = 20

	if len(accountsToFund) > maxAccountsPerRequest {
		return errors.Errorf("the number of accounts to fund (%d) is greater than the allowed maximum (%d)", len(accountsToFund), maxAccountsPerRequest)
	}

	req := fundingRequest{
		AccountsToFund: accountsToFund,
		FundedCh:       make(chan error, 1),
	}

	// This `select` block is essential for understanding how the algorithm works.
	// It decides if the caller of the function is the leader of the transaction or just a regular participant.
	// There are 3 possible scenarios:
	// - `<-tf.muCh` succeeds - the caller becomes a leader of the transaction. Its responsibility is to collect requests from
	//    other participants, broadcast transaction and await it.
	// - `tf.queue <- req` succeeds - the caller becomes a participant and his request was accepted by the leader, accounts will be funded in coming block
	//   Caller waits until `<-req.FundedCh` succeeds, meaning that accounts were successfully funded or process failed.
	// - none of the above - meaning that current leader finished the process of collecting requests from participants and now
	//   transaction is broadcasted or awaited. Once it is finished `muCh` is unlocked and another caller will become a new leader
	//   accepting requests from other participants again.

	select {
	case <-ctx.Done():
		return ctx.Err()
	case tf.queue <- req:
		// There is a leader who accepted this request. Now we must wait for transaction to be included in a block.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-req.FundedCh:
			return err
		}
	case <-tf.muCh:
		// This call is a leader, it will collect requests from participants and broadcast transaction.
	}

	// Code below is executed by the leader.

	// This call may fail only because of cancelled context, so we don't need to propagate it to
	// other participants
	requests, err := tf.collectRequests(ctx, req)
	if err != nil {
		return err
	}

	defer func() {
		// After transaction is broadcasted we unlock `muCh` so another leader for next transaction might be selected
		tf.muCh <- struct{}{}

		// If leader got an error during broadcasting, that error is propagated to all the other participants.
		for _, req := range requests {
			req.FundedCh <- retErr
		}
	}()

	// All requests are collected, let's create messages and broadcast tx
	return tf.broadcastTx(ctx, tf.collectMessages(requests))
}

func (tf *testingFaucet) collectRequests(ctx context.Context, leaderReq fundingRequest) ([]fundingRequest, error) {
	const (
		requestsPerTx   = 20
		timeoutDuration = 100 * time.Millisecond
	)

	requests := make([]fundingRequest, 0, requestsPerTx)

	// Leader adds his own request to the batch
	requests = append(requests, leaderReq)

	// In the loop, we wait a moment to give other participants to join.
	timeout := time.After(timeoutDuration)
	for len(requests) < requestsPerTx {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			// We close the window when other participants might join the batch.
			// If someone comes after timeout they must wait for next leader.
			return requests, nil
		case req := <-tf.queue:
			// Request from other participant is accepted and added to the batch.
			requests = append(requests, req)
		}
	}
	return requests, nil
}

func (tf *testingFaucet) collectMessages(requests []fundingRequest) []sdk.Msg {
	var messages []sdk.Msg
	for _, req := range requests {
		for _, acc := range req.AccountsToFund {
			messages = append(messages, &banktypes.MsgSend{
				FromAddress: tf.clientCtx.FromAddress.String(),
				ToAddress:   acc.Wallet.Key.Address(),
				Amount:      sdk.NewCoins(acc.Amount),
			})
		}
	}
	return messages
}

func (tf *testingFaucet) broadcastTx(ctx context.Context, msgs []sdk.Msg) error {
	log := logger.Get(ctx)
	log.Info("Funding accounts for tests, it might take a while...")
	// FIXME (wojtek): use estimation once it is available in `tx` package
	gasLimit := uint64(len(msgs)) * tf.chain.GasLimitByMsgs(&banktypes.MsgSend{})

	// Transaction is broadcasted and awaited
	resp, err := tx.BroadcastTx(
		ctx,
		tf.clientCtx,
		tf.chain.TxFactory().WithGas(gasLimit),
		msgs...,
	)
	if err != nil {
		return err
	}
	if _, err := tx.AwaitTx(ctx, tf.clientCtx, resp.TxHash); err != nil {
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
