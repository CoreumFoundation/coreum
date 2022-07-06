package config

import (
	"encoding/json"
	"time"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/pkg/errors"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	tmtypes "github.com/tendermint/tendermint/types"
)

// ChainID represents predefined chain ID
type ChainID string

// Predefined chain IDs
const (
	Mainnet ChainID = "coreum-mainnet-1"
	Devnet  ChainID = "coreum-devnet-1"
)

// Known TokenSymbols
const (
	// TODO (milad): rename TokenSymbol to acore or attocore
	// naming is coming from https://en.wikipedia.org/wiki/Metric_prefix
	TokenSymbolMain string = "core"
	TokenSymbolDev  string = "dacore"
)

func init() {
	networksList := []Network{
		{
			chainID:        Mainnet,
			genesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			addressPrefix:  "core",
			tokenSymbol:    TokenSymbolMain,
			fundedAccounts: []fundedAccount{},
		},
		{
			chainID:        Devnet,
			genesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			addressPrefix:  "devcore",
			tokenSymbol:    TokenSymbolDev,
			fundedAccounts: []fundedAccount{},
		},
	}

	for _, elem := range networksList {
		networks[elem.chainID] = elem
	}
}

var networks = map[ChainID]Network{}

// Network holds all the configuration for different predefined networks
type Network struct {
	genesisTime    time.Time
	chainID        ChainID
	addressPrefix  string
	tokenSymbol    string
	fundedAccounts []fundedAccount
}

type fundedAccount struct {
	PubKey  types.Secp256k1PublicKey
	Balance string
}

// SetupPrefixes sets the global account prefixes config for cosmos sdk.
func (n Network) SetupPrefixes() {
	cosmoscmd.SetPrefixes(n.addressPrefix)
}

// AddressPrefix returns the address prefix to be used in network config
func (n Network) AddressPrefix() string {
	return n.addressPrefix
}

// Genesis creates the genesis file for the given network config
func (n Network) Genesis() (*Genesis, error) {
	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	codec := encCfg.Marshaler
	genesis, err := genesis(n)
	if err != nil {
		return nil, errors.Wrap(err, "not able get genesis")
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesis)
	if err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis json bytes")
	}
	var appState map[string]json.RawMessage

	if err = json.Unmarshal(genesisDoc.AppState, &appState); err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis app state")
	}

	authState := authtypes.GetGenesisStateFromAppState(codec, appState)
	accountState, err := authtypes.UnpackAccounts(authState.Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "not able to unpack auth accounts")
	}
	g := &Genesis{
		codec:        codec,
		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(codec, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(codec, appState),
	}

	for _, fundedAccount := range n.fundedAccounts {
		if err = g.FundAccount(fundedAccount.PubKey, fundedAccount.Balance); err != nil {
			return nil, errors.Wrap(err, "not able to fund account")
		}
	}

	return g, nil
}

// NetworkByChainID returns config for a predefined config.
// predefined networks are "coreum-mainnet" and "coreum-devnet".
func NetworkByChainID(id ChainID) (Network, error) {
	nw, found := networks[id]
	if !found {
		return Network{}, errors.Errorf("chainID %s not found", nw.chainID)
	}

	return nw, nil
}
