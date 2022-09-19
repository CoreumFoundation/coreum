package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"sync"
	"text/template"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/pkg/errors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/x/auth"
	"github.com/CoreumFoundation/coreum/x/auth/ante"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
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
	// m prefix stands for milli, more info here https://en.wikipedia.org/wiki/Metric_prefix
	TokenSymbolMain string = "ucore"
	// d prefix stands for development
	TokenSymbolDev string = "ducore"
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
	feeConfig := FeeConfig{
		FeeModel:         feemodeltypes.DefaultModel(),
		DeterministicGas: auth.DefaultDeterministicGasRequirements(),
	}

	govConfig := GovConfig{
		ProposalConfig: GovProposalConfig{
			MinDepositAmount: "10000000",
			MinDepositPeriod: "172800s",
			VotingPeriod:     "172800s",
		},
	}

	list := []NetworkConfig{
		{
			ChainID:       Mainnet,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "core",
			TokenSymbol:   TokenSymbolMain,
			Fee:           feeConfig,
			GovConfig:     govConfig,
		},
		{
			ChainID:       Devnet,
			Enabled:       true,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "devcore",
			TokenSymbol:   TokenSymbolDev,
			Fee:           feeConfig,
			NodeConfig: NodeConfig{
				SeedPeers: []string{"602df7489bd45626af5c9a4ea7f700ceb2222b19@35.223.81.227:26656"},
			},
			GovConfig: govConfig,
			FundedAccounts: []FundedAccount{
				// Staker of validator 0
				{
					PublicKey: types.Secp256k1PublicKey{0x3, 0x4f, 0x9e, 0x3a, 0xb3, 0xce, 0xb1, 0x7, 0x15, 0xdd, 0x24, 0x62, 0xb8, 0xfb, 0xeb, 0xf3, 0x83, 0x4b, 0x17, 0xa8, 0x2, 0xaf, 0x5f, 0x36, 0xf, 0x2d, 0x2a, 0x4a, 0x5c, 0x16, 0x36, 0xf6, 0xba},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 1
				{
					PublicKey: types.Secp256k1PublicKey{0x3, 0x7a, 0x2c, 0x1a, 0x73, 0x2c, 0xbd, 0x5f, 0x15, 0xba, 0x5, 0xa8, 0x79, 0x40, 0xd4, 0xb1, 0x5e, 0xda, 0x57, 0x47, 0xb7, 0x3f, 0x6f, 0xec, 0xd0, 0x89, 0x44, 0x1, 0xc3, 0xde, 0x24, 0x89, 0xbb},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 2
				{
					PublicKey: types.Secp256k1PublicKey{0x2, 0x42, 0x34, 0x5, 0x85, 0x3a, 0x35, 0x97, 0xb3, 0x1a, 0x3d, 0xb8, 0x9b, 0xa0, 0x80, 0x60, 0x36, 0xb3, 0x35, 0x48, 0xe3, 0x9c, 0xd4, 0xf9, 0x32, 0xf7, 0x7, 0xe6, 0x86, 0xf3, 0xb7, 0x65, 0x86},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 3
				{
					PublicKey: types.Secp256k1PublicKey{0x3, 0x46, 0xd1, 0x99, 0x1c, 0x15, 0x97, 0x36, 0x5e, 0x47, 0x70, 0x89, 0x6b, 0xb, 0xba, 0x8c, 0x50, 0x6d, 0xa7, 0x5d, 0x45, 0xbe, 0x9c, 0x9b, 0x47, 0x1b, 0x42, 0xb4, 0xda, 0xd6, 0xd9, 0x9a, 0x3e},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Faucet's account storing the rest of total supply
				{
					PublicKey: types.Secp256k1PublicKey{0x3, 0xda, 0xe4, 0x29, 0xe4, 0xe6, 0x50, 0xbe, 0x1e, 0xc7, 0xe3, 0x64, 0x5c, 0x71, 0x23, 0x9e, 0xef, 0xf5, 0xc8, 0x6, 0x86, 0x1a, 0x62, 0x9b, 0x85, 0xed, 0x7a, 0x49, 0x3f, 0x3a, 0x2a, 0x71, 0x6d},
					Balances:  "10000000000000" + TokenSymbolDev,
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

// FeeConfig is the part of network config defining parameters of our fee model
type FeeConfig struct {
	FeeModel         feemodeltypes.Model
	DeterministicGas ante.DeterministicGasRequirements
}

// GovConfig contains gov module configs
type GovConfig struct {
	ProposalConfig GovProposalConfig
}

// GovProposalConfig contains gov module proposal-related configuration options
type GovProposalConfig struct {
	// MinDepositAmount is the minimum amount needed to create a proposal. Basically anti-spam policy.
	MinDepositAmount string

	// MinDepositPeriod is the minimum deposit period. Basically the duration when a proposal depositing is available.
	MinDepositPeriod string

	// VotingPeriod is the proposal voting period duration.
	VotingPeriod string
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
	GovConfig      GovConfig
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
	gov           GovConfig

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
		fee:            c.Fee,
		gov:            c.GovConfig,
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
	codec := NewEncodingConfig().Codec
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
	SetPrefixes(n.addressPrefix)
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

// FeeModel returns fee model configuration
func (n Network) FeeModel() feemodeltypes.Model {
	return n.fee.FeeModel
}

// DeterministicGas returns deterministic gas amounts required by some message types
func (n Network) DeterministicGas() ante.DeterministicGasRequirements {
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
		FeeModelParams feemodeltypes.Params
		Gov            GovConfig
	}{
		GenesisTimeUTC: n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.chainID,
		TokenSymbol:    n.tokenSymbol,
		FeeModelParams: n.FeeModel().Params(),
		Gov:            n.gov,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}

	return genesisBuf.Bytes(), nil
}
