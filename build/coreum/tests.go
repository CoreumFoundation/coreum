package coreum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/types"
	"github.com/CoreumFoundation/crust/infra"
	"github.com/CoreumFoundation/crust/infra/apps"
	"github.com/CoreumFoundation/crust/pkg/znet"
)

// Test names.
const (
	TestIBC     = "ibc"
	TestModules = "modules"
	TestUpgrade = "upgrade"
)

// RunAllIntegrationTests runs all the coreum integration tests.
func RunAllIntegrationTests(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		entries, err := os.ReadDir(testsDir)
		if err != nil {
			return errors.WithStack(err)
		}

		actions := make([]types.CommandFunc, 0, len(entries))
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

func RunIntegrationTestsModules(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileModulesSmartContracts, CompileAssetExtensionSmartContracts, BuildCoredLocally,
			BuildCoredDockerImage)

		znetConfig := &infra.ConfigFactory{
			EnvName:            "znet",
			Profiles:           []string{apps.Profile3Cored},
			TimeoutCommit:      500 * time.Millisecond,
			HomeDir:            filepath.Join(lo.Must(os.UserHomeDir()), ".crust"),
			RootDir:            "../",
			CoverageOutputFile: "coverage/coreum-integration-tests-modules",
		}

		flags := []string{
			"-tags=integrationtests",
			fmt.Sprintf("-parallel=%d", 2*runtime.NumCPU()),
			"-timeout=1h",
		}
		if runUnsafe {
			flags = append(flags, "--run-unsafe")
		}

		if err := znet.Remove(ctx, znetConfig); err != nil {
			return err
		}
		if err := znet.Start(ctx, znetConfig); err != nil {
			return err
		}
		if err := golang.RunTests(ctx, deps, golang.TestConfig{
			PackagePath: filepath.Join(testsDir, "modules"),
			Flags:       flags,
		}); err != nil {
			return err
		}
		if err := znet.Stop(ctx, znetConfig); err != nil {
			return err
		}
		if err := znet.CoverageConvert(ctx, znetConfig); err != nil {
			return err
		}
		return znet.Remove(ctx, znetConfig)
	}
}

// RunIntegrationTests returns function running integration tests.
func RunIntegrationTests(name string, runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		switch name {
		case TestModules:
			deps(CompileModulesSmartContracts, CompileAssetExtensionSmartContracts)
		case TestUpgrade:
			deps(CompileModulesSmartContracts)
		case TestIBC:
			deps(CompileIBCSmartContracts, CompileAssetExtensionSmartContracts)
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
