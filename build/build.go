package build

import (
	"context"
	"os"
	"runtime"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
)

const dockerGOOS = "linux"

func buildAll(deps build.DepsFunc) {
	deps(buildCored, buildCoreZNet)
}

func buildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	out := "bin/" + runtime.GOOS + "/cored"
	link := "bin/cored"
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := goBuildPkg(ctx, "cored/cmd", runtime.GOOS, out); err != nil {
		return err
	}
	if err := os.Link(out, link); err != nil {
		return err
	}
	if runtime.GOOS != dockerGOOS {
		// required to build docker images
		return goBuildPkg(ctx, "cored/cmd", dockerGOOS, "bin/"+dockerGOOS+"/cored")
	}
	return nil
}

func buildCoreZNet(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "coreznet/cmd", runtime.GOOS, "bin/coreznet")
}
