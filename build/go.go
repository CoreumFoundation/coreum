package build

import (
	"context"
	"os"
	"os/exec"

	"github.com/CoreumFoundation/coreum-build-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-build-tools/pkg/libexec"
)

func ensureGo(ctx context.Context) error {
	return ensure(ctx, "go")
}

func ensureGolangCI(ctx context.Context) error {
	return ensure(ctx, "golangci")
}

// goBuildPkg builds go package
func goBuildPkg(ctx context.Context, pkg, out string) error {
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-w -s", "-o", out, "./"+pkg)
	cmd.Env = append([]string{"CGO_ENABLED=0"}, os.Environ()...)
	return libexec.Exec(ctx, cmd)
}

// goLint runs golangci linter, runs go mod tidy and checks that git status is clean
func goLint(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo, ensureGolangCI)
	if err := libexec.Exec(ctx, exec.Command("golangci-lint", "run", "--config", "build/.golangci.yaml")); err != nil {
		return err
	}
	deps(goModTidy, gitStatusClean)
	return nil
}

// goTest runs go test
func goTest(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return libexec.Exec(ctx, exec.Command("go", "test", "-count=1", "-shuffle=on", "-race", "./..."))
}

func goModTidy(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	return libexec.Exec(ctx, exec.Command("go", "mod", "tidy"))
}
