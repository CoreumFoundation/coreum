package network

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	pruningtypes "cosmossdk.io/store/pruning/types"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/v5/app"
	"github.com/CoreumFoundation/coreum/v5/pkg/config"
	"github.com/CoreumFoundation/coreum/v5/pkg/config/constant"
)

type (
	// Network defines a local in-process testing network.
	Network = network.Network

	// Config defines the necessary configuration used to bootstrap and start an
	// in-process local testing network.
	Config = network.Config

	// ConfigOption option for the simapp configuration.
	ConfigOption func(cfg network.Config) (network.Config, error)
)

var setNetworkConfigOnce = sync.Once{}

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig(t)
	} else {
		cfg = configs[0]
	}

	var net *Network

	// Sometimes another process already binds the port to bind to.
	// So we need to retry to used another random port.
	// TODO: Remove the retry when upgrading to cosmos v0.52.x
	retryCtx, retryCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer retryCancel()
	err := retry.Do(retryCtx, 2*time.Second, func() error {
		n, err := network.New(t, t.TempDir(), cfg)
		if err != nil {
			if strings.Contains(err.Error(), "address already in use") {
				return retry.Retryable(err)
			}
			return err
		}
		net = n
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig.
func DefaultConfig(t *testing.T) network.Config {
	devNetwork, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(errors.Wrap(err, "can't get network config"))
	}
	// set to nil the devnet config we don't need
	provider := devNetwork.Provider.(config.DynamicConfigProvider)
	provider.BankBalances = nil
	provider.CustomParamsConfig.MinSelfDelegation = sdkmath.NewInt(1)

	devNetwork.Provider = provider

	// init the network and set params
	app.ChosenNetwork = devNetwork
	// set and seal once
	setNetworkConfigOnce.Do(func() {
		devNetwork.SetSDKConfig()
	})

	tempApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir()))

	clientCtx := cosmosclient.Context{}.
		WithCodec(tempApp.AppCodec()).
		WithInterfaceRegistry(tempApp.InterfaceRegistry()).
		WithTxConfig(tempApp.TxConfig())

	appState, err := devNetwork.Provider.AppState(context.Background(), clientCtx, tempApp.BasicModuleManager)
	if err != nil {
		panic(errors.Wrap(err, "can't get network's app state"))
	}

	chainID := "chain-" + tmrand.NewRand().Str(6)
	return network.Config{
		Codec:             tempApp.AppCodec(),
		LegacyAmino:       tempApp.LegacyAmino(),
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		TxConfig:          tempApp.TxConfig(),
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			return app.New(
				val.GetCtx().Logger,
				dbm.NewMemDB(),
				nil,
				true,
				simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(chainID),
			)
		},
		GenesisState:   appState,
		TimeoutCommit:  300 * time.Millisecond,
		ChainID:        chainID,
		NumValidators:  1,
		BondDenom:      devNetwork.Denom(),
		MinGasPrices:   "0.000006" + devNetwork.Denom(),
		AccountTokens:  sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:  sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:   sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		CleanupDir:     true,
		SigningAlgo:    string(hd.Secp256k1Type),
		KeyringOptions: []keyring.Option{},
	}
}

func tempDir() string {
	dir, err := os.MkdirTemp("", "cored")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir) //nolint:errcheck // we don't care

	return dir
}
