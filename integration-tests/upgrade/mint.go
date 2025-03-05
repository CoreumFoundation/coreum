//go:build integrationtests

package upgrade

import (
	"testing"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
)

type mint struct {
}

func (m *mint) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	client := minttypes.NewQueryClient(chain.ClientContext)
	params, err := client.Params(ctx, &minttypes.QueryParamsRequest{})
	requireT.NoError(err)
	oldMaxInflation, err := params.Params.InflationMax.Float64()
	requireT.NoError(err)
	requireT.InDelta(float64(0.20), oldMaxInflation, 0.01)
	requireT.EqualValues(17_900_000, params.Params.BlocksPerYear)

	inflation, err := client.Inflation(ctx, &minttypes.QueryInflationRequest{})
	requireT.NoError(err)
	oldInflation, err := inflation.Inflation.Float64()
	requireT.NoError(err)
	requireT.InDelta(float64(0.10), oldInflation, 0.01)
}

func (m *mint) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	client := minttypes.NewQueryClient(chain.ClientContext)
	params, err := client.Params(ctx, &minttypes.QueryParamsRequest{})
	requireT.NoError(err)
	newMaxInflation, err := params.Params.InflationMax.Float64()
	requireT.NoError(err)
	requireT.InDelta(float64(0.30), newMaxInflation, 0.01)

	inflation, err := client.Inflation(ctx, &minttypes.QueryInflationRequest{})
	requireT.NoError(err)
	inflationFloat, err := inflation.Inflation.Float64()
	requireT.NoError(err)
	requireT.InDelta(float64(0.26), inflationFloat, 0.01)
	requireT.EqualValues(28_700_000, params.Params.BlocksPerYear)
}
