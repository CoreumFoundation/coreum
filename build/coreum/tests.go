package coreum

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/samber/lo"

	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/types"
	"github.com/CoreumFoundation/crust/znet/infra"
	"github.com/CoreumFoundation/crust/znet/infra/apps"
	"github.com/CoreumFoundation/crust/znet/pkg/znet"
)

// Test names.
const (
	TestIBC     = "ibc"
	TestModules = "modules"
	TestUpgrade = "upgrade"
	TestStress  = "stress"
	TestExport  = "export"
)

// Test run unit tests in coreum repo.
func Test(ctx context.Context, deps types.DepsFunc) error {
	deps(CompileAllSmartContracts)

	return golang.Test(ctx, deps)
}

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
		deps(CompileModulesSmartContracts, CompileLegacyModulesSmartContracts, CompileAssetExtensionSmartContracts,
			CompileDEXSmartContracts, BuildCoredLocally, BuildCoredDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored}
		znetConfig.CoverageOutputFile = "coverage/coreum-integration-tests-modules"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestModules)
	}
}

// RunIntegrationTestsStress returns function running stress integration tests.
func RunIntegrationTestsStress(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(BuildCoredLocally, BuildCoredDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored, apps.ProfileDEX}
		znetConfig.CoverageOutputFile = "coverage/coreum-integration-tests-stress"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestStress)
	}
}

// RunIntegrationTestsIBC returns function running IBC integration tests.
func RunIntegrationTestsIBC(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileIBCSmartContracts, CompileAssetExtensionSmartContracts, CompileDEXSmartContracts,
			BuildCoredLocally, BuildCoredDockerImage, BuildGaiaDockerImage, BuildOsmosisDockerImage,
			BuildHermesDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored, apps.ProfileIBC}

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestIBC)
	}
}

// RunIntegrationTestsUpgrade returns function running upgrade integration tests.
func RunIntegrationTestsUpgrade(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileIBCSmartContracts, CompileAssetExtensionSmartContracts, CompileDEXSmartContracts,
			CompileModulesSmartContracts, CompileLegacyModulesSmartContracts, BuildCoredLocally, BuildCoredDockerImage,
			BuildGaiaDockerImage, BuildOsmosisDockerImage, BuildHermesDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.Profile3Cored, apps.ProfileIBC}
		znetConfig.CoredVersion = "v5.0.0"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestUpgrade, TestIBC, TestModules)
	}
}

func RunIntegrationTestsExport(runUnsafe bool) types.CommandFunc {
	return func(ctx context.Context, deps types.DepsFunc) error {
		deps(CompileModulesSmartContracts, CompileLegacyModulesSmartContracts, CompileAssetExtensionSmartContracts,
			CompileDEXSmartContracts, BuildCoredLocally, BuildCoredDockerImage)

		znetConfig := defaultZNetConfig()
		znetConfig.Profiles = []string{apps.ProfileDevNet}
		znetConfig.CoverageOutputFile = "coverage/coreum-integration-tests-export"

		return runIntegrationTests(ctx, deps, runUnsafe, znetConfig, TestExport)
	}
}

// TestFuzz run fuzz tests in coreum repo.
func TestFuzz(ctx context.Context, deps types.DepsFunc) error {
	deps(CompileAllSmartContracts)

	return golang.TestFuzz(ctx, deps, time.Minute)
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

	// err := znet.ExportCored(ctx, znetConfig)
	// if err != nil {
	// 	return err
	// }

	absRootDir, err := filepath.Abs(znetConfig.RootDir)
	if err != nil {
		return err
	}

	spec := infra.NewSpec(znetConfig)
	config := znet.NewConfig(znetConfig, spec)

	for _, testDir := range testDirs {
		if err := golang.RunTests(ctx, deps, golang.TestConfig{
			PackagePath: filepath.Join(integrationTestsDir, testDir),
			Flags:       flags,
			Envs: []string{
				fmt.Sprintf("CORED_BIN_PATH=%s", filepath.Join(absRootDir, "bin")),
				fmt.Sprintf("ZNET_HOME_DIR=%s", config.HomeDir),
			},
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
		CoredUpgrades: CoredUpgrades(),
	}
}
