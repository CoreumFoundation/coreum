package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v2/genesis"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
)

var (
	// GenesisV1Template is the genesis template used by v1 version of the chain.
	//go:embed genesis/genesis.v1.tmpl.json
	GenesisV1Template string

	// GenesisV2Template is the genesis template used by v2 version of the chain.
	//go:embed genesis/genesis.v2.tmpl.json
	GenesisV2Template string

	//go:embed genesis/gentx/coreum-devnet-1
	devGenTxsFS embed.FS

	networkConfigs map[constant.ChainID]NetworkConfig
)

func init() {
	// 10m delegated and 1m extra to the txs
	devStakerValidatorBalance := sdk.NewCoins(sdk.NewCoin(constant.DenomDev, sdk.NewInt(11_000_000_000_000)))

	// configs
	networkConfigs = map[constant.ChainID]NetworkConfig{
		constant.ChainIDMain: {
			Provider: NewStaticConfigProvider(genesis.MainnetGenesis),
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"0df493af80fbaad41b9b26d6f4520b39ceb1d210@34.171.208.193:26656", // seed-iron
					"cba16f4f32707d70a2a2d10861fac897f1e9aaa1@34.72.150.107:26656",  // seed-nickle
				},
			},
		},
		constant.ChainIDTest: {
			Provider: NewStaticConfigProvider(genesis.TestnetGenesis),
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"64391878009b8804d90fda13805e45041f492155@35.232.157.206:26656", // seed-sirius
					"53f2367d8f8291af8e3b6ca60efded0675ff6314@34.29.15.170:26656",   // seed-antares
				},
			},
		},
		constant.ChainIDDev: {
			Provider: DynamicConfigProvider{
				GenesisTemplate: GenesisV2Template,
				ChainID:         constant.ChainIDDev,
				GenesisTime:     time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
				BlockTimeIota:   time.Second,
				Denom:           constant.DenomDev,
				AddressPrefix:   constant.AddressPrefixDev,
				GovConfig: GovConfig{
					ProposalConfig: GovProposalConfig{
						MinDepositAmount: "4000000000", // 4,000 CORE
						VotingPeriod:     "4h",         // 4 hours
					},
				},
				CustomParamsConfig: CustomParamsConfig{
					Staking: CustomParamsStakingConfig{
						MinSelfDelegation: sdk.NewInt(20_000_000_000), // 20k core
					},
				},
				FundedAccounts: []FundedAccount{
					// Staker of validator Mercury
					{
						Address:  "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6",
						Balances: devStakerValidatorBalance,
					},
					// Staker of validator Venus
					{
						Address:  "devcore105ct3vl89ar53jrj23zl6e09cmqwym2ua5hegf",
						Balances: devStakerValidatorBalance,
					},
					// Staker of validator Earth
					{
						Address:  "devcore14x46r5eflga696sd5my900euvlplu2prhny5ae",
						Balances: devStakerValidatorBalance,
					},
					// Faucet's account storing the rest of total supply
					{
						Address:  "devcore1ckuncyw0hftdq5qfjs6ee2v6z73sq0urd390cd",
						Balances: sdk.NewCoins(sdk.NewCoin(constant.DenomDev, sdk.NewInt(100_000_000_000_000))), // 100m faucet
					},
				},
				GenTxs: readGenTxs(devGenTxsFS),
			},
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"602df7489bd45626af5c9a4ea7f700ceb2222b19@34.135.242.117:26656",
					"88d1266e086bfe33589886cc10d4c58e85a69d14@34.135.191.69:26656",
				},
			},
		},
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

// GovConfig contains gov module configs.
type GovConfig struct {
	ProposalConfig GovProposalConfig
}

// GovProposalConfig contains gov module proposal-related configuration options.
type GovProposalConfig struct {
	// MinDepositAmount is the minimum amount needed to create a proposal. Basically anti-spam policy.
	MinDepositAmount string

	// VotingPeriod is the proposal voting period duration.
	VotingPeriod string
}

// CustomParamsStakingConfig contains custom params for the staking module configuration.
type CustomParamsStakingConfig struct {
	// MinSelfDelegation is the minimum allowed amount of the stake coin for the validator to be created.
	MinSelfDelegation sdk.Int
}

// CustomParamsConfig contains custom params module configuration.
type CustomParamsConfig struct {
	Staking CustomParamsStakingConfig
}

// FundedAccount is used to provide information about prefunded
// accounts in network config.
type FundedAccount struct {
	// we can't use the sdk.AccAddress because of configurable prefixes
	Address  string
	Balances sdk.Coins
}

// NetworkConfig helps initialize Network instance.
type NetworkConfig struct {
	Provider   NetworkConfigProvider
	NodeConfig NodeConfig
}

// SetSDKConfig sets global SDK config to some network-specific values.
// In typical applications this func should be called right after network initialization.
func (c NetworkConfig) SetSDKConfig() {
	SetSDKConfig(c.Provider.GetAddressPrefix(), constant.CoinType)
}

// SetSDKConfig sets global SDK config.
func SetSDKConfig(addressPrefix string, coinType uint32) {
	config := sdk.GetConfig()

	// Set address & public key prefixes
	config.SetBech32PrefixForAccount(addressPrefix, addressPrefix+"pub")
	config.SetBech32PrefixForValidator(addressPrefix+"valoper", addressPrefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(addressPrefix+"valcons", addressPrefix+"valconspub")

	// Set BIP44 coin type corresponding to CORE
	config.SetCoinType(coinType)

	config.Seal()
}

// Denom returns denom.
func (c NetworkConfig) Denom() string {
	return c.Provider.GetDenom()
}

// ChainID returns chain ID.
func (c NetworkConfig) ChainID() constant.ChainID {
	return c.Provider.GetChainID()
}

// EncodeGenesis returns the json encoded representation of the genesis file.
func (c NetworkConfig) EncodeGenesis() ([]byte, error) {
	return c.Provider.EncodeGenesis()
}

// NetworkConfigByChainID returns predefined NetworkConfig for a ChainID.
func NetworkConfigByChainID(id constant.ChainID) (NetworkConfig, error) {
	nc, found := networkConfigs[id]
	if !found {
		return NetworkConfig{}, errors.Errorf("chainID %s not found", id)
	}

	nc.NodeConfig = nc.NodeConfig.clone()

	return nc, nil
}
