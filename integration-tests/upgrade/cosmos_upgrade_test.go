//go:build integrationtests

package upgrade

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
)

type cosmosUpgradeTest struct {
}

func (cut *cosmosUpgradeTest) Before(t *testing.T) {
	assertDependencyVersion(
		t,
		"github.com/cosmos/cosmos-sdk",
		"v0.47.12",
	)
}

func (cut *cosmosUpgradeTest) After(t *testing.T) {
	assertDependencyVersion(
		t,
		"github.com/cosmos/cosmos-sdk",
		"v0.47.15",
	)
}

func assertDependencyVersion(
	t *testing.T,
	depName string,
	versionPrefix string,
) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	cmtClient := tmservice.NewServiceClient(chain.ClientContext)
	nodeInfo, err := cmtClient.GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	requireT.NoError(err)
	for _, dep := range nodeInfo.ApplicationVersion.BuildDeps {
		if dep.Path == depName {
			t.Logf("dep %s. version:%s. expected version:%s", dep.Path, dep.Version, versionPrefix)
			requireT.True(strings.HasPrefix(dep.Version, versionPrefix))
			return
		}
	}
	requireT.Failf("dependency %s not found", depName)
}
