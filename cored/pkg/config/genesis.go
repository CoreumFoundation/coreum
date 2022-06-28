package config

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"os"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum/cored/pkg/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

const initialBalance = "1000000000000000core"

// Genesis is responsible for creating genesis configuration for coreum network
type Genesis struct {
	codec      *codec.ProtoCodec
	genesisDoc *tmtypes.GenesisDoc

	mu           *sync.Mutex
	appState     map[string]json.RawMessage
	finalized    bool
	accountState authtypes.GenesisAccounts
	bankState    *banktypes.GenesisState
	genutilState *genutiltypes.GenesisState
	authState    authtypes.GenesisState
}

// ChainID returns ID of chain
func (g Genesis) ChainID() string {
	return g.genesisDoc.ChainID
}

// FundAccount funds address with balances at genesis
func (g *Genesis) FundAccount(publicKey types.Secp256k1PublicKey, balances string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.verifyNotFinalized()

	pubKey := cosmossecp256k1.PubKey{Key: publicKey}

	accountAddress := sdk.AccAddress(pubKey.Address())
	g.accountState = append(g.accountState, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))

	coins, err := sdk.ParseCoinsNormalized(balances)
	if err != nil {
		return err
	}

	g.bankState.Balances = append(g.bankState.Balances, banktypes.Balance{Address: accountAddress.String(), Coins: coins})
	g.bankState.Supply = g.bankState.Supply.Add(coins...)
	return nil
}

// AddGenesisTx adds transaction to the genesis file
func (g *Genesis) AddGenesisTx(signedTx json.RawMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.verifyNotFinalized()
	g.genutilState.GenTxs = append(g.genutilState.GenTxs, signedTx)
}

// Save saves genesis configuration
func (g *Genesis) Save(homeDir string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.finalized = true

	genutiltypes.SetGenesisStateInAppState(g.codec, g.appState, g.genutilState)

	var err error
	g.authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(g.accountState))
	if err != nil {
		return err
	}
	g.appState[authtypes.ModuleName] = g.codec.MustMarshalJSON(&g.authState)

	g.bankState.Balances = banktypes.SanitizeGenesisBalances(g.bankState.Balances)
	g.appState[banktypes.ModuleName] = g.codec.MustMarshalJSON(g.bankState)

	g.genesisDoc.AppState, err = json.MarshalIndent(g.appState, "", "  ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(homeDir+"/config", 0o700)
	if err != nil {
		return err
	}

	err = g.genesisDoc.SaveAs(homeDir + "/config/genesis.json")
	if err != nil {
		return err
	}

	return nil
}

func (g *Genesis) verifyNotFinalized() {
	if g.finalized {
		panic("genesis has been already saved, no more operations are allowed")
	}
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
		GenesisTimeUTC: n.GenesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.ChainID,
		TokenSymbol:    n.TokenSymbol,
	})

	return genesisBuf.Bytes(), err
}
