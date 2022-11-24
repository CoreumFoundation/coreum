package network

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmdb "github.com/tendermint/tm-db"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

type (
	// Network defines a local in-process testing network
	Network = network.Network

	// Config defines the necessary configuration used to bootstrap and start an
	// in-process local testing network
	Config = network.Config

	// ConfigOption option for the simapp configuration.
	ConfigOption func(cfg network.Config) (network.Config, error)
)

var (
	setNetworkConfigOnce = sync.Once{}
)

// FundedAccount is struct used for WithChainDenomFundedAccounts function.
type FundedAccount struct {
	Address sdk.AccAddress
	Amount  sdk.Int
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
	net := network.New(t, cfg)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	devCfg, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(errors.Wrap(err, "can't get network config"))
	}
	// set to nil the devnet config we don't need
	devCfg.FundedAccounts = nil
	devCfg.GenTxs = nil

	// init the network and set params
	devNetwork := config.NewNetwork(devCfg)
	app.ChosenNetwork = devNetwork
	// set and seal once
	setNetworkConfigOnce.Do(func() {
		devNetwork.SetSDKConfig()
	})
	genesisDoc, err := devNetwork.GenesisDoc()
	if err != nil {
		panic(errors.Wrap(err, "can't get network genesis doc"))
	}

	var genesisAppState map[string]json.RawMessage
	if err = json.Unmarshal(genesisDoc.AppState, &genesisAppState); err != nil {
		panic(errors.Wrapf(err, "can unmarshal genesis app state to %T", genesisAppState))
	}

	encoding := config.NewEncodingConfig(app.ModuleBasics)

	return network.Config{
		Codec:             encoding.Codec,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.Validator) servertypes.Application {
			return app.New(
				val.Ctx.Logger, tmdb.NewMemDB(), nil, true, map[int64]bool{}, val.Ctx.Config.RootDir, 0,
				encoding,
				simapp.EmptyAppOptions{},
				baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
				baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
			)
		},
		GenesisState:    genesisAppState,
		TimeoutCommit:   2 * time.Second,
		ChainID:         "chain-" + tmrand.NewRand().Str(6),
		NumValidators:   1,
		BondDenom:       devCfg.Denom,
		MinGasPrices:    fmt.Sprintf("0.000006%s", devCfg.Denom),
		AccountTokens:   sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens:   sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:    sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),
		PruningStrategy: storetypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
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
