package config

import (
	"crypto/ed25519"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func init() {
	n, err := NetworkByChainID(Devnet)
	if err != nil {
		panic(err)
	}

	n.SetupPrefixes()
}

func TestAddressPrefixIsSet(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)
	n, err := NetworkByChainID(Devnet)
	requireT.NoError(err)
	pubKey, _ := types.GenerateSecp256k1Key()
	secp256k1 := cosmossecp256k1.PubKey{Key: pubKey}
	accountAddress := sdk.AccAddress(secp256k1.Address())
	assertT.True(strings.HasPrefix(accountAddress.String(), n.addressPrefix))
}

func TestGenesisValidation(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := NetworkByChainID(Devnet)
	requireT.NoError(err)

	gen, err := n.Genesis()
	requireT.NoError(err)

	encCfg := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	requireT.NoError(app.ModuleBasics.ValidateGenesis(encCfg.Marshaler, encCfg.TxConfig, gen.appState))

	genDocBytes, err := gen.EncodeAsJSON()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	assertT.EqualValues(parsedGenesisDoc.ChainID, gen.genesisDoc.ChainID)
	assertT.EqualValues(parsedGenesisDoc.ConsensusParams, gen.genesisDoc.ConsensusParams)
	assertT.EqualValues(parsedGenesisDoc.GenesisTime, gen.genesisDoc.GenesisTime)
	assertT.EqualValues(parsedGenesisDoc.InitialHeight, gen.genesisDoc.InitialHeight)
	assertT.EqualValues(parsedGenesisDoc.Validators, gen.genesisDoc.Validators)

	// In order to compare app state, we need to unmarshal it first
	// because comparing json.RawMessage may give false negatives.
	appStateMap := map[string]interface{}{}
	err = json.Unmarshal(gen.genesisDoc.AppState, &appStateMap)
	requireT.NoError(err)
	parsedAppStateMap := map[string]interface{}{}
	err = json.Unmarshal(parsedGenesisDoc.AppState, &parsedAppStateMap)
	requireT.NoError(err)
	assertT.EqualValues(appStateMap, parsedAppStateMap)

	var appStateMapJSONRawMessage map[string]json.RawMessage
	err = json.Unmarshal(gen.genesisDoc.AppState, &appStateMapJSONRawMessage)
	requireT.NoError(err)
	requireT.NoError(
		app.ModuleBasics.ValidateGenesis(
			encCfg.Marshaler,
			encCfg.TxConfig,
			appStateMapJSONRawMessage,
		))
}

func TestAddFundsToGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := NetworkByChainID(Devnet)
	requireT.NoError(err)

	gen, err := n.Genesis()
	requireT.NoError(err)

	pubKey, _ := types.GenerateSecp256k1Key()
	requireT.NoError(gen.FundAccount(pubKey, "1000someTestToken"))
	key1 := cosmossecp256k1.PubKey{Key: pubKey}
	accountAddress := sdk.AccAddress(key1.Address())

	pubKey2, _ := types.GenerateSecp256k1Key()
	requireT.NoError(gen.FundAccount(pubKey2, "2000someTestToken"))
	key2 := cosmossecp256k1.PubKey{Key: pubKey2}
	accountAddress2 := sdk.AccAddress(key2.Address())

	genDocBytes, err := gen.EncodeAsJSON()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	type coin struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	}
	type balance struct {
		Address string `json:"address"`
		Coins   []coin `json:"coins"`
	}
	type account struct {
		Address string `json:"address"`
	}
	var state struct {
		Bank struct {
			Balances []balance `json:"balances"`
			Supply   []coin    `json:"supply"`
		} `json:"bank"`
		Auth struct {
			Accounts []account `json:"accounts"`
		} `json:"auth"`
	}

	err = json.Unmarshal(parsedGenesisDoc.AppState, &state)
	requireT.NoError(err)

	assertT.ElementsMatch(state.Bank.Balances, []balance{
		{
			Address: accountAddress.String(),
			Coins: []coin{
				{Denom: "someTestToken", Amount: "1000"},
			},
		},
		{
			Address: accountAddress2.String(),
			Coins: []coin{
				{Denom: "someTestToken", Amount: "2000"},
			},
		},
	})

	assertT.Contains(
		state.Bank.Supply,
		coin{Denom: "someTestToken", Amount: "3000"},
	)
	requireT.Len(state.Auth.Accounts, 2)
	assertT.ElementsMatch(state.Auth.Accounts, []account{
		{Address: accountAddress.String()},
		{Address: accountAddress2.String()},
	})
}

func TestAddGenTx(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := NetworkByChainID(Devnet)
	requireT.NoError(err)

	gen, err := n.Genesis()
	requireT.NoError(err)

	pubKey, privKey := types.GenerateSecp256k1Key()

	encodingConfig := cosmoscmd.MakeEncodingConfig(app.ModuleBasics)
	clientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithBroadcastMode(flags.BroadcastBlock)
	tx, err := GenerateAddValidatorTx(clientCtx, ed25519.PublicKey(pubKey), privKey, "1000core")
	requireT.NoError(err)
	gen.AddGenesisTx(tx)

	genDocBytes, err := gen.EncodeAsJSON()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	var state struct {
		GenUtil struct {
			GenTxs []json.RawMessage `json:"gen_txs"` //nolint:tagliatelle
		} `json:"genutil"`
	}

	err = json.Unmarshal(parsedGenesisDoc.AppState, &state)
	requireT.NoError(err)
	assertT.Len(state.GenUtil.GenTxs, 1)
}
