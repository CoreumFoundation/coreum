package cored

import (
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"os"
	"sync"
	"text/template"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const initialBalance = "1000000000000000core"

// NewGenesis creates new configuration for genesis block
func NewGenesis(chainID string) *Genesis {
	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesis(chainID))
	must.OK(err)
	var appState map[string]json.RawMessage
	must.OK(json.Unmarshal(genesisDoc.AppState, &appState))

	clientCtx := NewContext(chainID, nil)
	authState := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
	accountState, err := authtypes.UnpackAccounts(authState.Accounts)
	must.OK(err)
	g := &Genesis{
		clientCtx:    clientCtx,
		mu:           &sync.Mutex{},
		genesisDoc:   genesisDoc,
		appState:     appState,
		genutilState: genutiltypes.GetGenesisStateFromAppState(clientCtx.Codec, appState),
		authState:    authState,
		accountState: accountState,
		bankState:    banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState),
	}
	g.AddWallet(AlicePrivKey.PubKey(), initialBalance)
	g.AddWallet(BobPrivKey.PubKey(), initialBalance)
	g.AddWallet(CharliePrivKey.PubKey(), initialBalance)

	for _, key := range RandomWallets {
		g.AddWallet(key.PubKey(), initialBalance)
	}

	return g
}

// Genesis is responsible for creating genesis configuration for coreum network
type Genesis struct {
	clientCtx client.Context

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
	return g.clientCtx.ChainID
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
func (g *Genesis) AddValidator(validatorPublicKey ed25519.PublicKey, stakerPrivateKey Secp256k1PrivateKey, stakedBalance string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.verifyNotFinalized()

	amount, err := sdk.ParseCoinNormalized(stakedBalance)
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

	g.genutilState.GenTxs = append(g.genutilState.GenTxs, must.Bytes(g.clientCtx.TxConfig.TxJSONEncoder()(signTx(g.clientCtx, stakerPrivateKey, 0, 0, msg))))
}

// Save saves genesis configuration
func (g *Genesis) Save(homeDir string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.finalized = true

	genutiltypes.SetGenesisStateInAppState(g.clientCtx.Codec, g.appState, g.genutilState)

	var err error
	g.authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(g.accountState))
	must.OK(err)
	g.appState[authtypes.ModuleName] = g.clientCtx.Codec.MustMarshalJSON(&g.authState)

	g.bankState.Balances = banktypes.SanitizeGenesisBalances(g.bankState.Balances)
	g.appState[banktypes.ModuleName] = g.clientCtx.Codec.MustMarshalJSON(g.bankState)

	g.genesisDoc.AppState = must.Bytes(json.MarshalIndent(g.appState, "", "  "))

	must.OK(os.MkdirAll(homeDir+"/config", 0o700))
	must.OK(g.genesisDoc.SaveAs(homeDir + "/config/genesis.json"))
}

func (g *Genesis) verifyNotFinalized() {
	if g.finalized {
		panic("genesis has been already saved, no more operations are allowed")
	}
}

// MetadataTemplate contains hasura metadata template
//go:embed genesis/genesis.tmpl.json
var genesisTemplate string

func genesis(chainID string) []byte {
	genesisBuf := new(bytes.Buffer)

	must.OK(template.Must(template.New("genesis").Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC string
		ChainID        string
	}{
		GenesisTimeUTC: time.Now().UTC().Format(time.RFC3339),
		ChainID:        chainID,
	}))

	return genesisBuf.Bytes()
}
