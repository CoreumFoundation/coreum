package coreum

import (
	"context"
	"path/filepath"

	coreumtools "github.com/CoreumFoundation/coreum/build/tools"
	crusttools "github.com/CoreumFoundation/crust/build/tools"
)

func ensureCosmovisorWithInstalledBinary(
	ctx context.Context, platform crusttools.TargetPlatform, binaryName string,
) error {
	if err := crusttools.Ensure(ctx, coreumtools.Cosmovisor, platform); err != nil {
		return err
	}

	return CopyToolBinaries(coreumtools.Cosmovisor,
		platform,
		filepath.Join("bin", ".cache", binaryName, platform.String()),
		cosmovisorBinaryPath)
}
