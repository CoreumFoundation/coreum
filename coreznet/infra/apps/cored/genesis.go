package cored

import (
	"crypto/ed25519"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// NewGenesis creates new configuration for genesis block
func NewGenesis(chainID string) *Genesis {
	interfaceRegistry := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &stakingtypes.MsgCreateValidator{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesis(chainID))
	must.OK(err)
	var appState map[string]json.RawMessage
	must.OK(json.Unmarshal(genesisDoc.AppState, &appState))

	authState := authtypes.GetGenesisStateFromAppState(marshaler, appState)

	accountState, err := authtypes.UnpackAccounts(authState.Accounts)
	must.OK(err)

	return &Genesis{
		chainID:      chainID,
		marshaler:    marshaler,
		txConfig:     tx.NewTxConfig(marshaler, []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_DIRECT}),
		mu:           &sync.Mutex{},
		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(marshaler, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(marshaler, appState),
	}
}

// Genesis is responsible for creating genesis configuration for coreum network
type Genesis struct {
	chainID   string
	marshaler *codec.ProtoCodec
	txConfig  client.TxConfig

	mu           *sync.Mutex
	genesisDoc   *tmtypes.GenesisDoc
	appState     map[string]json.RawMessage
	genutilState *genutiltypes.GenesisState
	authState    authtypes.GenesisState
	accountState authtypes.GenesisAccounts
	bankState    *banktypes.GenesisState
	finalized    bool
}

// ChainID returns ID of chain
func (g Genesis) ChainID() string {
	return g.chainID
}

// AddWallet adds wallet with balances to the genesis
func (g *Genesis) AddWallet(publicKey Secp256k1PublicKey, balances string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.verifyNotFinalized()

	pubKey := cosmossecp256k1.PubKey{Key: publicKey}

	accountAddress := sdk.AccAddress(pubKey.Address())
	g.accountState = append(g.accountState, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))

	coins, err := sdk.ParseCoinsNormalized(balances)
	must.OK(err)

	g.bankState.Balances = append(g.bankState.Balances, banktypes.Balance{Address: accountAddress.String(), Coins: coins})
	g.bankState.Supply = g.bankState.Supply.Add(coins...)
}

// AddValidator adds transaction configuring validator to genesis block
func (g *Genesis) AddValidator(validatorPublicKey ed25519.PublicKey, stakerPrivateKey Secp256k1PrivateKey) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.verifyNotFinalized()

	amount, err := sdk.ParseCoinNormalized("100000000stake")
	must.OK(err)

	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	valPubKey := &cosmosed25519.PubKey{Key: validatorPublicKey}
	stakerPrivKey := &cosmossecp256k1.PrivKey{Key: stakerPrivateKey}
	stakerAddress := sdk.AccAddress(stakerPrivKey.PubKey().Address())

	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(stakerAddress), valPubKey, amount, stakingtypes.Description{Moniker: stakerAddress.String()}, commission, sdk.OneInt())
	must.OK(err)

	txBuilder := g.txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	must.OK(txBuilder.SetMsgs(msg))

	signerData := authsigning.SignerData{
		ChainID:       g.chainID,
		AccountNumber: 0,
		Sequence:      0,
	}
	sigData := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   stakerPrivKey.PubKey(),
		Data:     &sigData,
		Sequence: 0,
	}
	must.OK(txBuilder.SetSignatures(sig))

	bytesToSign := must.Bytes(g.txConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx()))
	sigBytes, err := stakerPrivKey.Sign(bytesToSign)
	must.OK(err)

	sigData = signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   stakerPrivKey.PubKey(),
		Data:     &sigData,
		Sequence: 0,
	}
	must.OK(txBuilder.SetSignatures(sig))

	g.genutilState.GenTxs = append(g.genutilState.GenTxs, must.Bytes(g.txConfig.TxJSONEncoder()(txBuilder.GetTx())))
}

// Save saves genesis configuration
func (g *Genesis) Save(homeDir string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.finalized = true

	genutiltypes.SetGenesisStateInAppState(g.marshaler, g.appState, g.genutilState)

	var err error
	g.authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(g.accountState))
	must.OK(err)
	g.appState[authtypes.ModuleName] = g.marshaler.MustMarshalJSON(&g.authState)

	g.bankState.Balances = banktypes.SanitizeGenesisBalances(g.bankState.Balances)
	g.appState[banktypes.ModuleName] = g.marshaler.MustMarshalJSON(g.bankState)

	g.genesisDoc.AppState = must.Bytes(json.MarshalIndent(g.appState, "", "  "))

	must.OK(os.MkdirAll(homeDir+"/config", 0o700))
	must.OK(g.genesisDoc.SaveAs(homeDir + "/config/genesis.json"))
}

