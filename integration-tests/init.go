package integrationtests

import (
	"context"
	"flag"
	"fmt"
	"testing"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	protobufgrpc "github.com/gogo/protobuf/grpc"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
	feemodeltypes "github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
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
	Coreum  CoreumChain
	Gaia    Chain
	Osmosis Chain
}

var (
	ctx       context.Context
	chains    Chains
	runUnsafe bool
)

func init() { //nolint:funlen // will be shortened after the crust merge
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

	// set the default staker mnemonic used in the dev znet by default
	if len(coreumStakerMnemonics) == 0 {
		coreumStakerMnemonics = []string{
			"biology rigid design broccoli adult hood modify tissue swallow arctic option improve quiz cliff inject soup ozone suffer fantasy layer negative eagle leader priority",
			"enemy fix tribe swift alcohol metal salad edge episode dry tired address bless cloth error useful define rough fold swift confirm century wasp acoustic",
			"act electric demand cancel duck invest below once obvious estate interest solution drink mango reason already clean host limit stadium smoke census pattern express",
		}
	}

	queryCtx, queryCtxCancel := context.WithTimeout(ctx, client.DefaultContextConfig().TimeoutConfig.RequestTimeout)
	defer queryCtxCancel()

	// ********** Coreum **********

	coreumGRPCClient, err := grpc.Dial(coreumGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}
	coreumSettings := queryCommonSettings(queryCtx, coreumGRPCClient)

	coreumClientCtx := client.NewContext(client.DefaultContextConfig(), app.ModuleBasics).
		WithGRPCClient(coreumGRPCClient)

	coreumFeemodelParamsRes, err := feemodeltypes.NewQueryClient(coreumClientCtx).Params(queryCtx, &feemodeltypes.QueryParamsRequest{})
	if err != nil {
		panic(errors.WithStack(err))
	}
	coreumSettings.GasPrice = coreumFeemodelParamsRes.Params.Model.InitialGasPrice
	coreumSettings.CoinType = constant.CoinType
	coreumSettings.RPCAddress = coreumRPCAddress

	config.SetSDKConfig(coreumSettings.AddressPrefix, constant.CoinType)

	coreumRPClient, err := sdkclient.NewClientFromNode(coreumRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	coreumChain := NewCoreumChain(NewChain(
		coreumGRPCClient,
		coreumRPClient,
		coreumSettings,
		coreumFundingMnemonic), coreumStakerMnemonics)

	// ********** Gaia **********

	gaiaGRPClient, err := grpc.Dial(gaiaGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaSettings := queryCommonSettings(queryCtx, gaiaGRPClient)
	gaiaSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	gaiaSettings.GasAdjustment = 1.5
	gaiaSettings.CoinType = sdk.CoinType // gaia coin type
	gaiaSettings.RPCAddress = gaiaRPCAddress

	gaiaRPClient, err := sdkclient.NewClientFromNode(gaiaRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	gaiaChain := NewChain(
		gaiaGRPClient,
		gaiaRPClient,
		gaiaSettings,
		gaiaFundingMnemonic)

	// ********** Osmosis **********

	osmosisGRPClient, err := grpc.Dial(osmosisGRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChainSettings := queryCommonSettings(queryCtx, osmosisGRPClient)
	osmosisChainSettings.GasPrice = sdk.MustNewDecFromStr("0.01")
	osmosisChainSettings.GasAdjustment = 1.5
	osmosisChainSettings.CoinType = sdk.CoinType // osmosis coin type
	osmosisChainSettings.RPCAddress = osmosisRPCAddress

	osmosisRPClient, err := sdkclient.NewClientFromNode(osmosisRPCAddress)
	if err != nil {
		panic(errors.WithStack(err))
	}

	osmosisChain := NewChain(
		osmosisGRPClient,
		osmosisRPClient,
		osmosisChainSettings,
		osmosisFundingMnemonic)

	chains = Chains{
		Coreum:  coreumChain,
		Gaia:    gaiaChain,
		Osmosis: osmosisChain,
	}
}

// NewCoreumTestingContext returns the configured coreum chain and new context for the integration tests.
func NewCoreumTestingContext(t *testing.T) (context.Context, CoreumChain) {
	testCtx, testCtxCancel := context.WithCancel(ctx)
	t.Cleanup(testCtxCancel)

	return testCtx, chains.Coreum
}

// NewChainsTestingContext returns the configured chains and new context for the integration tests.
func NewChainsTestingContext(t *testing.T) (context.Context, Chains) {
	testCtx, testCtxCancel := context.WithCancel(ctx)
	t.Cleanup(testCtxCancel)

	return testCtx, chains
}

func queryCommonSettings(ctx context.Context, grpcClient protobufgrpc.ClientConn) ChainSettings {
	clientCtx := client.NewContext(client.DefaultContextConfig(), app.ModuleBasics).
		WithGRPCClient(grpcClient)

	infoBeforeRes, err := tmservice.NewServiceClient(clientCtx).GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	if err != nil {
		panic(fmt.Sprintf("can't get node info, err: %s", err))
	}

	chainID := infoBeforeRes.DefaultNodeInfo.Network

	paramsRes, err := stakingtypes.NewQueryClient(clientCtx).Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		panic(fmt.Sprintf("can't get staking params, err: %s", err))
	}

	denom := paramsRes.Params.BondDenom

	accountsRes, err := authtypes.NewQueryClient(clientCtx).Accounts(ctx, &authtypes.QueryAccountsRequest{})
	if err != nil {
		panic(fmt.Sprintf("can't get account params, err: %s", err))
	}

	var addressPrefix string
	for _, account := range accountsRes.Accounts {
		if account != nil && account.TypeUrl == fmt.Sprintf("/%s", proto.MessageName(&authtypes.BaseAccount{})) {
			var acc authtypes.BaseAccount
			if err := proto.Unmarshal(account.Value, &acc); err != nil {
				panic(fmt.Sprintf("can't unpack account, err: %s", err))
			}

			addressPrefix, _, err = bech32.DecodeAndConvert(acc.Address)
			if err != nil {
				panic(fmt.Sprintf("can't extract address prefix address:%s, err: %s", acc.Address, err))
			}
			break
		}
	}
	if addressPrefix == "" {
		panic("address prefix is empty")
	}

	return ChainSettings{
		ChainID:       chainID,
		Denom:         denom,
		AddressPrefix: addressPrefix,
	}
}
