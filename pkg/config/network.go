package config

import (
	_ "embed"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v4/genesis"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
)

var (
	// GenesisV1Template is the genesis template used by v1 version of the chain.
	//go:embed genesis/genesis.v1.tmpl.json
	GenesisV1Template string

	// GenesisV2Template is the genesis template used by v2 version of the chain.
	//go:embed genesis/genesis.v2.tmpl.json
	GenesisV2Template string

	// GenesisV3Template is the genesis template used by v3 version of the chain.
	//go:embed genesis/genesis.v3.tmpl.json
	GenesisV3Template string

	networkConfigs map[constant.ChainID]NetworkConfig
)

func init() {
	// configs
	networkConfigs = map[constant.ChainID]NetworkConfig{
		constant.ChainIDMain: {
			Provider: NewStaticConfigProvider(genesis.MainnetGenesis),
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"0df493af80fbaad41b9b26d6f4520b39ceb1d210@seed-iron.mainnet-1.coreum.dev:26656",   // seed-iron
					"cba16f4f32707d70a2a2d10861fac897f1e9aaa1@seed-nickle.mainnet-1.coreum.dev:26656", // seed-nickle
				},
			},
		},
		constant.ChainIDTest: {
			Provider: NewStaticConfigProvider(genesis.TestnetGenesis),
			NodeConfig: NodeConfig{
				SeedPeers: []string{
					"64391878009b8804d90fda13805e45041f492155@seed-sirius.testnet-1.coreum.dev:26656",  // seed-sirius
					"53f2367d8f8291af8e3b6ca60efded0675ff6314@seed-antares.testnet-1.coreum.dev:26656", // seed-antares
				},
			},
		},
		constant.ChainIDDev: {
			Provider: DynamicConfigProvider{
				GenesisTemplate: GenesisV3Template,
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
						MinSelfDelegation: sdkmath.NewInt(20_000_000_000), // 20k core
					},
				},
				FundedAccounts: []FundedAccount{
					// Faucet's account
					{
						Address:  "devcore1ckuncyw0hftdq5qfjs6ee2v6z73sq0urd390cd",
						Balances: sdk.NewCoins(sdk.NewCoin(constant.DenomDev, sdkmath.NewInt(100_000_000_000_000))), // 100m faucet
					},
				},
			},
			NodeConfig: NodeConfig{},
		},
	}
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
	MinSelfDelegation sdkmath.Int
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

// ValPrefixFromAddressPrefix returns validator operator prefix.
func ValPrefixFromAddressPrefix(addressPrefix string) string {
	return addressPrefix + "valoper"
}

// ConsPrefixFromAddressPrefix returns consensus prefix.
func ConsPrefixFromAddressPrefix(addressPrefix string) string {
	return addressPrefix + "valcons"
}
