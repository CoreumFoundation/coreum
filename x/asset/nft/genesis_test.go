package nft_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/nft"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	// prepare the genesis data

	// class definitions
	var classDefinitions []types.ClassDefinition
	for i := 0; i < 5; i++ {
		classDefinition := types.ClassDefinition{
			ID: fmt.Sprintf("id%d", i),
			Features: []types.ClassFeature{
				types.ClassFeature_burning, //nolint:nosnakecase // proto enum
			},
		}
		classDefinitions = append(classDefinitions, classDefinition)
	}

	genState := types.GenesisState{
		Params:           types.DefaultParams(),
		ClassDefinitions: classDefinitions,
	}

	// init the keeper
	nft.InitGenesis(ctx, nftKeeper, genState)

	// assert the keeper state

	// class definitions
	for _, definition := range classDefinitions {
		storedDefinition, err := nftKeeper.GetClassDefinition(ctx, definition.ID)
		requireT.NoError(err)
		assertT.EqualValues(definition, storedDefinition)
	}

	// params
	params := nftKeeper.GetParams(ctx)
	assertT.EqualValues(types.DefaultParams(), params)

	// check that export is equal import
	exportedGenState := nft.ExportGenesis(ctx, nftKeeper)
	assertT.ElementsMatch(genState.ClassDefinitions, exportedGenState.ClassDefinitions)
}
