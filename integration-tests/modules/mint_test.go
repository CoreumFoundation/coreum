//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
)

// TestMintQueryInflation tests that querying inflation works.
func TestMintQueryInflation(t *testing.T) {
	t.Parallel()

	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	mintClient := minttypes.NewQueryClient(chain.ClientContext)
	resp, err := mintClient.Inflation(ctx, &minttypes.QueryInflationRequest{})
	requireT.NoError(err)
	requireT.True(resp.Inflation.GT(sdkmath.LegacyZeroDec()))
}
