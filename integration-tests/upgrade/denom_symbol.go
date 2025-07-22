//go:build integrationtests

package upgrade

import (
	"fmt"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

type denomSymbol struct {
}

func (d *denomSymbol) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	client := banktypes.NewQueryClient(chain.ClientContext)
	denomMetadata, err := client.DenomMetadata(ctx, &banktypes.QueryDenomMetadataRequest{
		Denom: chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	prefix := ""
	if chain.ChainSettings.ChainID == string(constant.ChainIDTest) {
		prefix = "test"
	} else if chain.ChainSettings.ChainID == string(constant.ChainIDDev) {
		prefix = "dev"
	}

	requireT.Equal(denomMetadata.Metadata.Description, fmt.Sprintf("%score coin", prefix))
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom: fmt.Sprintf("u%score", prefix),
	})
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom:    fmt.Sprintf("%score", prefix),
		Exponent: 6,
	})
	requireT.Equal(denomMetadata.Metadata.Base, fmt.Sprintf("u%score", prefix))
	requireT.Equal(denomMetadata.Metadata.Display, fmt.Sprintf("%score", prefix))
	requireT.Equal(denomMetadata.Metadata.Name, fmt.Sprintf("u%score", prefix))
	requireT.Equal(denomMetadata.Metadata.Symbol, fmt.Sprintf("u%score", prefix))
}

func (d *denomSymbol) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	client := banktypes.NewQueryClient(chain.ClientContext)
	denomMetadata, err := client.DenomMetadata(ctx, &banktypes.QueryDenomMetadataRequest{
		Denom: chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	prefix := ""
	if chain.ChainSettings.ChainID == string(constant.ChainIDTest) {
		prefix = "test"
	} else if chain.ChainSettings.ChainID == string(constant.ChainIDDev) {
		prefix = "dev"
	}

	requireT.Equal(denomMetadata.Metadata.Description, fmt.Sprintf("%stx coin", prefix))
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom: fmt.Sprintf("u%stx", prefix),
	})
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom:    fmt.Sprintf("%stx", prefix),
		Exponent: 6,
	})
	requireT.Equal(denomMetadata.Metadata.Base, fmt.Sprintf("u%stx", prefix))
	requireT.Equal(denomMetadata.Metadata.Display, fmt.Sprintf("%stx", prefix))
	requireT.Equal(denomMetadata.Metadata.Name, fmt.Sprintf("u%stx", prefix))
	requireT.Equal(denomMetadata.Metadata.Symbol, fmt.Sprintf("u%stx", prefix))
}
