package config

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum/cored/app"
	"github.com/CoreumFoundation/coreum/cored/pkg/types"
	"github.com/pkg/errors"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	tmtypes "github.com/tendermint/tendermint/types"
)

// ChainID represents predefined chain-ids
type ChainID string

// Predefined chainIDs
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

var networks = map[ChainID]Network{
	Mainnet: {
		GenesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		ChainID:        Mainnet,
		AddressPrefix:  "core",
		TokenSymbol:    TokenSymbolMain,
		FundedAccounts: []fundedAccount{},
	},

	Devnet: {
		GenesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		ChainID:        Devnet,
		AddressPrefix:  "devcore",
		TokenSymbol:    TokenSymbolDev,
		FundedAccounts: []fundedAccount{},
	},
}

// Network holds all the configuration for different predefined networks
type Network struct {
	GenesisTime         time.Time
	ChainID             ChainID
	AddressPrefix       string
	TokenSymbol         string
	GenesisTransactions []json.RawMessage
	FundedAccounts      []fundedAccount
}

type fundedAccount struct {
	PubKey  types.Secp256k1PublicKey
	Balance string
}

// SetupPrefixes sets the global account prefixes config for cosmos sdk.
func (n Network) SetupPrefixes() {
	cosmoscmd.SetPrefixes(n.AddressPrefix)
}

// Genesis creates the genesis file for the given network config
func (n Network) Genesis() (*Genesis, error) {
	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	codec := encCfg.Marshaler
	genesis, err := genesis(n)
	if err != nil {
		return nil, err
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesis)
	if err != nil {
		return nil, err
	}
	var appState map[string]json.RawMessage
	err = json.Unmarshal(genesisDoc.AppState, &appState)
	if err != nil {
		return nil, err
	}

	authState := authtypes.GetGenesisStateFromAppState(codec, appState)
	accountState, err := authtypes.UnpackAccounts(authState.Accounts)
	if err != nil {
		return nil, err
	}
	g := &Genesis{
		codec:        codec,
		mu:           &sync.Mutex{},
		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(codec, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(codec, appState),
	}

	for _, fundedAccount := range n.FundedAccounts {
		err = g.FundAccount(fundedAccount.PubKey, fundedAccount.Balance)
		if err != nil {
			return nil, err
		}
	}

	return g, nil
}

// NetworkByChainID returns config for a predefined config.
// predefined networks are "coreum-mainnet" and "coreum-devnet".
func NetworkByChainID(id ChainID) (Network, error) {
	nw, found := networks[ChainID(id)]
	if !found {
		return Network{}, errors.Errorf("chainID %s not found", nw.ChainID)
	}

	return nw, nil
}
