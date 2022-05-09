package build

import (
	"context"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
)

func buildAll(deps build.DepsFunc) {
	deps(buildCored, buildCoreZNet)
}

func buildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "cored/cmd", "bin/cored")
}

func buildCoreZNet(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "coreznet/cmd", "bin/coreznet")
}
