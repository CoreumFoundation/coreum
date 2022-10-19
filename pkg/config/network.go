package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"sync"
	"text/template"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcosmostypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/pkg/errors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// ChainID represents predefined chain ID
type ChainID string

// Predefined chain IDs
const (
	Mainnet ChainID = "coreum-mainnet-1"
	Devnet  ChainID = "coreum-devnet-1"
)

// EnableFakeUpgradeHandler is set to true during compilation to enable fake upgrade handler on devnet allowing us to test upgrade procedure.
// It is string, not bool, because -X flag supports strings only.
var EnableFakeUpgradeHandler string

// Known TokenSymbols
const (
	// u (μ) prefix stands for micro, more info here https://en.wikipedia.org/wiki/Metric_prefix
	// We also add another prefix for non mainnet network symbols to differentiate them from mainnet.
	// 'd' prefix in ducore stands for devnet.
	TokenSymbolMain string = "ucore"
	TokenSymbolDev  string = "ducore"
)

const (
	// CoinType is the CORE coin type as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	CoinType uint32 = 990
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
		DeterministicGas: DefaultDeterministicGasRequirements(),
	}

	govConfig := GovConfig{
		ProposalConfig: GovProposalConfig{
			MinDepositAmount: "10000000",
			MinDepositPeriod: "120h", // 5 days
			VotingPeriod:     "120h", // 5 days
		},
	}

	stakingConfig := StakingConfig{
		UnbondingTime: "168h", // 7 days
		MaxValidators: 32,
	}

	list := []NetworkConfig{
		{
			ChainID:       Mainnet,
			Enabled:       false,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "core",
			TokenSymbol:   TokenSymbolMain,
			Fee:           feeConfig,
			GovConfig:     govConfig,
			StakingConfig: stakingConfig,
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
			GovConfig:     govConfig,
			StakingConfig: stakingConfig,
			FundedAccounts: []FundedAccount{
				// Staker of validator 0
				{
					PublicKey: &cosmossecp256k1.PubKey{Key: []byte{0x2, 0x2f, 0xae, 0x96, 0x14, 0xe, 0x4e, 0x4e, 0xfc, 0x42, 0xaa, 0xce, 0xc2, 0xbf, 0x72, 0x49, 0xd6, 0x50, 0xf8, 0xde, 0x85, 0xe4, 0xfc, 0xe4, 0x45, 0x4e, 0xcb, 0xb1, 0x85, 0xc0, 0xdb, 0x81, 0xa5}},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 1
				{
					PublicKey: &cosmossecp256k1.PubKey{Key: []byte{0x2, 0x64, 0xfd, 0xa6, 0x29, 0xc4, 0x89, 0x7b, 0xcf, 0x9b, 0xa6, 0x1f, 0xd9, 0xbe, 0xae, 0x61, 0x20, 0x49, 0xfd, 0x93, 0xb6, 0x3, 0xa5, 0xab, 0xe8, 0xdf, 0x6, 0xe0, 0xcf, 0x61, 0xd1, 0x8d, 0xa7}},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 2
				{
					PublicKey: &cosmossecp256k1.PubKey{Key: []byte{0x2, 0x68, 0x60, 0xc0, 0xa3, 0xcf, 0x14, 0x8c, 0xb, 0xdd, 0xd5, 0xe0, 0xbf, 0xf1, 0xb5, 0x3d, 0xd7, 0xee, 0x0, 0xf9, 0xab, 0x61, 0xd9, 0xa5, 0x82, 0x6f, 0x56, 0x21, 0x7, 0x50, 0x60, 0xd8, 0xd0}},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Staker of validator 3
				{
					PublicKey: &cosmossecp256k1.PubKey{Key: []byte{0x3, 0x93, 0xa9, 0x5b, 0xd4, 0x80, 0xa9, 0x1c, 0x6, 0xe6, 0x5d, 0xc7, 0xdd, 0x9c, 0xa4, 0xf6, 0x97, 0xfc, 0xd, 0x6b, 0x83, 0xb1, 0x37, 0x1c, 0xf9, 0x75, 0x68, 0xd3, 0x3c, 0x24, 0x85, 0xe6, 0x94}},
					Balances:  "10000000000000" + TokenSymbolDev,
				},

				// Faucet's account storing the rest of total supply
				{
					PublicKey: &cosmossecp256k1.PubKey{Key: []byte{0x2, 0x5b, 0xb9, 0x1c, 0x57, 0xec, 0x12, 0x10, 0x92, 0x58, 0xef, 0xf9, 0x5, 0x7b, 0x70, 0x9d, 0x96, 0xbb, 0x57, 0xc5, 0xaa, 0x38, 0x61, 0x60, 0xca, 0xb2, 0x9, 0x21, 0xf, 0x45, 0x32, 0xc6, 0x6b}},
					Balances:  "10000000000000" + TokenSymbolDev,
				},
			},
			GenTxs: []json.RawMessage{
				coreumDevnet1Validator0,
				coreumDevnet1Validator1,
				coreumDevnet1Validator2,
				coreumDevnet1Validator3,
			},
			EnableFakeUpgradeHandler: EnableFakeUpgradeHandler != "",
		},
	}

	for _, elem := range list {
		networks[elem.ChainID] = elem
	}
}

var networks = map[ChainID]NetworkConfig{}

// EnabledNetworks returns enabled networks
func EnabledNetworks() []Network {
	enabledNetworks := make([]Network, 0, len(networks))
	for _, nc := range networks {
		if nc.Enabled {
			enabledNetworks = append(enabledNetworks, NewNetwork(nc))
		}
	}
	return enabledNetworks
}

// FeeConfig is the part of network config defining parameters of our fee model
type FeeConfig struct {
	FeeModel         feemodeltypes.Model
	DeterministicGas DeterministicGasRequirements
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

// StakingConfig contains staking module configuration
type StakingConfig struct {
	// UnbondingTime is the time duration after which bonded coins will become to be released
	UnbondingTime string

	// MaxValidators is the maximum number of validators that could be created
	MaxValidators int
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
	StakingConfig  StakingConfig
	// TODO: remove this field once all preconfigured networks are enabled
	Enabled bool
	// TODO: remove this field once we have real upgrade handler
	EnableFakeUpgradeHandler bool
}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID                  ChainID
	genesisTime              time.Time
	addressPrefix            string
	tokenSymbol              string
	fee                      FeeConfig
	nodeConfig               NodeConfig
	gov                      GovConfig
	staking                  StakingConfig
	enableFakeUpgradeHandler bool

	mu             *sync.Mutex
	fundedAccounts []FundedAccount
	genTxs         []json.RawMessage
}

// NewNetwork returns a new instance of Network
func NewNetwork(c NetworkConfig) Network {
	n := Network{
		genesisTime:              c.GenesisTime,
		chainID:                  c.ChainID,
		addressPrefix:            c.AddressPrefix,
		tokenSymbol:              c.TokenSymbol,
		nodeConfig:               c.NodeConfig.Clone(),
		fee:                      c.Fee,
		gov:                      c.GovConfig,
		staking:                  c.StakingConfig,
		mu:                       &sync.Mutex{},
		fundedAccounts:           append([]FundedAccount{}, c.FundedAccounts...),
		genTxs:                   append([]json.RawMessage{}, c.GenTxs...),
		enableFakeUpgradeHandler: c.EnableFakeUpgradeHandler,
	}

	return n
}

// FundedAccount is used to provide information about pre funded
// accounts in network config
// TODO(dhil) refactor to use the address instead of PublicKey.
type FundedAccount struct {
	PublicKey cryptotypes.PubKey
	Balances  string
}

// FundAccount funds address with balances at genesis
func (n *Network) FundAccount(publicKey cryptotypes.PubKey, balances string) error {
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
	accountState authcosmostypes.GenesisAccounts,
	bankState *banktypes.GenesisState,
) (authcosmostypes.GenesisAccounts, error) {
	accountAddress := sdk.AccAddress(fa.PublicKey.Address())
	accountState = append(accountState, authcosmostypes.NewBaseAccount(accountAddress, nil, 0, 0))
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
	codec := NewEncodingConfig(module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
	)).Codec

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

	authState := authcosmostypes.GetGenesisStateFromAppState(codec, appState)
	accountState, err := authcosmostypes.UnpackAccounts(authState.Accounts)
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
	authState.Accounts, err = authcosmostypes.PackAccounts(authcosmostypes.SanitizeGenesisAccounts(accountState))
	if err != nil {
		return nil, errors.Wrap(err, "not able to sanitize and pack accounts")
	}
	appState[authcosmostypes.ModuleName] = codec.MustMarshalJSON(&authState)

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

// SetSDKConfig sets global SDK config to some network-specific values.
// In typical applications this func should be called right after network initialization.
func (n Network) SetSDKConfig() {
	config := sdk.GetConfig()

	// Set address & public key prefixes
	config.SetBech32PrefixForAccount(n.addressPrefix, n.addressPrefix+"pub")
	config.SetBech32PrefixForValidator(n.addressPrefix+"valoper", n.addressPrefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(n.addressPrefix+"valcons", n.addressPrefix+"valconspub")

	// Set BIP44 coin type corresponding to CORE
	config.SetCoinType(CoinType)

	config.Seal()
}

// AddressPrefix returns the address prefix to be used in network config
func (n Network) AddressPrefix() string {
	return n.addressPrefix
}

// ChainID returns the chain ID used in network config
func (n Network) ChainID() ChainID {
	return n.chainID
}

// GenesisTime returns the genesis time of the network
func (n Network) GenesisTime() time.Time {
	return n.genesisTime
}

// FundedAccounts returns the funded accounts
func (n Network) FundedAccounts() []FundedAccount {
	n.mu.Lock()
	defer n.mu.Unlock()

	fundedAccounts := make([]FundedAccount, len(n.fundedAccounts))
	copy(fundedAccounts, n.fundedAccounts)
	return fundedAccounts
}

// GenTxs returns the genesis transactions
func (n Network) GenTxs() []json.RawMessage {
	n.mu.Lock()
	defer n.mu.Unlock()

	genTxs := make([]json.RawMessage, len(n.genTxs))
	copy(genTxs, n.genTxs)
	return genTxs
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

// EnableFakeUpgradeHandler enables temporry fake upgrade handler until we have real one
func (n Network) EnableFakeUpgradeHandler() bool {
	return n.enableFakeUpgradeHandler
}

// DeterministicGas returns deterministic gas amounts required by some message types
func (n Network) DeterministicGas() DeterministicGasRequirements {
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
		FeeModelParams feemodeltypes.ModelParams
		Gov            GovConfig
		Staking        StakingConfig
	}{
		GenesisTimeUTC: n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.chainID,
		TokenSymbol:    n.tokenSymbol,
		FeeModelParams: n.FeeModel().Params(),
		Gov:            n.gov,
		Staking:        n.staking,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}

	return genesisBuf.Bytes(), nil
}
