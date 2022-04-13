package build

import "context"

func buildNode(ctx context.Context) error {
	return goBuildPkg(ctx, "node/cmd", "bin/cored")
}
