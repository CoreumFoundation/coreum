package config

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

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

// Predefined chain ids
const (
	ChainIDMain ChainID = "coreum-mainnet-1"
	ChainIDDev  ChainID = "coreum-devnet-1"
)

const (
	// CoinType is the CORE coin type as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	CoinType uint32 = 990
)

// EnableFakeUpgradeHandler is set to true during compilation to enable fake upgrade handler on devnet allowing us to test upgrade procedure.
// It is string, not bool, because -X flag supports strings only.
var EnableFakeUpgradeHandler string

var (
	//go:embed networks/coreum-devnet-1
	coreumDevnet1GenTxsFS embed.FS
)

func init() {
	// common vars
	var (
		feeConfig = FeeConfig{
			FeeModel:         feemodeltypes.DefaultModel(),
			DeterministicGas: DefaultDeterministicGasRequirements(),
		}

		govConfig = GovConfig{
			ProposalConfig: GovProposalConfig{
				MinDepositAmount: "10000000",
				VotingPeriod:     "120h", // 5 days
			},
		}

		stakingConfig = StakingConfig{
			UnbondingTime: "168h", // 7 days
			MaxValidators: 32,
		}
	)

	const denomDev = "ducore"

	// devnet vars
	var (
		// 10m delegated and 1m extra to the txs
		stakerValidatorBalance = sdk.NewCoins(sdk.NewCoin(denomDev, sdk.NewInt(11_000_000_000_000)))
	)

	list := []NetworkConfig{
		{
			ChainID:              ChainIDMain,
			Enabled:              false,
			GenesisTime:          time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix:        "core",
			MetadataDisplayDenom: "core",
			Denom:                "ucore",
			Fee:                  feeConfig,
			GovConfig:            govConfig,
			StakingConfig:        stakingConfig,
		},
		{
			ChainID:              ChainIDDev,
			Enabled:              true,
			GenesisTime:          time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix:        "devcore",
			MetadataDisplayDenom: "dcore",
			Denom:                denomDev,
			Fee:                  feeConfig,
			NodeConfig: NodeConfig{
				SeedPeers: []string{"602df7489bd45626af5c9a4ea7f700ceb2222b19@35.223.81.227:26656"},
			},
			GovConfig:     govConfig,
			StakingConfig: stakingConfig,
			FundedAccounts: []FundedAccount{
				// Staker of validator 0
				{
					Address:  "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6",
					Balances: stakerValidatorBalance,
				},
				// Staker of validator 1
				{
					Address:  "devcore105ct3vl89ar53jrj23zl6e09cmqwym2ua5hegf",
					Balances: stakerValidatorBalance,
				},
				// Staker of validator 2
				{
					Address:  "devcore14x46r5eflga696sd5my900euvlplu2prhny5ae",
					Balances: stakerValidatorBalance,
				},
				// Staker of validator 3
				{
					Address:  "devcore1xsthw036vst75rhh4py57lt7nx59qpvzez3a8k",
					Balances: stakerValidatorBalance,
				},
				// Faucet's account storing the rest of total supply
				{
					Address:  "devcore1ckuncyw0hftdq5qfjs6ee2v6z73sq0urd390cd",
					Balances: sdk.NewCoins(sdk.NewCoin(denomDev, sdk.NewInt(100_000_000_000_000))), // 100m faucet
				},
			},
			GenTxs:                      readGenTxs(coreumDevnet1GenTxsFS),
			IsFakeUpgradeHandlerEnabled: EnableFakeUpgradeHandler != "",
		},
	}

	for _, elem := range list {
		networkConfigs[elem.ChainID] = elem
	}
}

func readGenTxs(genTxsFs fs.FS) []json.RawMessage {
	genTxs := make([]json.RawMessage, 0)
	err := fs.WalkDir(genTxsFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic("can't open GenTxs FS")
		}
		if d.IsDir() {
			return nil
		}

		file, err := genTxsFs.Open(path)
		if err != nil {
			panic(fmt.Sprintf("can't open file %q from GenTxs FS", path))
		}
		defer file.Close()
		txBytes, err := io.ReadAll(file)
		if err != nil {
			panic(fmt.Sprintf("can't read file %+v from GenTxs FS", file))
		}
		genTxs = append(genTxs, txBytes)
		return nil
	})
	if err != nil {
		panic("can't read files from GenTxs FS")
	}

	return genTxs
}

var networkConfigs = map[ChainID]NetworkConfig{}

