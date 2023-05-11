package integrationtests

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
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
	GaiaGRPCAddress string
	NetworkConfig   config.NetworkConfig
	FundingMnemonic string
	StakerMnemonics []string
	LogFormat       logger.Format
	LogVerbose      bool
	RunUnsafe       bool
}

var (
	ctx   context.Context
	cfg   testingConfig
	chain Chain
)

func init() {
	var (
		fundingMnemonic, coredAddress, logFormat string
		chainID                                  string
		stakerMnemonics                          stringsFlag
		runUnsafe                                bool
		gaiaAddress, gaiaChainID                 string
	)

	flag.StringVar(&coredAddress, "cored-address", "localhost:9090", "Address of cored node started by znet")
	flag.StringVar(&fundingMnemonic, "funding-mnemonic", "pitch basic bundle cause toe sound warm love town crucial divorce shell olympic convince scene middle garment glimpse narrow during fix fruit suffer honey", "Funding account mnemonic required by tests")
	flag.Var(&stakerMnemonics, "staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
	flag.StringVar(&logFormat, "log-format", string(logger.ToolDefaultConfig.Format), "Format of logs produced by tests")
	flag.StringVar(&chainID, "chain-id", string(constant.ChainIDDev), "Which chain-id to use (coreum-devnet-1, coreum-testnet-1,...)")
	flag.BoolVar(&runUnsafe, "run-unsafe", false, "run unsafe tests for example ones related to governance")
	flag.StringVar(&gaiaAddress, "gaia-address", "localhost:9080", "Address of gaia node started by znet")
	flag.StringVar(&gaiaChainID, "gaia-chain-id", "gaia-localnet-1", "gaia chain-id started by znet")

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
		GaiaGRPCAddress: gaiaAddress,
		NetworkConfig:   networkConfig,
		FundingMnemonic: fundingMnemonic,
		StakerMnemonics: stakerMnemonics,
		LogFormat:       logger.Format(logFormat),
		LogVerbose:      flag.Lookup("test.v").Value.String() == "true",
		RunUnsafe:       runUnsafe,
	}

	loggerConfig := logger.Config{
		Format:  cfg.LogFormat,
		Verbose: cfg.LogVerbose,
	}
	ctx = logger.WithLogger(context.Background(), logger.New(loggerConfig))

	cfg.NetworkConfig.SetSDKConfig()

	grpcClient, err := grpc.Dial(coredAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}
	clientCtx := client.NewContext(client.DefaultContextConfig(), app.ModuleBasics).
		WithChainID(string(cfg.NetworkConfig.ChainID())).
		WithKeyring(newConcurrentSafeKeyring(keyring.NewInMemory())).
		WithBroadcastMode(flags.BroadcastBlock).
		WithGRPCClient(grpcClient)

	gaiaGRPClient, err := grpc.Dial(gaiaAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaCtx := client.NewContext(client.DefaultContextConfig(), app.ModuleBasics).
		WithChainID(gaiaChainID).
		WithKeyring(newConcurrentSafeKeyring(keyring.NewInMemory())).
		WithBroadcastMode(flags.BroadcastBlock).
		WithGRPCClient(gaiaGRPClient)

	feemodelClient := feemodeltypes.NewQueryClient(clientCtx)

	ctx, cancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
	defer cancel()

	resp, err := feemodelClient.Params(ctx, &feemodeltypes.QueryParamsRequest{})
	if err != nil {
		panic(errors.WithStack(err))
	}

	chain = NewChain(ChainConfig{
		ClientContext:     clientCtx,
		GRPCAddress:       cfg.GRPCAddress,
		GaiaClientContext: gaiaCtx,
		NetworkConfig:     cfg.NetworkConfig,
		InitialGasPrice:   resp.Params.Model.InitialGasPrice,
		FundingMnemonic:   cfg.FundingMnemonic,
		StakerMnemonics:   cfg.StakerMnemonics,
	})
}

// NewTestingContext returns the configured chain and new context for the integration tests.
func NewTestingContext(t *testing.T) (context.Context, Chain) {
	ctx, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)

	return ctx, chain
}