func (g *Genesis) verifyNotFinalized() {
	if g.finalized {
		panic("genesis has been already saved, no more operations are allowed")
	}
}

type a = []interface{}
type o = map[string]interface{}

func genesis(chainID string) []byte {
	return must.Bytes(json.Marshal(o{
		"genesis_time":   time.Now().UTC(),
		"chain_id":       chainID,
		"initial_height": "1",
		"consensus_params": o{
			"block": o{
				"max_bytes":    "22020096",
				"max_gas":      "-1",
				"time_iota_ms": "1000",
			},
			"evidence": o{
				"max_age_num_blocks": "100000",
				"max_age_duration":   "172800000000000",
				"max_bytes":          "1048576",
			},
			"validator": o{
				"pub_key_types": a{"ed25519"},
			},
		},
		"app_state": o{
			"auth": o{
				"params": o{
					"max_memo_characters":       "256",
					"tx_sig_limit":              "7",
					"tx_size_cost_per_byte":     "10",
					"sig_verify_cost_ed25519":   "590",
					"sig_verify_cost_secp256k1": "1000",
				},
			},
			"bank": o{
				"params": o{
					"default_send_enabled": true,
				},
			},
			"capability": o{
				"index": "1",
			},
			"crisis": o{
				"constant_fee": o{
					"denom":  "stake",
					"amount": "1000",
				},
			},
			"distribution": o{
				"params": o{
					"community_tax":         "0.020000000000000000",
					"base_proposer_reward":  "0.010000000000000000",
					"bonus_proposer_reward": "0.040000000000000000",
					"withdraw_addr_enabled": true,
				},
			},
			"gov": o{
				"starting_proposal_id": "1",
				"deposit_params": o{
					"min_deposit": a{
						o{
							"denom":  "stake",
							"amount": "10000000",
						},
					},
					"max_deposit_period": "172800s",
				},
				"voting_params": o{
					"voting_period": "172800s",
				},
				"tally_params": o{
					"quorum":         "0.334000000000000000",
					"threshold":      "0.500000000000000000",
					"veto_threshold": "0.334000000000000000",
				},
			},
			"ibc": o{
				"client_genesis": o{
					"params": o{
						"allowed_clients": a{
							"06-solomachine",
							"07-tendermint",
						},
					},
					"create_localhost":     false,
					"next_client_sequence": "0",
				},
				"connection_genesis": o{
					"next_connection_sequence": "0",
					"params": o{
						"max_expected_time_per_block": "30000000000",
					},
				},
				"channel_genesis": o{
					"next_channel_sequence": "0",
				},
			},
			"mint": o{
				"minter": o{
					"inflation":         "0.130000000000000000",
					"annual_provisions": "0.000000000000000000",
				},
				"params": o{
					"mint_denom":            "stake",
					"inflation_rate_change": "0.130000000000000000",
					"inflation_max":         "0.200000000000000000",
					"inflation_min":         "0.070000000000000000",
					"goal_bonded":           "0.670000000000000000",
					"blocks_per_year":       "6311520",
				},
			},
			"monitoringp": o{
				"port_id": "monitoring",
				"params": o{
					"lastBlockHeight":         "1",
					"consumerChainID":         "spn-1",
					"consumerUnbondingPeriod": "1814400",
					"consumerRevisionHeight":  "1",
				},
			},
			"slashing": o{
				"params": o{
					"signed_blocks_window":       "100",
					"min_signed_per_window":      "0.500000000000000000",
					"downtime_jail_duration":     "600s",
					"slash_fraction_double_sign": "0.050000000000000000",
					"slash_fraction_downtime":    "0.010000000000000000",
				},
			},
			"staking": o{
				"params": o{
					"unbonding_time":     "1814400s",
					"max_validators":     100,
					"max_entries":        7,
					"historical_entries": 10000,
					"bond_denom":         "stake",
				},
				"last_total_power": "0",
			},
			"transfer": o{
				"port_id": "transfer",
				"params": o{
					"send_enabled":    true,
					"receive_enabled": true,
				},
			},
		},
	}))
}
