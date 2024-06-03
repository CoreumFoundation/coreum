package coreum

import (
	"context"
	"path/filepath"

	"github.com/CoreumFoundation/crust/build/tools"
)

func ensureCosmovisor(ctx context.Context, platform tools.TargetPlatform) error {
	if err := tools.Ensure(ctx, tools.Cosmovisor, platform); err != nil {
		return err
	}

	for _, target := range []string{"cored", "cored-ext"} {
		err := tools.CopyToolBinaries(tools.Cosmovisor,
			platform,
			filepath.Join("bin", ".cache", target, platform.String()),
			cosmovisorBinaryPath)
		if err != nil {
			return err
		}
	}

	return nil
}
