//go:build integration
// +build integration

package tests_test

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	tests "github.com/CoreumFoundation/coreum/integration-tests"
	coreumtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/config"
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

type testingConfig struct {
	RPCAddress             string
	UpgradeRPCAddress      string
	NetworkConfig          config.NetworkConfig
	FundingMnemonic        string
	StakerMnemonics        []string
	UpgradeStakerMnemonics []string
	Filter                 *regexp.Regexp
	LogFormat              logger.Format
	LogVerbose             bool
}

var cfg = testingConfig{
	NetworkConfig: coreumtesting.NetworkConfig,
}

func TestMain(m *testing.M) {
	var fundingMnemonic, coredAddress, coredUpgradeAddress, logFormat, filter string
	var stakerMnemonics, upgradeStakerMnemonics stringsFlag

	flag.StringVar(&coredAddress, "cored-address", "tcp://localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&coredUpgradeAddress, "cored-upgrade-address", "tcp://localhost:46657", "Address of cored node started by znet used to test upgrades")
	flag.StringVar(&fundingMnemonic, "funding-mnemonic", "", "Funding account mnemonic required by tests")
	flag.Var(&stakerMnemonics, "staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
	flag.Var(&upgradeStakerMnemonics, "upgrade-staker-mnemonic", "Staker account mnemonics required by upgrade tests, supports multiple")
	flag.StringVar(&filter, "filter", "", "Regular expression used to run only a subset of tests")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.Parse()
	// set the default staker mnemonic used in the dev znet by default
	if len(stakerMnemonics) == 0 {
		stakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
		}
	}

	cfg.FundingMnemonic = fundingMnemonic
	cfg.StakerMnemonics = stakerMnemonics
	cfg.UpgradeStakerMnemonics = upgradeStakerMnemonics
	cfg.RPCAddress = coredAddress
	cfg.UpgradeRPCAddress = coredUpgradeAddress
	cfg.Filter = regexp.MustCompile(filter)
	cfg.LogFormat = logger.Format(logFormat)
	cfg.LogVerbose = flag.Lookup("test.v").Value.String() == "true"

	// FIXME (wojtek): remove this once we have our own address encoder
	config.NewNetwork(cfg.NetworkConfig).SetSDKConfig()

	m.Run()
}

func Test(t *testing.T) {
	t.Parallel()

	testSet := tests.Tests()
	ctx := newContext(t, cfg)

	chain := coreumtesting.NewChain(coreumtesting.ChainConfig{
		RPCAddress:      cfg.RPCAddress,
		NetworkConfig:   cfg.NetworkConfig,
		FundingMnemonic: cfg.FundingMnemonic,
		StakerMnemonics: cfg.StakerMnemonics,
	})
	upgradeChain := coreumtesting.NewChain(coreumtesting.ChainConfig{
		RPCAddress:      cfg.UpgradeRPCAddress,
		NetworkConfig:   cfg.NetworkConfig,
		FundingMnemonic: cfg.FundingMnemonic,
		StakerMnemonics: cfg.UpgradeStakerMnemonics,
	})

	testCases := collectSingleChainTests(testSet.SingleChain, chain, cfg.Filter)
	testCases = append(testCases, collectSingleChainTests(testSet.UpgradeChain, upgradeChain, cfg.Filter)...)

	if len(testCases) == 0 {
		logger.Get(ctx).Warn("No tests to run")
		return
	}

	runTests(ctx, t, testCases)
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

func collectSingleChainTests(tests []coreumtesting.SingleChainSignature, chain coreumtesting.Chain, testFilter *regexp.Regexp) []testCase {
	testCases := make([]testCase, 0, len(tests))
	for _, testFunc := range tests {
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
