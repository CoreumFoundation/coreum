package build

import (
	"context"

	"github.com/outofforest/build"
)

func buildNode(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "node/cmd", "bin/cored")
}
