package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestGenesisValidation(t *testing.T) {
	assertion := assert.New(t)
	required := require.New(t)
	dirPath, err := ioutil.TempDir("", "genesis_test")
	required.NoError(err)
	defer required.NoError(os.RemoveAll(dirPath))

	n, err := NetworkByChainID(string(Devnet))
	required.NoError(err)

	gen, err := n.Genesis()
	required.NoError(err)
	required.NoError(gen.Save(dirPath))

	parsedGenesisDoc, err := tmtypes.GenesisDocFromFile(dirPath + "/config/genesis.json")
	required.NoError(err)
	assertion.EqualValues(parsedGenesisDoc.ChainID, gen.genesisDoc.ChainID)
	assertion.EqualValues(parsedGenesisDoc.ConsensusParams, gen.genesisDoc.ConsensusParams)
	assertion.EqualValues(parsedGenesisDoc.GenesisTime, gen.genesisDoc.GenesisTime)
	assertion.EqualValues(parsedGenesisDoc.InitialHeight, gen.genesisDoc.InitialHeight)
	assertion.EqualValues(parsedGenesisDoc.Validators, gen.genesisDoc.Validators)

	// In order to compare app state, we need to unmarshal it first
	// because comparing json.RawMessage may give false negatives.
	appStateMap := map[string]interface{}{}
	err = json.Unmarshal(gen.genesisDoc.AppState, &appStateMap)
	assertion.NoError(err)
	parsedAppStateMap := map[string]interface{}{}
	err = json.Unmarshal(parsedGenesisDoc.AppState, &parsedAppStateMap)
	assertion.NoError(err)
	assertion.EqualValues(appStateMap, parsedAppStateMap)
}
