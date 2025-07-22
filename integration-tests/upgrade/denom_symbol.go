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

//nolint:dupl
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

	requireT.Equal(prefix+"core coin", denomMetadata.Metadata.Description)
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom: fmt.Sprintf("u%score", prefix),
	})
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom:    prefix + "core",
		Exponent: 6,
	})
	requireT.Equal(fmt.Sprintf("u%score", prefix), denomMetadata.Metadata.Base)
	requireT.Equal(prefix+"core", denomMetadata.Metadata.Display)
	requireT.Equal(fmt.Sprintf("u%score", prefix), denomMetadata.Metadata.Name)
	requireT.Equal(fmt.Sprintf("u%score", prefix), denomMetadata.Metadata.Symbol)
}

//nolint:dupl
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

	panic(fmt.Sprintf("%+v", denomMetadata.Metadata))
	requireT.Equal(prefix+"tx coin", denomMetadata.Metadata.Description)
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom: fmt.Sprintf("u%stx", prefix),
	})
	requireT.Contains(denomMetadata.Metadata.DenomUnits, &banktypes.DenomUnit{
		Denom:    prefix + "tx",
		Exponent: 6,
	})
	requireT.Equal(fmt.Sprintf("u%stx", prefix), denomMetadata.Metadata.Base)
	requireT.Equal(prefix+"tx", denomMetadata.Metadata.Display)
	requireT.Equal(fmt.Sprintf("u%stx", prefix), denomMetadata.Metadata.Name)
	requireT.Equal(fmt.Sprintf("u%stx", prefix), denomMetadata.Metadata.Symbol)
}
