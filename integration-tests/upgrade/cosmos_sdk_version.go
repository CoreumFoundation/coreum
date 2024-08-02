//go:build integrationtests

package upgrade

import (
	"context"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

type cosmosSDKVersion struct {
	token assetfttypes.Token
}

func (ftt *cosmosSDKVersion) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assertCosmosSDKVersion(ctx, chain.ClientContext, requireT, "v0.47.")
}

func (ftt *cosmosSDKVersion) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	assertCosmosSDKVersion(ctx, chain.ClientContext, requireT, "v0.50.")
}

func assertCosmosSDKVersion(ctx context.Context, clientCtx client.Context, requireT *require.Assertions, prefix string) {
	cmtClient := cmtservice.NewServiceClient(clientCtx)
	nodeInfo, err := cmtClient.GetNodeInfo(ctx, &cmtservice.GetNodeInfoRequest{})
	requireT.NoError(err)
	requireT.True(strings.HasPrefix(nodeInfo.ApplicationVersion.CosmosSdkVersion, prefix))
}
