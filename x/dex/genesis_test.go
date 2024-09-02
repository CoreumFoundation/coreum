package dex_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	prams := dextypes.DefaultParams()
	genState := dextypes.GenesisState{
		Params: prams,
	}

	// init the keeper
	dex.InitGenesis(ctx, dexKeeper, genState)

	// check imported state
	params := dexKeeper.GetParams(ctx)
	requireT.EqualValues(prams, params)

	// check that export is equal import
	exportedGenState := dex.ExportGenesis(ctx, dexKeeper)

	requireT.EqualValues(genState.Params, exportedGenState.Params)
}
