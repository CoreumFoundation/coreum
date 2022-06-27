package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/CoreumFoundation/coreum/cored/pkg/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	ErrChainIDNotDefined = errors.New("chain-id not defined")
)

type chainID string

const (
	Mainnet chainID = "coreum-mainnet"
	Testnet chainID = "coreum-testnet"
	Devnet  chainID = "coreum-devnet"
)

const (
	TokenSymbol     string = "acore"
	TokenSymbolTest string = "tacore"
)

var networks = map[chainID]Network{
	Mainnet: {
		ChainID:       Mainnet,
		AddressPrefix: "core",
		TokenSymbol:   "acore",
	},
	Testnet: {
		ChainID:       Testnet,
		AddressPrefix: "tcore",
		TokenSymbol:   "tacore",
		FundedAccounts: []struct {
			PubKey  types.Secp256k1PublicKey
			Balance string
		}{
			{
				PubKey:  AlicePrivKey.PubKey(),
				Balance: initialBalance,
			},
			{
				PubKey:  BobPrivKey.PubKey(),
				Balance: initialBalance,
			},
			{
				PubKey:  CharliePrivKey.PubKey(),
				Balance: initialBalance,
			},
		},
	},
	Devnet: {
		ChainID:       Devnet,
		AddressPrefix: "tcore",
		TokenSymbol:   "tacore",
		FundedAccounts: []struct {
			PubKey  types.Secp256k1PublicKey
			Balance string
		}{
			{
				PubKey:  AlicePrivKey.PubKey(),
				Balance: initialBalance,
			},
			{
				PubKey:  BobPrivKey.PubKey(),
				Balance: initialBalance,
			},
			{
				PubKey:  CharliePrivKey.PubKey(),
				Balance: initialBalance,
			},
		},
	},
}

type Network struct {
	ChainID             chainID
	AddressPrefix       string
	TokenSymbol         string
	GenesisTransactions []json.RawMessage
	FundedAccounts      []struct {
		PubKey  types.Secp256k1PublicKey
		Balance string
	}
}

// SetupPrefixes sets the global account prefixes config for cosmos sdk.
func (n Network) SetupPrefixes() {
	cosmoscmd.SetPrefixes(n.AddressPrefix)
}

func (n Network) GetGenesis() (*Genesis, error) {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	codec := codec.NewProtoCodec(interfaceRegistry)
	genesis, err := genesis(n.ChainID)
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
		mu:           &sync.Mutex{},
		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(codec, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(codec, appState),
	}

	for _, fundedAccount := range n.FundedAccounts {
		g.FundAccount(fundedAccount.PubKey, fundedAccount.Balance)
	}

	return g, nil
}

// GetNetworkByChainID returns config for a predefined config.
// predefined networks are "coreum-mainnet", "coreum-testnet" and "coreum-devnet".
func GetNetworkByChainID(id string) (Network, error) {
	network, found := networks[chainID(id)]
	if !found {
		return Network{}, fmt.Errorf("chain-id: %s, err: %w", id, ErrChainIDNotDefined)
	}

	return network, nil
}
