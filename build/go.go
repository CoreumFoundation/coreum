package build

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum-tools/pkg/libexec"
	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"go.uber.org/zap"
)

func ensureGo(ctx context.Context) error {
	return ensure(ctx, "go")
}

func ensureGolangCI(ctx context.Context) error {
	return ensure(ctx, "golangci")
}

// goBuildPkg builds go package
func goBuildPkg(ctx context.Context, pkg, out string) error {
	logger.Get(ctx).Info("Building go package", zap.String("package", pkg), zap.String("binary", out))
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-w -s", "-o", must.String(filepath.Abs(out)), ".")
	cmd.Dir = pkg
	cmd.Env = append([]string{"CGO_ENABLED=0"}, os.Environ()...)
	return libexec.Exec(ctx, cmd)
}

// goLint runs golangci linter, runs go mod tidy and checks that git status is clean
func goLint(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo, ensureGolangCI)
	log := logger.Get(ctx)
	config := must.String(filepath.Abs("build/.golangci.yaml"))
	err := onModule(func(path string) error {
		log.Info("Running linter", zap.String("path", path))
		cmd := exec.Command("golangci-lint", "run", "--config", config)
		cmd.Dir = path
		return libexec.Exec(ctx, cmd)
	})
	if err != nil {
		return err
	}
	deps(goModTidy, gitStatusClean)
	return nil
}

// goTest runs go test
func goTest(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	log := logger.Get(ctx)
	return onModule(func(path string) error {
		log.Info("Running go tests", zap.String("path", path))
		cmd := exec.Command("go", "test", "-count=1", "-shuffle=on", "-race", "./...")
		cmd.Dir = path
		return libexec.Exec(ctx, cmd)
	})
}

func goModTidy(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	log := logger.Get(ctx)
	return onModule(func(path string) error {
		log.Info("Running go mod tidy", zap.String("path", path))
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = path
		return libexec.Exec(ctx, cmd)
	})
}

func onModule(fn func(path string) error) error {
	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || d.Name() != "go.mod" {
			return nil
		}
		return fn(filepath.Dir(path))
	})
}
