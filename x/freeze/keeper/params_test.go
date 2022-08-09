package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "github.com/CoreumFoundation/coreum/testutil/keeper"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.FreezeKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
