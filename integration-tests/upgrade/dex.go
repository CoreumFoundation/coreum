//go:build integrationtests

package upgrade

import (
	"testing"

	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

type dex struct{}

func (d *dex) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	_, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.ErrorContains(err, "unknown service coreum.dex.v1.Query")
}

func (d *dex) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	paramsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(dextypes.DefaultParams(), paramsRes.Params)
}
