package build

import (
	"context"
	"fmt"
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
func goBuildPkg(ctx context.Context, pkg, targetOS, out string) error {
	logger.Get(ctx).Info("Building go package", zap.String("package", pkg), zap.String("binary", out), zap.String("targetOS", targetOS))
	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-w -s", "-o", must.String(filepath.Abs(out)), ".")
	cmd.Dir = pkg
	cmd.Env = append([]string{"CGO_ENABLED=0", "GOOS=" + targetOS}, os.Environ()...)
	if err := libexec.Exec(ctx, cmd); err != nil {
		return fmt.Errorf("building go package '%s' failed: %w", pkg, err)
	}
	return nil
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
		if err := libexec.Exec(ctx, cmd); err != nil {
			return fmt.Errorf("linter errors found in module '%s': %w", path, err)
		}
		return nil
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
		if err := libexec.Exec(ctx, cmd); err != nil {
			return fmt.Errorf("unit tests failed in module '%s': %w", path, err)
		}
		return nil
	})
}

func goModTidy(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureGo)
	log := logger.Get(ctx)
	return onModule(func(path string) error {
		log.Info("Running go mod tidy", zap.String("path", path))
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = path
		if err := libexec.Exec(ctx, cmd); err != nil {
			return fmt.Errorf("'go mod tidy' failed in module '%s': %w", path, err)
		}
		return nil
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
