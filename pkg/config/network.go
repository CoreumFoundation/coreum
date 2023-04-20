package config

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/genesis"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

var (
	//go:embed genesis/genesis.tmpl.json
	genesisTemplate string

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
			Provider:      NewJSONConfigProvider(genesis.MainnetGenesis),
			AddressPrefix: constant.AddressPrefixMain,
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"0df493af80fbaad41b9b26d6f4520b39ceb1d210@34.171.208.193:26656", // seed-iron
					"cba16f4f32707d70a2a2d10861fac897f1e9aaa1@34.72.150.107:26656",  // seed-nickle
				},
			},
		},
		constant.ChainIDTest: {
			Provider:      NewJSONConfigProvider(genesis.TestnetGenesis),
			AddressPrefix: constant.AddressPrefixTest,
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"64391878009b8804d90fda13805e45041f492155@35.232.157.206:26656", // seed-sirius
					"53f2367d8f8291af8e3b6ca60efded0675ff6314@34.29.15.170:26656",   // seed-antares
				},
			},
		},
		constant.ChainIDDev: {
			Provider: DirectConfigProvider{
				ChainID:     constant.ChainIDDev,
				GenesisTime: time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
				Denom:       constant.DenomDev,
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
			AddressPrefix: constant.AddressPrefixDev,
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
	Provider NetworkConfigProvider

	AddressPrefix string
	NodeConfig    NodeConfig
}

// SetSDKConfig sets global SDK config to some network-specific values.
// In typical applications this func should be called right after network initialization.
func (c NetworkConfig) SetSDKConfig() {
	config := sdk.GetConfig()

	// Set address & public key prefixes
	config.SetBech32PrefixForAccount(c.AddressPrefix, c.AddressPrefix+"pub")
	config.SetBech32PrefixForValidator(c.AddressPrefix+"valoper", c.AddressPrefix+"valoperpub")
	config.SetBech32PrefixForConsensusNode(c.AddressPrefix+"valcons", c.AddressPrefix+"valconspub")

	// Set BIP44 coin type corresponding to CORE
	config.SetCoinType(constant.CoinType)

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

// GenesisDoc returns the genesis doc of the network.
func (c NetworkConfig) GenesisDoc() (*tmtypes.GenesisDoc, error) {
	return c.Provider.GenesisDoc()
}

// EncodeGenesis returns the json encoded representation of the genesis file.
func (c NetworkConfig) EncodeGenesis() ([]byte, error) {
	genesisDoc, err := c.GenesisDoc()
	if err != nil {
		return nil, errors.Wrap(err, "not able to get genesis doc")
	}

	bs, err := tmjson.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal genesis doc")
	}

	return append(bs, '\n'), nil
}

// SaveGenesis saves json encoded representation of the genesis config into file.
func (c NetworkConfig) SaveGenesis(homeDir string) error {
	genDocBytes, err := c.EncodeGenesis()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(homeDir+"/config", 0o700); err != nil {
		return errors.Wrap(err, "unable to make config directory")
	}

	err = os.WriteFile(homeDir+"/config/genesis.json", genDocBytes, 0644)
	return errors.Wrap(err, "unable to write genesis bytes to file")
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
