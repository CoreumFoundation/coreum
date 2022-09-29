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

	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	tests "github.com/CoreumFoundation/coreum/integration-tests"
	coreumtesting "github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/config"
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

type testingConfig struct {
	RPCAddress      string
	NetworkConfig   config.NetworkConfig
	FundingPrivKey  types.Secp256k1PrivateKey
	StakerMnemonics []string
	Filter          *regexp.Regexp
	LogFormat       logger.Format
	LogVerbose      bool
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
	// set the default staker mnemonic used in the dev znet by default
	if len(stakerMnemonics) == 0 {
		stakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
		}
	}

	decodedFundingPrivKey, err := base64.RawURLEncoding.DecodeString(fundingPrivKey)
	if err != nil {
		panic(err)
	}
	cfg.FundingPrivKey = decodedFundingPrivKey
	cfg.StakerMnemonics = stakerMnemonics
	cfg.RPCAddress = coredAddress
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

	chainCfg := coreumtesting.ChainConfig{
		RPCAddress:      cfg.RPCAddress,
		NetworkConfig:   cfg.NetworkConfig,
		FundingPrivKey:  cfg.FundingPrivKey,
		StakerMnemonics: cfg.StakerMnemonics,
	}
	chain := coreumtesting.NewChain(chainCfg)

	testCases := collectTestCases(chain, testSet, cfg.Filter)
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
