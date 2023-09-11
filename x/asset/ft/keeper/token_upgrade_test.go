package keeper_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
)

func TestOneTokenUpgradeAtATimeIsAllowed(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	ftKeeper := testApp.AssetFTKeeper

	// first call succeeds
	requireT.NoError(ftKeeper.SetPendingVersion(ctx, "denom", 1))

	// second call is rejected
	requireT.Error(ftKeeper.SetPendingVersion(ctx, "denom", 1))

	// second call is rejected even if version is higher
	requireT.Error(ftKeeper.SetPendingVersion(ctx, "denom", 2))

	// but it should succeed for another denom
	requireT.NoError(ftKeeper.SetPendingVersion(ctx, "denom2", 1))

	// upgrade happened
	ftKeeper.ClearPendingVersion(ctx, "denom")

	// but for second denom it should still fail
	requireT.Error(ftKeeper.SetPendingVersion(ctx, "denom2", 2))

	// for first denom it should work now
	requireT.NoError(ftKeeper.SetPendingVersion(ctx, "denom", 1))
}
