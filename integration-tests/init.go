package integrationtests

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
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
	GRPCAddress     string
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
		chainID                                  string
		stakerMnemonics                          stringsFlag
	)

	flag.StringVar(&coredAddress, "cored-address", "localhost:9090", "Address of cored node started by znet")
	flag.StringVar(&fundingMnemonic, "funding-mnemonic", "pitch basic bundle cause toe sound warm love town crucial divorce shell olympic convince scene middle garment glimpse narrow during fix fruit suffer honey", "Funding account mnemonic required by tests")
	flag.Var(&stakerMnemonics, "staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.StringVar(&chainID, "chain-id", string(constant.ChainIDDev), "Which chain-id to use (coreum-devnet-1, coreum-testnet-1,...)")

	// accept testing flags
	testing.Init()
	// parse additional flags
	flag.Parse()

	// set the default staker mnemonic used in the dev znet by default
	if len(stakerMnemonics) == 0 {
		stakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
			"enemy fix tribe swift alcohol metal salad edge episode dry tired address bless cloth error useful define rough fold swift confirm century wasp acoustic",
			"act electric demand cancel duck invest below once obvious estate interest solution drink mango reason already clean host limit stadium smoke census pattern express",
		}
	}

	networkConfig, err := NewNetworkConfig(constant.ChainID(chainID))
	if err != nil {
		panic(fmt.Sprintf("can't create network config for the integration tests: %s", err))
	}
	cfg = testingConfig{
		GRPCAddress:     coredAddress,
		NetworkConfig:   networkConfig,
		FundingMnemonic: fundingMnemonic,
		StakerMnemonics: stakerMnemonics,
		LogFormat:       logger.Format(logFormat),
		LogVerbose:      flag.Lookup("test.v").Value.String() == "true",
	}

	config.NewNetwork(cfg.NetworkConfig).SetSDKConfig()

	chain = NewChain(ChainConfig{
		GRPCAddress:     cfg.GRPCAddress,
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
