package config

import (
	"encoding/json"
	"testing"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ignite-hq/cli/ignite/pkg/cosmoscmd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

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
	assertT.NoError(err)
	parsedAppStateMap := map[string]interface{}{}
	err = json.Unmarshal(parsedGenesisDoc.AppState, &parsedAppStateMap)
	assertT.NoError(err)
	assertT.EqualValues(appStateMap, parsedAppStateMap)

	var appStateMapJSONRawMessage map[string]json.RawMessage
	err = json.Unmarshal(gen.genesisDoc.AppState, &appStateMapJSONRawMessage)
	assertT.NoError(err)
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

	secp256k1 := cosmossecp256k1.PubKey{Key: pubKey}
	accountAddress := sdk.AccAddress(secp256k1.Address())

	genDocBytes, err := gen.EncodeAsJSON()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	type balance struct {
		Address string `json:"address"`
		Coins   []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"coins"`
	}
	var state struct {
		Bank struct {
			Balances []balance `json:"balances"`
		} `json:"bank"`
	}

	err = json.Unmarshal(parsedGenesisDoc.AppState, &state)
	requireT.NoError(err)

	assertT.Contains(state.Bank.Balances, balance{
		Address: accountAddress.String(),
		Coins: []struct {
			Denom  string "json:\"denom\""
			Amount string "json:\"amount\""
		}{
			{Denom: "someTestToken", Amount: "1000"},
		},
	})
}
