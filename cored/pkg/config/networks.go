package config

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum/cored/pkg/types"
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	// ErrChainIDNotDefined chain-id is not a predefined id
	ErrChainIDNotDefined = errors.New("chain-id not defined")
)

type chainID string

// Predefined chainIDs
const (
	Mainnet chainID = "coreum-mainnet"
	Devnet  chainID = "coreum-devnet"
)

// Known TokenSymbols
const (
	// TODO (milad): rename TokenSymbol to acore or attocore
	TokenSymbolMain string = "core"
	TokenSymbolTest string = "tacore"
)

var networks = map[chainID]network{
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
		AddressPrefix:  "tcore",
		TokenSymbol:    TokenSymbolTest,
		FundedAccounts: []fundedAccount{},
	},
}

// network holds all the configuration for different predefined networks
type network struct {
	GenesisTime         time.Time
	ChainID             chainID
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
func (n network) SetupPrefixes() {
	cosmoscmd.SetPrefixes(n.AddressPrefix)
}

// Genesis creates the genesis file for the given network config
func (n network) Genesis() (*Genesis, error) {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
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
func NetworkByChainID(id string) (network, error) {
	nw, found := networks[chainID(id)]
	if !found {
		return network{}, errors.Wrapf(ErrChainIDNotDefined, "chain-id: %s", id)
	}

	return nw, nil
}
