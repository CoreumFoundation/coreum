package coreum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/samber/lo"

	"github.com/CoreumFoundation/crust/build/gaia"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/hermes"
	"github.com/CoreumFoundation/crust/build/osmosis"
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
		deps(
			RunIntegrationTestsModules(runUnsafe),
			RunIntegrationTestsIBC(runUnsafe),
			RunIntegrationTestsUpgrade(runUnsafe),
		)
		return nil
	}
}

// RunIntegrationTestsModules returns function running modules integration tests.
func RunIntegrationTestsModules(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileModulesSmartContracts, CompileAssetExtensionSmartContracts, BuildCoredLocally,
			BuildCoredDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored}
		znetConfig.CoverageOutputFile = "coverage/coreum-integration-tests-modules"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestModules)
	}
}

// RunIntegrationTestsIBC returns function running IBC integration tests.
func RunIntegrationTestsIBC(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileIBCSmartContracts, CompileAssetExtensionSmartContracts, BuildCoredLocally,
			BuildCoredDockerImage, gaia.BuildDockerImage, osmosis.BuildDockerImage, hermes.BuildDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored, apps.ProfileIBC}

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestIBC)
	}
}

// RunIntegrationTestsUpgrade returns function running upgrade integration tests.
func RunIntegrationTestsUpgrade(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileIBCSmartContracts, CompileAssetExtensionSmartContracts, CompileModulesSmartContracts,
			BuildCoredLocally, BuildCoredDockerImage, gaia.BuildDockerImage, osmosis.BuildDockerImage,
			hermes.BuildDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored, apps.ProfileIBC}
		znetConfig.CoredVersion = "v3.0.3"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestUpgrade, TestIBC, TestModules)
	}
}

func runIntegrationTests(
	ctx context.Context,
	deps types.DepsFunc,
	runUnsafe bool,
	znetConfig *infra.ConfigFactory,
	testDirs ...string,
) error {
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

	for _, testDir := range testDirs {
		if err := golang.RunTests(ctx, deps, golang.TestConfig{
			PackagePath: filepath.Join(testsDir, testDir),
			Flags:       flags,
		}); err != nil {
			return err
		}
	}

	if znetConfig.CoverageOutputFile != "" {
		if err := znet.Stop(ctx, znetConfig); err != nil {
			return err
		}
		if err := znet.CoverageConvert(ctx, znetConfig); err != nil {
			return err
		}
	}

	return znet.Remove(ctx, znetConfig)
}

func defaultZNetConfig() *infra.ConfigFactory {
	return &infra.ConfigFactory{
		EnvName:       "znet",
		TimeoutCommit: 500 * time.Millisecond,
		HomeDir:       filepath.Join(lo.Must(os.UserHomeDir()), ".crust", "znet"),
		RootDir:       ".",
	}
}
