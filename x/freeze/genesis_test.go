package freeze_test

import (
	"testing"

	keepertest "github.com/CoreumFoundation/coreum/testutil/keeper"
	"github.com/CoreumFoundation/coreum/testutil/nullify"
	"github.com/CoreumFoundation/coreum/x/freeze"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.FreezeKeeper(t)
	freeze.InitGenesis(ctx, k, genesisState)
	got := freeze.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
