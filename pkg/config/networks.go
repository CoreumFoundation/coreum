package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/pkg/errors"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	tmjson "github.com/tendermint/tendermint/libs/json"
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
		New(
			Mainnet,
			time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			"core",
			TokenSymbolMain,
			[]FundedAccount{},
			[]json.RawMessage{},
		),
		New(
			Devnet,
			time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			"devcore",
			TokenSymbolDev,
			[]FundedAccount{},
			[]json.RawMessage{},
		),
	}

	for _, elem := range networksList {
		networks[elem.chainID] = elem
	}
}

var networks = map[ChainID]Network{}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID        ChainID
	genesisTime    time.Time
	addressPrefix  string
	tokenSymbol    string
	fundedAccounts []FundedAccount
	genTxs         []json.RawMessage
	mu             *sync.Mutex
}

// New returns a new instance of Network
func New(
	chainID ChainID,
	genesisTime time.Time,
	addressPrefix string,
	tokenSymbol string,
	fundedAccounts []FundedAccount,
	genTxs []json.RawMessage,
) Network {
	return Network{
		genesisTime:    genesisTime,
		chainID:        chainID,
		addressPrefix:  addressPrefix,
		tokenSymbol:    tokenSymbol,
		fundedAccounts: fundedAccounts,
		genTxs:         genTxs,
		mu:             &sync.Mutex{},
	}
}

// FundedAccount is used to provide information about pre funded
// accounts in network config
type FundedAccount struct {
	PublicKey types.Secp256k1PublicKey
	Balances  string
}

// FundAccount funds address with balances at genesis
func (n *Network) FundAccount(publicKey types.Secp256k1PublicKey, balances string) error {
	_, err := sdk.ParseCoinsNormalized(balances)
	if err != nil {
		return errors.Wrapf(err, "not able to parse balances %s", balances)
	}

	n.mu.Lock()
	defer n.mu.Unlock()
	n.fundedAccounts = append(n.fundedAccounts, FundedAccount{
		PublicKey: publicKey,
		Balances:  balances,
	})
	return nil
}

// AddGenesisTx adds transaction to the genesis file
func (n *Network) AddGenesisTx(signedTx json.RawMessage) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.genTxs = append(n.genTxs, signedTx)
}

// EncodeGenesis returns the json encoded representation of the genesis file
func (n Network) EncodeGenesis() ([]byte, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	codec := client.NewEncodingConfig().Marshaler
	genesisJSON, err := genesis(n)
	if err != nil {
		return nil, errors.Wrap(err, "not able get genesis")
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesisJSON)
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

	genutilState := genutiltypes.GetGenesisStateFromAppState(codec, appState)
	bankState := banktypes.GetGenesisStateFromAppState(codec, appState)

	for _, fundedAcc := range n.fundedAccounts {
		pubKey := cosmossecp256k1.PubKey{Key: fundedAcc.PublicKey}
		accountAddress := sdk.AccAddress(pubKey.Address())
		accountState = append(accountState, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))
		coins, err := sdk.ParseCoinsNormalized(fundedAcc.Balances)
		if err != nil {
			return nil, errors.Wrapf(err, "not able to parse balances %s", fundedAcc.Balances)
		}

		bankState.Balances = append(
			bankState.Balances,
			banktypes.Balance{Address: accountAddress.String(), Coins: coins},
		)
		bankState.Supply = bankState.Supply.Add(coins...)
	}

	genutilState.GenTxs = append(genutilState.GenTxs, n.genTxs...)

	genutiltypes.SetGenesisStateInAppState(codec, appState, genutilState)
	authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(accountState))
	if err != nil {
		return nil, errors.Wrap(err, "not able to sanitize and pack accounts")
	}
	appState[authtypes.ModuleName] = codec.MustMarshalJSON(&authState)

	bankState.Balances = banktypes.SanitizeGenesisBalances(bankState.Balances)
	appState[banktypes.ModuleName] = codec.MustMarshalJSON(bankState)

	genesisDoc.AppState, err = json.MarshalIndent(appState, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal app state")
	}

	bs, err := tmjson.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal genesis doc")
	}

	return bs, nil
}

// SaveGenesis saves json encoded representation of the genesis config into file
func (n Network) SaveGenesis(homeDir string) error {
	genDocBytes, err := n.EncodeGenesis()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(homeDir+"/config", 0o700); err != nil {
		return errors.Wrap(err, "unable to make config directory")
	}

	err = ioutil.WriteFile(homeDir+"/config/genesis.json", genDocBytes, 0644)
	return errors.Wrap(err, "unable to write genesis bytes to file")
}

// SetupPrefixes sets the global account prefixes config for cosmos sdk.
func (n Network) SetupPrefixes() {
	cosmoscmd.SetPrefixes(n.addressPrefix)
}

// AddressPrefix returns the address prefix to be used in network config
func (n Network) AddressPrefix() string {
	return n.addressPrefix
}

// ChainID returns the chain ID used in network config
func (n Network) ChainID() string {
	return string(n.chainID)
}

// TokenSymbol returns the governance token symbol. This is different
// for each network(i.e mainnet, testnet, etc)
func (n Network) TokenSymbol() string {
	return n.tokenSymbol
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
