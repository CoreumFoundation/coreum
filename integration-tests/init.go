package integrationtests

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/app"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
	feemodeltypes "github.com/CoreumFoundation/coreum/v6/x/feemodel/types"
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

// Chains defines the all chains used for the tests.
type Chains struct {
	Coreum  integration.CoreumChain
	Gaia    integration.Chain
	Osmosis integration.Chain
}

var (
	ctx            context.Context
	chains         Chains
	chainsSyncOnce sync.Once
	runUnsafe      bool
)

// flag variables.
var (
	coreumGRPCAddress string
	coreumRPCAddress  string

	coreumFundingMnemonic string
	coreumStakerMnemonics stringsFlag

	gaiaGRPCAddress     string
	gaiaRPCAddress      string
	gaiaFundingMnemonic string

	osmosisGRPCAddress     string
	osmosisRPCAddress      string
	osmosisFundingMnemonic string
)

//nolint:lll // this function contains flag description and mnemonic which cannot be broken down.
func init() {
	flag.BoolVar(&runUnsafe, "run-unsafe", false, "run unsafe tests for example ones related to governance")

	flag.StringVar(&coreumGRPCAddress, "coreum-grpc-address", "localhost:9090", "GRPC address of cored node started by znet")
	flag.StringVar(&coreumRPCAddress, "coreum-rpc-address", "http://localhost:26657", "RPC address of cored node started by znet")
	flag.StringVar(&coreumFundingMnemonic, "coreum-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flag.Var(&coreumStakerMnemonics, "coreum-staker-mnemonic", "Staker account mnemonics required by tests, supports multiple")
	flag.StringVar(&gaiaGRPCAddress, "gaia-grpc-address", "localhost:9080", "GRPC address of gaia node started by znet")
	flag.StringVar(&gaiaRPCAddress, "gaia-rpc-address", "http://localhost:26557", "RPC address of gaia node started by znet")
	flag.StringVar(&gaiaFundingMnemonic, "gaia-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")
	flag.StringVar(&osmosisGRPCAddress, "osmosis-grpc-address", "localhost:9070", "GRPC address of osmosis node started by znet")
	flag.StringVar(&osmosisRPCAddress, "osmosis-rpc-address", "http://localhost:26457", "RPC address of osmosis node started by znet")
	flag.StringVar(&osmosisFundingMnemonic, "osmosis-funding-mnemonic", "sad hobby filter tray ordinary gap half web cat hard call mystery describe member round trend friend beyond such clap frozen segment fan mistake", "Funding account mnemonic required by tests")

	// accept testing flags
	testing.Init()
	// parse additional flags
	flag.Parse()

	ctx = context.Background()
	if !runUnsafe {
		ctx = integration.WithSkipUnsafe(ctx) //nolint:fatcontext // one level of nesting is okay
	}

	// set the default staker mnemonic used in the dev znet by default
	if len(coreumStakerMnemonics) == 0 {
		coreumStakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
			"enemy fix tribe swift alcohol metal salad edge episode dry tired address bless cloth error useful define rough fold swift confirm century wasp acoustic",
			"act electric demand cancel duck invest below once obvious estate interest solution drink mango reason already clean host limit stadium smoke census pattern express",
		}
	}

	queryCtx, queryCtxCancel := context.WithTimeout(ctx, integration.DefaultClientContextConfig().TimeoutConfig.RequestTimeout)
	defer queryCtxCancel()

	// ********** Coreum **********

	coreumGRPCClient, err := integration.DialGRPCClient(coreumGRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	coreumSettings := integration.QueryChainSettings(queryCtx, coreumGRPCClient)

	coreumClientCtx := client.NewContext(integration.DefaultClientContextConfig()).
		WithGRPCClient(coreumGRPCClient)

	coreumFeemodelParamsRes, err := feemodeltypes.
		NewQueryClient(coreumClientCtx).
		Params(queryCtx, &feemodeltypes.QueryParamsRequest{})
	if err != nil {
		panic(errors.WithStack(err))
	}
	coreumSettings.GasPrice = coreumFeemodelParamsRes.Params.Model.InitialGasPrice
	coreumSettings.CoinType = constant.CoinType
	coreumSettings.RPCAddress = coreumRPCAddress
	network, err := config.NetworkConfigByChainID(constant.ChainID(coreumSettings.ChainID))
	if err != nil {
		panic(errors.WithStack(err))
	}
	app.ChosenNetwork = network
	network.SetSDKConfig()

	coreumRPCClient, err := sdkclient.NewClientFromNode(coreumRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	chains.Coreum = integration.NewCoreumChain(integration.NewChain(
		coreumGRPCClient,
		coreumRPCClient,
		coreumSettings,
		coreumFundingMnemonic), coreumStakerMnemonics)
}

// NewCoreumTestingContext returns the configured coreum chain and new context for the integration tests.
func NewCoreumTestingContext(t *testing.T) (context.Context, integration.CoreumChain) {
	testCtx, testCtxCancel := context.WithCancel(ctx)
	t.Cleanup(testCtxCancel)

	return testCtx, chains.Coreum
}

// NewChainsTestingContext returns the configured chains and new context for the integration tests.
func NewChainsTestingContext(t *testing.T) (context.Context, Chains) {
	testCtx, testCtxCancel := context.WithCancel(ctx)
	t.Cleanup(testCtxCancel)

	chainsSyncOnce.Do(func() {
		queryCtx, queryCtxCancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
		defer queryCtxCancel()
		// ********** Gaia **********

		gaiaGRPClient, err := integration.DialGRPCClient(gaiaGRPCAddress)
		require.NoError(t, err)
		gaiaSettings := integration.QueryChainSettings(queryCtx, gaiaGRPClient)
		gaiaSettings.GasPrice = sdkmath.LegacyMustNewDecFromStr("1.0")
		gaiaSettings.CoinType = sdk.CoinType // gaia coin type
		gaiaSettings.RPCAddress = gaiaRPCAddress

		gaiaRPClient, err := sdkclient.NewClientFromNode(gaiaRPCAddress)
		if err != nil {
			panic(errors.WithStack(err))
		}

		chains.Gaia = integration.NewChain(
			gaiaGRPClient,
			gaiaRPClient,
			gaiaSettings,
			gaiaFundingMnemonic)

		// ********** Osmosis **********

		osmosisGRPClient, err := integration.DialGRPCClient(osmosisGRPCAddress)
		require.NoError(t, err)
		osmosisChainSettings := integration.QueryChainSettings(queryCtx, osmosisGRPClient)
		osmosisChainSettings.GasPrice = sdkmath.LegacyMustNewDecFromStr("0.01")
		osmosisChainSettings.CoinType = sdk.CoinType // osmosis coin type
		osmosisChainSettings.RPCAddress = osmosisRPCAddress

		osmosisRPClient, err := sdkclient.NewClientFromNode(osmosisRPCAddress)
		if err != nil {
			panic(errors.WithStack(err))
		}

		chains.Osmosis = integration.NewChain(
			osmosisGRPClient,
			osmosisRPClient,
			osmosisChainSettings,
			osmosisFundingMnemonic)
	})

	return testCtx, chains
}
