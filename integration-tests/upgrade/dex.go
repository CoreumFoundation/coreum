//go:build integrationtests

package upgrade

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
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
	t.Logf("DEX params after upgrade: %v", paramsRes.Params)

	expectedParams := dextypes.DefaultParams()
	expectedParams.OrderReserve = sdk.NewInt64Coin(chain.ChainSettings.Denom, 10_000_000)
	requireT.Equal(expectedParams, paramsRes.Params)
}
