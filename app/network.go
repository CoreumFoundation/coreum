package app

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io/ioutil"
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

func init() {
	list := []NetworkConfig{
		{
			ChainID:       Mainnet,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "core",
			TokenSymbol:   TokenSymbolMain,
		},
		{
			ChainID:       Devnet,
			GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
			AddressPrefix: "devcore",
			TokenSymbol:   TokenSymbolDev,
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
				// validator 0
				json.RawMessage(`{
				  "body": {
					"messages": [
					  {
						"@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
						"description": {
						  "moniker": "devcore10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
						  "identity": "",
						  "website": "",
						  "security_contact": "",
						  "details": ""
						},
						"commission": {
						  "rate": "0.100000000000000000",
						  "max_rate": "0.200000000000000000",
						  "max_change_rate": "0.010000000000000000"
						},
						"min_self_delegation": "1",
						"delegator_address": "devcore10krrrqxxy948n5p9xvwgq6krgy9hg5g8svaz62",
						"validator_address": "devcorevaloper10krrrqxxy948n5p9xvwgq6krgy9hg5g8fnr84l",
						"pubkey": {
						  "@type": "/cosmos.crypto.ed25519.PubKey",
						  "key": "lMMi0GqO68wCsWUQc8GwNBazv7z6lpQSMUJW+qVLGdk="
						},
						"value": {
						  "denom": "dacore",
						  "amount": "100000000"
						}
					  }
					],
					"memo": "",
					"timeout_height": "0",
					"extension_options": [],
					"non_critical_extension_options": []
				  },
				  "auth_info": {
					"signer_infos": [
					  {
						"public_key": {
						  "@type": "/cosmos.crypto.secp256k1.PubKey",
						  "key": "Ak03QePuJN+xe9v/zcNR3J/RQxlTTXwzNV35CEJYy0VZ"
						},
						"mode_info": {
						  "single": {
							"mode": "SIGN_MODE_DIRECT"
						  }
						},
						"sequence": "0"
					  }
					],
					"fee": {
					  "amount": [],
					  "gas_limit": "200000",
					  "payer": "",
					  "granter": ""
					}
				  },
				  "signatures": [
					"P3/EMiLx2mwX1HqVFFFIrl60nvq/V5bVlrUBqLMOLq11uKTrNM/D+NJYGsbXSvekbS2OadVA0o1zs2OUwg5fGA=="
				  ]
				}`),

				// validator 1
				json.RawMessage(`{
				  "body": {
					"messages": [
					  {
						"@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
						"description": {
						  "moniker": "devcore1fvnwq8605fgex6qyr96enlhpgsnzwge62hs7er",
						  "identity": "",
						  "website": "",
						  "security_contact": "",
						  "details": ""
						},
						"commission": {
						  "rate": "0.100000000000000000",
						  "max_rate": "0.200000000000000000",
						  "max_change_rate": "0.010000000000000000"
						},
						"min_self_delegation": "1",
						"delegator_address": "devcore1fvnwq8605fgex6qyr96enlhpgsnzwge62hs7er",
						"validator_address": "devcorevaloper1fvnwq8605fgex6qyr96enlhpgsnzwge6ngwmkk",
						"pubkey": {
						  "@type": "/cosmos.crypto.ed25519.PubKey",
						  "key": "z9YAf63ZM+kDOEeGPNv2X2LcsyRX3V+++qLiP1xjQbw="
						},
						"value": {
						  "denom": "dacore",
						  "amount": "100000000"
						}
					  }
					],
					"memo": "",
					"timeout_height": "0",
					"extension_options": [],
					"non_critical_extension_options": []
				  },
				  "auth_info": {
					"signer_infos": [
					  {
						"public_key": {
						  "@type": "/cosmos.crypto.secp256k1.PubKey",
						  "key": "AsF3rd/0kROfqMBQ0LpDzqUQOg3Q527YdpR/wBUOtJNW"
						},
						"mode_info": {
						  "single": {
							"mode": "SIGN_MODE_DIRECT"
						  }
						},
						"sequence": "0"
					  }
					],
					"fee": {
					  "amount": [],
					  "gas_limit": "200000",
					  "payer": "",
					  "granter": ""
					}
				  },
				  "signatures": [
					"tBjqZTA4m1yvrdszHZVK69k+xYtQbbhxNEjZsagzOFlo72ZpdYuUjEmeq4Vi3MuvGrr4igLxwf0x/kL1iaPxUQ=="
				  ]
				}`),

				// validator 2
				json.RawMessage(`{
				  "body": {
					"messages": [
					  {
						"@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
						"description": {
						  "moniker": "devcore15pwm9e2jkp5e0knudj0np6mf4dd30g0uych2ug",
						  "identity": "",
						  "website": "",
						  "security_contact": "",
						  "details": ""
						},
						"commission": {
						  "rate": "0.100000000000000000",
						  "max_rate": "0.200000000000000000",
						  "max_change_rate": "0.010000000000000000"
						},
						"min_self_delegation": "1",
						"delegator_address": "devcore15pwm9e2jkp5e0knudj0np6mf4dd30g0uych2ug",
						"validator_address": "devcorevaloper15pwm9e2jkp5e0knudj0np6mf4dd30g0ua8f0na",
						"pubkey": {
						  "@type": "/cosmos.crypto.ed25519.PubKey",
						  "key": "4sT/QxjkUPfNx6cNlhSelryF3AEzJ3unn7rRewFNyx0="
						},
						"value": {
						  "denom": "dacore",
						  "amount": "100000000"
						}
					  }
					],
					"memo": "",
					"timeout_height": "0",
					"extension_options": [],
					"non_critical_extension_options": []
				  },
				  "auth_info": {
					"signer_infos": [
					  {
						"public_key": {
						  "@type": "/cosmos.crypto.secp256k1.PubKey",
						  "key": "Az4hdim6VdSt/jX0zMuWjn9QU8AXFRarfBHYRHre3w1S"
						},
						"mode_info": {
						  "single": {
							"mode": "SIGN_MODE_DIRECT"
						  }
						},
						"sequence": "0"
					  }
					],
					"fee": {
					  "amount": [],
					  "gas_limit": "200000",
					  "payer": "",
					  "granter": ""
					}
				  },
				  "signatures": [
					"zmPq2tjbtKMcNfEpmbgM7Agxykwg1pkSquq4325nm25O2tHCgwkmnomWBnZ7Vc5ODUUc+UOAZkFKGlJPh2iiAQ=="
				  ]
				}`),

				// validator 3
				json.RawMessage(`{
				  "body": {
					"messages": [
					  {
						"@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
						"description": {
						  "moniker": "devcore1dawh59rtggwpxe4pevtaklu0l3z322reg86qyd",
						  "identity": "",
						  "website": "",
						  "security_contact": "",
						  "details": ""
						},
						"commission": {
						  "rate": "0.100000000000000000",
						  "max_rate": "0.200000000000000000",
						  "max_change_rate": "0.010000000000000000"
						},
						"min_self_delegation": "1",
						"delegator_address": "devcore1dawh59rtggwpxe4pevtaklu0l3z322reg86qyd",
						"validator_address": "devcorevaloper1dawh59rtggwpxe4pevtaklu0l3z322re3cy9tc",
						"pubkey": {
						  "@type": "/cosmos.crypto.ed25519.PubKey",
						  "key": "hVFxUbrNW/KoiqQFQYYncQWLx4o884hdeV/mOOakaqQ="
						},
						"value": {
						  "denom": "dacore",
						  "amount": "100000000"
						}
					  }
					],
					"memo": "",
					"timeout_height": "0",
					"extension_options": [],
					"non_critical_extension_options": []
				  },
				  "auth_info": {
					"signer_infos": [
					  {
						"public_key": {
						  "@type": "/cosmos.crypto.secp256k1.PubKey",
						  "key": "Arv7raYNjP2faiTEuNAshLRXS9jBi9r5GCFj3nv9XfXn"
						},
						"mode_info": {
						  "single": {
							"mode": "SIGN_MODE_DIRECT"
						  }
						},
						"sequence": "0"
					  }
					],
					"fee": {
					  "amount": [],
					  "gas_limit": "200000",
					  "payer": "",
					  "granter": ""
					}
				  },
				  "signatures": [
					"MxT5CHJiurOyYkTa5jOw+EbfJiruNPJe4zB70G1sZaBKrmdX9t66ZElMol853oz4uIQJ/hFrBpGe/W0C0rbfAQ=="
				  ]
				}`),
			},
		},
	}

	for _, elem := range list {
		networks[elem.ChainID] = elem
	}
}

var networks = map[ChainID]NetworkConfig{}

// NetworkConfig helps initialize Network instance
type NetworkConfig struct {
	ChainID        ChainID
	GenesisTime    time.Time
	AddressPrefix  string
	TokenSymbol    string
	FundedAccounts []FundedAccount
	GenTxs         []json.RawMessage
}

// Network holds all the configuration for different predefined networks
type Network struct {
	chainID       ChainID
	genesisTime   time.Time
	addressPrefix string
	tokenSymbol   string

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
