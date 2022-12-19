package integrationtests

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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
	RPCAddress      string
	NetworkConfig   config.NetworkConfig
	FundingMnemonic string
	StakerMnemonics []string
	LogFormat       logger.Format
	LogVerbose      bool
}

var (
	cfg   testingConfig
	chain Chain
)

func init() {
	var (
		fundingMnemonic, coredAddress, logFormat string
		stakerMnemonics                          stringsFlag
	)

	flag.StringVar(&coredAddress, "cored-address", "tcp://localhost:26657", "Address of cored node started by znet")
	flag.StringVar(&fundingMnemonic, "funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flag.Var(&stakerMnemonics, "staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")

	// accept testing flags
	testing.Init()
	// parse additional flags
	flag.Parse()

	// set the default staker mnemonic used in the dev znet by default
	if len(stakerMnemonics) == 0 {
		stakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
		}
	}

	networkConfig, err := NewNetworkConfig()
	if err != nil {
		panic(fmt.Sprintf("can't create network config for the integration tests: %s", err))
	}
	cfg = testingConfig{
		RPCAddress:      coredAddress,
		NetworkConfig:   networkConfig,
		FundingMnemonic: fundingMnemonic,
		StakerMnemonics: stakerMnemonics,
		LogFormat:       logger.Format(logFormat),
		LogVerbose:      flag.Lookup("test.v").Value.String() == "true",
	}

	// FIXME (wojtek): remove this once we have our own address encoder
	config.NewNetwork(cfg.NetworkConfig).SetSDKConfig()

	chain = NewChain(ChainConfig{
		RPCAddress:      cfg.RPCAddress,
		NetworkConfig:   cfg.NetworkConfig,
		FundingMnemonic: cfg.FundingMnemonic,
		StakerMnemonics: cfg.StakerMnemonics,
	})
}

// NewTestingContext returns the configured chain and new context for the integration tests.
func NewTestingContext(t *testing.T) (context.Context, Chain) {
	loggerConfig := logger.Config{
		Format:  cfg.LogFormat,
		Verbose: cfg.LogVerbose,
	}

	ctx, cancel := context.WithCancel(logger.WithLogger(context.Background(), logger.New(loggerConfig)))
	t.Cleanup(cancel)

	return ctx, chain
}
