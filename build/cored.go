package build

import (
	"context"

	"github.com/CoreumFoundation/coreum-build-tools/pkg/build"
)

func buildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "cored/cmd", "bin/cored")
}
