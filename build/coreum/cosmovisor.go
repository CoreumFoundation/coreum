package coreum

import (
	"context"
	"path/filepath"

	"github.com/CoreumFoundation/crust/build/tools"
)

func ensureCosmovisorWithInstalledBinary(ctx context.Context, platform tools.TargetPlatform, binaryName string) error {
	if err := tools.Ensure(ctx, tools.Cosmovisor, platform); err != nil {
		return err
	}

	return tools.CopyToolBinaries(tools.Cosmovisor,
		platform,
		filepath.Join("bin", ".cache", binaryName, platform.String()),
		cosmovisorBinaryPath)
}
