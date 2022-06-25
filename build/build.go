package build

import (
	"context"
	"os"
	"runtime"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/pkg/errors"
)

const dockerGOOS = "linux"

// FIXME (wojciech): This assumes that repositories are public which is not true at the moment but will be soon
const coreumRepoURL = "https://github.com/CoreumFoundation/coreum.git"

func buildAll(deps build.DepsFunc) {
	deps(buildCored, buildCrustZNet, buildCrustZStress)
}

func buildCored(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo, ensureCoreumRepo)

	pkg := "../coreum/cmd/cored"

	// FIXME (wojciech): Remove this code once `cored` package is moved to root directory of the repository after migration
	if _, err := os.Stat(pkg); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithStack(err)
		}
		pkg = "../coreum/cored/cmd/cored"
	}

	return buildNativeAndDocker(ctx, pkg, "cored")
}

func buildCrustZNet(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return goBuildPkg(ctx, "crust/cmd/crustznet", runtime.GOOS, "bin/crustznet")
}

func buildCrustZStress(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return buildNativeAndDocker(ctx, "crust/cmd/crustzstress", "crustzstress")
}

func ensureAllRepos(deps build.DepsFunc) {
	deps(ensureCoreumRepo)
}

func ensureCoreumRepo(ctx context.Context) error {
	return ensureRepo(ctx, coreumRepoURL)
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
