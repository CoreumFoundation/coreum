package coreum

import (
	"context"
	"path/filepath"

	"github.com/CoreumFoundation/coreum/build/tools"
	buildtools "github.com/CoreumFoundation/crust/build/tools"
)

func ensureCosmovisorWithInstalledBinary(ctx context.Context, platform buildtools.TargetPlatform, binaryName string) error {
	if err := buildtools.Ensure(ctx, tools.Cosmovisor, platform); err != nil {
		return err
	}

	return CopyToolBinaries(tools.Cosmovisor,
		platform,
		filepath.Join("bin", ".cache", binaryName, platform.String()),
		cosmovisorBinaryPath)
}
