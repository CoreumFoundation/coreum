package coreum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/golang"
)

// Test names.
const (
	TestIBC     = "ibc"
	TestModules = "modules"
	TestUpgrade = "upgrade"
)

// RunAllIntegrationTests runs all the coreum integration tests.
func RunAllIntegrationTests(runUnsafe bool) build.CommandFunc {
	return func(ctx context.Context, deps build.DepsFunc) error {
		entries, err := os.ReadDir(testsDir)
		if err != nil {
			return errors.WithStack(err)
		}

		actions := make([]build.CommandFunc, 0, len(entries))
		for _, e := range entries {
			if !e.IsDir() || e.Name() == "contracts" {
				continue
			}

			actions = append(actions, RunIntegrationTests(e.Name(), runUnsafe))
		}
		deps(actions...)
		return nil
	}
}

// RunIntegrationTests returns function running integration tests.
func RunIntegrationTests(name string, runUnsafe bool) build.CommandFunc {
	return func(ctx context.Context, deps build.DepsFunc) error {
		switch name {
		case TestModules:
			deps(CompileModulesSmartContracts, CompileAssetExtensionSmartContracts)
		case TestUpgrade:
			deps(CompileModulesSmartContracts)
		case TestIBC:
			deps(CompileIBCSmartContracts)
		}

		flags := []string{
			"-tags=integrationtests",
			fmt.Sprintf("-parallel=%d", 2*runtime.NumCPU()),
			"-timeout=1h",
		}
		if runUnsafe {
			flags = append(flags, "--run-unsafe")
		}
		return golang.RunTests(ctx, deps, golang.TestConfig{
			PackagePath: filepath.Join(testsDir, name),
			Flags:       flags,
		})
	}
}
