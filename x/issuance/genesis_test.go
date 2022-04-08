package issuance_test

import (
	"testing"

	keepertest "github.com/coreumfoundation/coreum/coreum/testutil/keeper"
	"github.com/coreumfoundation/coreum/coreum/testutil/nullify"
	"github.com/coreumfoundation/coreum/coreum/x/issuance"
	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.IssuanceKeeper(t)
	issuance.InitGenesis(ctx, *k, genesisState)
	got := issuance.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
