//go:build integrationtests

package upgrade

import (
	"strings"
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
	switch chain.ChainSettings.ChainID {
	case string(constant.ChainIDMain):
		requireT.Equal(strings.ToUpper(constant.DenomMainDisplay), denomMetadata.Metadata.Symbol)
	case string(constant.ChainIDTest):
		requireT.Equal(strings.ToUpper(constant.DenomTestDisplay), denomMetadata.Metadata.Symbol)
	case string(constant.ChainIDDev):
		requireT.Equal(strings.ToUpper(constant.DenomDevDisplay), denomMetadata.Metadata.Symbol)
	default:
		requireT.FailNowf("unknown chain id: %s", chain.ChainSettings.ChainID)
	}
}

func (d *denomSymbol) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	client := banktypes.NewQueryClient(chain.ClientContext)
	denomMetadata, err := client.DenomMetadata(ctx, &banktypes.QueryDenomMetadataRequest{
		Denom: chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	switch chain.ChainSettings.ChainID {
	case string(constant.ChainIDMain):
		requireT.Equal("TX", denomMetadata.Metadata.Symbol)
	case string(constant.ChainIDTest):
		requireT.Equal("TESTTX", denomMetadata.Metadata.Symbol)
	case string(constant.ChainIDDev):
		requireT.Equal("DEVTX", denomMetadata.Metadata.Symbol)
	default:
		requireT.FailNowf("unknown chain id: %s", chain.ChainSettings.ChainID)
	}
}