// EnabledNetworks returns enabled networks
func EnabledNetworks() []Network {
	enabledNetworks := make([]Network, 0, len(networkConfigs))
	for _, nc := range networkConfigs {
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
	ChainID              ChainID
	GenesisTime          time.Time
	AddressPrefix        string
	MetadataDisplayDenom string
	Denom                string
	Fee                  FeeConfig
	FundedAccounts       []FundedAccount
	GenTxs               []json.RawMessage
	NodeConfig           NodeConfig
	GovConfig            GovConfig
	StakingConfig        StakingConfig
	// TODO: remove this field once all preconfigured networks are enabled
	Enabled bool
	// TODO: remove this field once we have real upgrade handler
	IsFakeUpgradeHandlerEnabled bool
}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID                  ChainID
	genesisTime              time.Time
	addressPrefix            string
	metadataDisplayDenom     string
	denom                    string
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
		metadataDisplayDenom:     c.MetadataDisplayDenom,
		denom:                    c.Denom,
		nodeConfig:               c.NodeConfig.Clone(),
		fee:                      c.Fee,
		gov:                      c.GovConfig,
		staking:                  c.StakingConfig,
		mu:                       &sync.Mutex{},
		fundedAccounts:           append([]FundedAccount{}, c.FundedAccounts...),
		genTxs:                   append([]json.RawMessage{}, c.GenTxs...),
		enableFakeUpgradeHandler: c.IsFakeUpgradeHandlerEnabled,
	}

	return n
}

// FundedAccount is used to provide information about prefunded
// accounts in network config
type FundedAccount struct {
	// we can't use the sdk.AccAddress because of configurable prefixes
	Address  string
	Balances sdk.Coins
}

// FundAccount funds address with balances at genesis
func (n *Network) FundAccount(accAddress sdk.AccAddress, balances sdk.Coins) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.fundedAccounts = append(n.fundedAccounts, FundedAccount{
		Address:  accAddress.String(),
		Balances: balances,
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
) authcosmostypes.GenesisAccounts {
	accountAddress := sdk.MustAccAddressFromBech32(fa.Address)
	accountState = append(accountState, authcosmostypes.NewBaseAccount(accountAddress, nil, 0, 0))
	coins := fa.Balances
	bankState.Balances = append(
		bankState.Balances,
		banktypes.Balance{Address: accountAddress.String(), Coins: coins},
	)
	bankState.Supply = bankState.Supply.Add(coins...)

	return accountState
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
		accountState = applyFundedAccountToGenesis(fundedAcc, accountState, bankState)
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

// Denom returns the base chain denom. This is different
// for each network(i.e. mainnet, testnet, etc)
func (n Network) Denom() string {
	return n.denom
}

// FeeModel returns fee model configuration
func (n Network) FeeModel() feemodeltypes.Model {
	return n.fee.FeeModel
}

// IsFakeUpgradeHandlerEnabled enables temporary fake upgrade handler until we have real one
func (n Network) IsFakeUpgradeHandlerEnabled() bool {
	return n.enableFakeUpgradeHandler
}

// DeterministicGas returns deterministic gas amounts required by some message types
func (n Network) DeterministicGas() DeterministicGasRequirements {
	return n.fee.DeterministicGas
}

// NetworkConfigByChainID returns predefined NetworkConfig for a ChainID.
func NetworkConfigByChainID(id ChainID) (NetworkConfig, error) {
	nc, found := networkConfigs[id]
	if !found {
		return NetworkConfig{}, errors.Errorf("chainID %s not found", id)
	}

	return nc, nil
}

// NetworkByChainID returns predefined Network for a ChainID.
func NetworkByChainID(id ChainID) (Network, error) {
	nc, err := NetworkConfigByChainID(id)
	if err != nil {
		return Network{}, err
	}

	// TODO: remove this check once all preconfigured networkConfigs are enabled
	if !nc.Enabled {
		return Network{}, errors.Errorf("%s is not yet ready, use --chain-id=%s for devnet", id, string(ChainIDDev))
	}

	return NewNetwork(nc), nil
}

//go:embed genesis/genesis.tmpl.json
var genesisTemplate string

func genesis(n Network) ([]byte, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Funcs(funcMap).Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC       string
		ChainID              ChainID
		MetadataDisplayDenom string
		Denom                string
		FeeModelParams       feemodeltypes.ModelParams
		Gov                  GovConfig
		Staking              StakingConfig
	}{
		GenesisTimeUTC:       n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:              n.chainID,
		MetadataDisplayDenom: n.metadataDisplayDenom,
		Denom:                n.denom,
		FeeModelParams:       n.FeeModel().Params(),
		Gov:                  n.gov,
		Staking:              n.staking,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}

	return genesisBuf.Bytes(), nil
}
