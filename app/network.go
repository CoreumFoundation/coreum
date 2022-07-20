package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/types"

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

// DefaultNetwork is the network cored is configured to connect to
// FIXME (milad): Remove this hack once app loads appropriate network config based on CLI flag
var DefaultNetwork Network

func init() {
	list := []NetworkConfig{
		{
			ChainID:       Mainnet,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "core",
			TokenSymbol:   TokenSymbolMain,
			Fee: FeeConfig{
				InitialGasPrice:       big.NewInt(1500),
				MinDiscountedGasPrice: big.NewInt(1000),
				DeterministicGas: DeterministicGasConfig{
					BankSend: 120000,
				},
			},
		},
		{
			ChainID:       Devnet,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "devcore",
			TokenSymbol:   TokenSymbolDev,
			Fee: FeeConfig{
				InitialGasPrice:       big.NewInt(1500),
				MinDiscountedGasPrice: big.NewInt(1000),
				DeterministicGas: DeterministicGasConfig{
					BankSend: 120000,
				},
			},
		},
	}

	for _, elem := range list {
		networks[elem.ChainID] = elem
	}

	var err error
	DefaultNetwork, err = NetworkByChainID(Mainnet)
	if err != nil {
		panic(err)
	}
}

var networks = map[ChainID]NetworkConfig{}

// DeterministicGasConfig keeps config about deterministic gas for some message types
type DeterministicGasConfig struct {
	BankSend uint64
}

// FeeConfig is the part of network config defining parameters of our fee model
type FeeConfig struct {
	InitialGasPrice       *big.Int
	MinDiscountedGasPrice *big.Int
	DeterministicGas      DeterministicGasConfig
}

// NetworkConfig helps initialize Network instance
type NetworkConfig struct {
	ChainID        ChainID
	GenesisTime    time.Time
	AddressPrefix  string
	TokenSymbol    string
	Fee            FeeConfig
	FundedAccounts []FundedAccount
	GenTxs         []json.RawMessage
}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID       ChainID
	genesisTime   time.Time
	addressPrefix string
	tokenSymbol   string
	fee           FeeConfig

	mu             *sync.Mutex
	fundedAccounts []FundedAccount
	genTxs         []json.RawMessage
}

// NewNetwork returns a new instance of Network
func NewNetwork(c NetworkConfig) Network {
	fee := c.Fee
	fee.InitialGasPrice = big.NewInt(0).Set(c.Fee.InitialGasPrice)
	fee.MinDiscountedGasPrice = big.NewInt(0).Set(c.Fee.MinDiscountedGasPrice)
	n := Network{
		genesisTime:    c.GenesisTime,
		chainID:        c.ChainID,
		addressPrefix:  c.AddressPrefix,
		tokenSymbol:    c.TokenSymbol,
		fee:            fee,
		mu:             &sync.Mutex{},
		fundedAccounts: append([]FundedAccount{}, c.FundedAccounts...),
		genTxs:         append([]json.RawMessage{}, c.GenTxs...),
	}

	return n
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

func applyFundedAccountToGenesis(
	fa FundedAccount,
	accountState authtypes.GenesisAccounts,
	bankState *banktypes.GenesisState,
) (authtypes.GenesisAccounts, error) {
	pubKey := cosmossecp256k1.PubKey{Key: fa.PublicKey}
	accountAddress := sdk.AccAddress(pubKey.Address())
	accountState = append(accountState, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))
	coins, err := sdk.ParseCoinsNormalized(fa.Balances)
	if err != nil {
		return nil, errors.Wrapf(err, "not able to parse balances %s", fa.Balances)
	}

	bankState.Balances = append(
		bankState.Balances,
		banktypes.Balance{Address: accountAddress.String(), Coins: coins},
	)
	bankState.Supply = bankState.Supply.Add(coins...)
	return accountState, nil
}

// EncodeGenesis returns the json encoded representation of the genesis file
func (n Network) EncodeGenesis() ([]byte, error) {
	codec := NewEncodingConfig().Marshaler
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

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, fundedAcc := range n.fundedAccounts {
		accountState, err = applyFundedAccountToGenesis(fundedAcc, accountState, bankState)
		if err != nil {
			return nil, err
		}
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
func (n Network) ChainID() ChainID {
	return n.chainID
}

// TokenSymbol returns the governance token symbol. This is different
// for each network(i.e mainnet, testnet, etc)
func (n Network) TokenSymbol() string {
	return n.tokenSymbol
}

// InitialGasPrice returns initial gas price used by the first block
func (n Network) InitialGasPrice() *big.Int {
	return big.NewInt(0).Set(n.fee.InitialGasPrice)
}

// MinDiscountedGasPrice returns minimum gas price after giving maximum discount
func (n Network) MinDiscountedGasPrice() *big.Int {
	return big.NewInt(0).Set(n.fee.MinDiscountedGasPrice)
}

// DeterministicGas returns deterministic gas amounts required by some message types
func (n Network) DeterministicGas() DeterministicGasConfig {
	return n.fee.DeterministicGas
}

// NetworkByChainID returns config for a predefined config.
// predefined networks are "coreum-mainnet" and "coreum-devnet".
func NetworkByChainID(id ChainID) (Network, error) {
	nw, found := networks[id]
	if !found {
		return Network{}, errors.Errorf("chainID %s not found", id)
	}

	return NewNetwork(nw), nil
}

//go:embed genesis/genesis.tmpl.json
var genesisTemplate string

func genesis(n Network) ([]byte, error) {
	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC string
		ChainID        ChainID
		TokenSymbol    string
	}{
		GenesisTimeUTC: n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.chainID,
		TokenSymbol:    n.tokenSymbol,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}
	return genesisBuf.Bytes(), nil
}
