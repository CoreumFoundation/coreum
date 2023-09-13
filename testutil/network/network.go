package network

import (
	"fmt"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/app"
	"github.com/CoreumFoundation/coreum/v3/pkg/config"
	"github.com/CoreumFoundation/coreum/v3/pkg/config/constant"
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

// FundedAccount is struct used for WithChainDenomFundedAccounts function.
type FundedAccount struct {
	Address sdk.AccAddress
	Amount  sdkmath.Int
}

// WithChainDenomFundedAccounts adds the funded account the config genesis.
func WithChainDenomFundedAccounts(fundedAccounts []FundedAccount) ConfigOption {
	return func(cfg network.Config) (network.Config, error) {
		genesisAppState := cfg.GenesisState

		var bankState banktypes.GenesisState
		cfg.Codec.MustUnmarshalJSON(genesisAppState[banktypes.ModuleName], &bankState)

		var authState authtypes.GenesisState
		cfg.Codec.MustUnmarshalJSON(genesisAppState[authtypes.ModuleName], &authState)

		for _, fundedAccount := range fundedAccounts {
			bankState.Balances = append(bankState.Balances, banktypes.Balance{
				Address: fundedAccount.Address.String(),
				Coins:   sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, fundedAccount.Amount)),
			})

			account := authtypes.NewBaseAccount(fundedAccount.Address, nil, 0, 0)
			packedAccounts, err := authtypes.PackAccounts(authtypes.GenesisAccounts{account})
			if err != nil {
				panic(errors.Wrap(err, "can pack genesis accounts"))
			}
			authState.Accounts = append(authState.Accounts, packedAccounts...)
		}

		genesisAppState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankState)
		genesisAppState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authState)

		return cfg, nil
	}
}

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig.
func DefaultConfig() network.Config {
	devNetwork, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(errors.Wrap(err, "can't get network config"))
	}
	// set to nil the devnet config we don't need
	provider := devNetwork.Provider.(config.DynamicConfigProvider)
	provider.FundedAccounts = nil
	provider.GenTxs = nil
	provider.CustomParamsConfig.Staking.MinSelfDelegation = sdkmath.NewInt(1)

	devNetwork.Provider = provider

	// init the network and set params
	app.ChosenNetwork = devNetwork
	// set and seal once
	setNetworkConfigOnce.Do(func() {
		devNetwork.SetSDKConfig()
	})
	appState, err := devNetwork.Provider.AppState()
	if err != nil {
		panic(errors.Wrap(err, "can't get network's app state"))
	}

	encoding := config.NewEncodingConfig(app.ModuleBasics)
	chainID := "chain-" + tmrand.NewRand().Str(6)
	return network.Config{
		Codec:             encoding.Codec,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		TxConfig:          encoding.TxConfig,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			return app.New(
				val.GetCtx().Logger,
				dbm.NewMemDB(),
				nil,
				true,
				simtestutil.NewAppOptionsWithFlagHome(app.DefaultNodeHome),
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
		MinGasPrices:   fmt.Sprintf("0.000006%s", devNetwork.Denom()),
		AccountTokens:  sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:  sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:   sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		CleanupDir:     true,
		SigningAlgo:    string(hd.Secp256k1Type),
		KeyringOptions: []keyring.Option{},
	}
}

// ApplyConfigOptions updates the simapp configuration with the provided ConfigOptions.
// We use the ApplyConfigOptions as separate function since the DefaultConfig set's the required global variables required for the ConfigOptions.
func ApplyConfigOptions(cfg network.Config, options ...ConfigOption) (network.Config, error) {
	for _, option := range options {
		var err error
		cfg, err = option(cfg)
		if err != nil {
			return network.Config{}, err
		}
	}

	return cfg, nil
}
