package keeper_test

import (
	"testing"

	testkeeper "github.com/coreumfoundation/coreum/coreum/testutil/keeper"
	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.IssuanceKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
