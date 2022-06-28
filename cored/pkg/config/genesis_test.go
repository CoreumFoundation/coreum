package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestGenesisValidation(t *testing.T) {
	assert := assert.New(t)
	dirPath, err := ioutil.TempDir("", "genesis_test")
	defer os.RemoveAll(dirPath)

	n, err := NetworkByChainID(string(Devnet))
	assert.NoError(err)

	gen, err := n.Genesis()
	assert.NoError(err)
	err = gen.Save(dirPath)
	assert.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromFile(dirPath + "/config/genesis.json")
	assert.NoError(err)
	assert.EqualValues(parsedGenesisDoc.ChainID, gen.genesisDoc.ChainID)
	assert.EqualValues(parsedGenesisDoc.ConsensusParams, gen.genesisDoc.ConsensusParams)
	assert.EqualValues(parsedGenesisDoc.GenesisTime, gen.genesisDoc.GenesisTime)
	assert.EqualValues(parsedGenesisDoc.InitialHeight, gen.genesisDoc.InitialHeight)
	assert.EqualValues(parsedGenesisDoc.Validators, gen.genesisDoc.Validators)

	// In order to compare app state, we need to unmarshal it first
	// because comparing json.RawMessage may give false negatives.
	appStateMap := map[string]interface{}{}
	err = json.Unmarshal(gen.genesisDoc.AppState, &appStateMap)
	assert.NoError(err)
	parsedAppStateMap := map[string]interface{}{}
	err = json.Unmarshal(parsedGenesisDoc.AppState, &parsedAppStateMap)
	assert.NoError(err)
	assert.EqualValues(appStateMap, parsedAppStateMap)
}
