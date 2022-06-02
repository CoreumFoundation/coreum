package build

import (
	"context"
	"os"
	"runtime"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/pkg/errors"
)

const dockerGOOS = "linux"

func buildAll(deps build.DepsFunc) {
	deps(buildCored, buildCoreZNet, buildCoreZStress)
}

func buildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return buildNativeAndDocker(ctx, "cored/cmd", "cored")
}

func buildCoreZNet(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "coreznet/cmd/coreznet", runtime.GOOS, "bin/coreznet")
}

func buildCoreZStress(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return buildNativeAndDocker(ctx, "coreznet/cmd/corezstress", "corezstress")
}

func buildNativeAndDocker(ctx context.Context, pkg, exeName string) error {
	out := "bin/" + runtime.GOOS + "/" + exeName
	link := "bin/" + exeName
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}
	if err := goBuildPkg(ctx, pkg, runtime.GOOS, out); err != nil {
		return err
	}
	if err := os.Link(out, link); err != nil {
		return errors.WithStack(err)
	}
	if runtime.GOOS != dockerGOOS {
		// required to build docker images
		return goBuildPkg(ctx, pkg, dockerGOOS, "bin/"+dockerGOOS+"/"+exeName)
	}
	return nil
}
