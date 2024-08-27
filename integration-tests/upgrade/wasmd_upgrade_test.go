//go:build integrationtests

package upgrade

import (
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
)

type wasmdUpgradeTest struct {
}

func (wut *wasmdUpgradeTest) Before(t *testing.T) {
	assertDependencyVersion(
		t,
		"github.com/CosmWasm/wasmd",
		"v0.45.",
	)
	assertDependencyVersion(
		t,
		"github.com/CosmWasm/wasmvm",
		"v1.5.2",
	)
}

func (wut *wasmdUpgradeTest) After(t *testing.T) {
	assertDependencyVersion(
		t,
		"github.com/CosmWasm/wasmd",
		"v0.46.0",
	)
	assertDependencyVersion(
		t,
		"github.com/CosmWasm/wasmvm",
		"v1.5.4",
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
