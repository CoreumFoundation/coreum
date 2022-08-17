package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"math/big"
	"os"
	"sync"
	"text/template"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"
	"github.com/pkg/errors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/types"
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

var (
	//go:embed networks/coreum-devnet-1/validator-0.json
	coreumDevnet1Validator0 json.RawMessage

	//go:embed networks/coreum-devnet-1/validator-1.json
	coreumDevnet1Validator1 json.RawMessage

	//go:embed networks/coreum-devnet-1/validator-2.json
	coreumDevnet1Validator2 json.RawMessage

	//go:embed networks/coreum-devnet-1/validator-3.json
	coreumDevnet1Validator3 json.RawMessage
)

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
			Enabled:       true,
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
			NodeConfig: NodeConfig{
				SeedPeers: []string{"4ae4593aff8dd5ececd217f273195549503e2df8@35.223.81.227:26656"},
			},
			FundedAccounts: []FundedAccount{
				// Staker of validator 0
				{
					PublicKey: types.Secp256k1PublicKey{0x2, 0x4d, 0x37, 0x41, 0xe3, 0xee, 0x24, 0xdf, 0xb1, 0x7b, 0xdb, 0xff, 0xcd, 0xc3, 0x51, 0xdc, 0x9f, 0xd1, 0x43, 0x19, 0x53, 0x4d, 0x7c, 0x33, 0x35, 0x5d, 0xf9, 0x8, 0x42, 0x58, 0xcb, 0x45, 0x59},
					Balances:  "100000000" + TokenSymbolDev,
				},

				// Staker of validator 1
				{
					PublicKey: types.Secp256k1PublicKey{0x2, 0xc1, 0x77, 0xad, 0xdf, 0xf4, 0x91, 0x13, 0x9f, 0xa8, 0xc0, 0x50, 0xd0, 0xba, 0x43, 0xce, 0xa5, 0x10, 0x3a, 0xd, 0xd0, 0xe7, 0x6e, 0xd8, 0x76, 0x94, 0x7f, 0xc0, 0x15, 0xe, 0xb4, 0x93, 0x56},
					Balances:  "100000000" + TokenSymbolDev,
				},

				// Staker of validator 2
				{
					PublicKey: types.Secp256k1PublicKey{0x3, 0x3e, 0x21, 0x76, 0x29, 0xba, 0x55, 0xd4, 0xad, 0xfe, 0x35, 0xf4, 0xcc, 0xcb, 0x96, 0x8e, 0x7f, 0x50, 0x53, 0xc0, 0x17, 0x15, 0x16, 0xab, 0x7c, 0x11, 0xd8, 0x44, 0x7a, 0xde, 0xdf, 0xd, 0x52},
					Balances:  "100000000" + TokenSymbolDev,
				},

				// Staker of validator 3
				{
					PublicKey: types.Secp256k1PublicKey{0x2, 0xbb, 0xfb, 0xad, 0xa6, 0xd, 0x8c, 0xfd, 0x9f, 0x6a, 0x24, 0xc4, 0xb8, 0xd0, 0x2c, 0x84, 0xb4, 0x57, 0x4b, 0xd8, 0xc1, 0x8b, 0xda, 0xf9, 0x18, 0x21, 0x63, 0xde, 0x7b, 0xfd, 0x5d, 0xf5, 0xe7},
					Balances:  "100000000" + TokenSymbolDev,
				},

				// Faucet's account storing the rest of total supply
				// FIXME (wojciech): generate new key once faucet is done
				{
					PublicKey: types.Secp256k1PublicKey{0x2, 0xf4, 0x17, 0xba, 0x2d, 0x78, 0x39, 0x29, 0xe3, 0x15, 0xd1, 0x71, 0xac, 0x79, 0x32, 0xe6, 0x8, 0xb1, 0xf6, 0x90, 0x34, 0xc5, 0x8c, 0xcf, 0xed, 0x55, 0x2e, 0xd6, 0x79, 0x2d, 0x40, 0xc6, 0xe9},
					Balances:  "499999999999999999600000000" + TokenSymbolDev,
				},
			},
			GenTxs: []json.RawMessage{
				coreumDevnet1Validator0,
				coreumDevnet1Validator1,
				coreumDevnet1Validator2,
				coreumDevnet1Validator3,
			},
		},
	}

	for _, elem := range list {
		networks[elem.ChainID] = elem
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

// Clone creates a copy of FeeConfig to allow to pass by reference
func (f FeeConfig) Clone() FeeConfig {
	return FeeConfig{
		InitialGasPrice:       big.NewInt(0).Set(f.InitialGasPrice),
		MinDiscountedGasPrice: big.NewInt(0).Set(f.MinDiscountedGasPrice),
		DeterministicGas: DeterministicGasConfig{
			BankSend: f.DeterministicGas.BankSend,
		},
	}
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
	NodeConfig     NodeConfig
	// TODO: remove this field once all preconfigured networks are enabled
	Enabled bool
}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID       ChainID
	genesisTime   time.Time
	addressPrefix string
	tokenSymbol   string
	fee           FeeConfig
	nodeConfig    NodeConfig

	mu             *sync.Mutex
	fundedAccounts []FundedAccount
	genTxs         []json.RawMessage
}

// NewNetwork returns a new instance of Network
func NewNetwork(c NetworkConfig) Network {
	n := Network{
		genesisTime:    c.GenesisTime,
		chainID:        c.ChainID,
		addressPrefix:  c.AddressPrefix,
		tokenSymbol:    c.TokenSymbol,
		nodeConfig:     c.NodeConfig.Clone(),
		fee:            c.Fee.Clone(),
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

// NodeConfig returns NodeConfig
func (n *Network) NodeConfig() *NodeConfig {
	nodeConfig := n.nodeConfig.Clone()
	return &nodeConfig
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

// genesisDoc returns the genesis doc of the network
func (n Network) genesisDoc() (*tmtypes.GenesisDoc, error) {
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
		return nil, err
	}

	return genesisDoc, nil
}

// EncodeGenesis returns the json encoded representation of the genesis file
func (n Network) EncodeGenesis() ([]byte, error) {
	genesisDoc, err := n.genesisDoc()
	if err != nil {
		return nil, errors.Wrap(err, "not able to get genesis doc")
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

	err = os.WriteFile(homeDir+"/config/genesis.json", genDocBytes, 0644)
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

	// TODO: remove this check once all preconfigured networks are enabled
	if !nw.Enabled {
		return Network{}, errors.Errorf("%s is not yet ready, use --chain-id=%s for devnet", id, string(Devnet))
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
