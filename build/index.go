package build

import (
	"context"

	"github.com/CoreumFoundation/coreum/build/coreum"
	"github.com/CoreumFoundation/crust/build/crust"
	"github.com/CoreumFoundation/crust/build/golang"
	"github.com/CoreumFoundation/crust/build/tools"
	"github.com/CoreumFoundation/crust/build/types"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]types.Command{
	"build/me":   {Fn: coreum.BuildBuilder, Description: "Builds the builder"},
	"build/znet": {Fn: crust.BuildZNet, Description: "Builds znet binary"},
	"build": {Fn: func(ctx context.Context, deps types.DepsFunc) error {
		deps(
			coreum.BuildCored,
			coreum.BuildExtendedCoredInDocker,
		)
		return nil
	}, Description: "Builds cored binaries"},
	"build/cored":     {Fn: coreum.BuildCored, Description: "Builds cored binary"},
	"build/cored-ext": {Fn: coreum.BuildExtendedCoredInDocker, Description: "Builds extended cored binary"},
	"generate":        {Fn: coreum.Generate, Description: "Generates artifacts"},
	"setup":           {Fn: tools.InstallAll, Description: "Installs all the required tools"},
	"images": {Fn: func(ctx context.Context, deps types.DepsFunc) error {
		deps(
			coreum.BuildCoredDockerImage,
			coreum.BuildExtendedCoredDockerImage,
			coreum.BuildGaiaDockerImage,
			coreum.BuildHermesDockerImage,
			coreum.BuildOsmosisDockerImage,
		)
		return nil
	}, Description: "Builds cored docker images"},
	"images/cored":     {Fn: coreum.BuildCoredDockerImage, Description: "Builds cored docker image"},
	"images/cored-ext": {Fn: coreum.BuildExtendedCoredDockerImage, Description: "Builds extended cored docker image"},
	"images/gaiad":     {Fn: coreum.BuildGaiaDockerImage, Description: "Builds gaia docker image"},
	"images/hermes":    {Fn: coreum.BuildHermesDockerImage, Description: "Builds hermes docker image"},
	"images/osmosis":   {Fn: coreum.BuildOsmosisDockerImage, Description: "Builds osmosis docker image"},
	"integration-tests": {
		Fn:          coreum.RunAllIntegrationTests(false),
		Description: "Runs all safe integration tests",
	},
	"integration-tests-unsafe": {
		Fn:          coreum.RunAllIntegrationTests(true),
		Description: "Runs all the integration tests including unsafe",
	},
	"integration-tests/ibc": {
		Fn:          coreum.RunIntegrationTestsIBC(false),
		Description: "Runs safe IBC integration tests",
	},
	"integration-tests-unsafe/ibc": {
		Fn:          coreum.RunIntegrationTestsIBC(true),
		Description: "Runs all IBC integration tests including unsafe",
	},
	"integration-tests/modules": {
		Fn:          coreum.RunIntegrationTestsModules(false),
		Description: "Runs safe modules integration tests",
	},
	"integration-tests-unsafe/modules": {
		Fn:          coreum.RunIntegrationTestsModules(true),
		Description: "Runs all modules integration tests including unsafe",
	},
	"integration-tests-unsafe/stress": {
		Fn:          coreum.RunIntegrationTestsStress(true),
		Description: "Runs unsafe stress integration tests",
	},
	"integration-tests/upgrade": {
		Fn:          coreum.RunIntegrationTestsUpgrade(false),
		Description: "Runs safe upgrade integration tests",
	},
	"integration-tests-unsafe/upgrade": {
		Fn:          coreum.RunIntegrationTestsUpgrade(true),
		Description: "Runs all upgrade integration tests including unsafe",
	},
	"lint":           {Fn: coreum.Lint, Description: "Lints code"},
	"release":        {Fn: coreum.ReleaseCored, Description: "Releases cored binary"},
	"release/images": {Fn: coreum.ReleaseCoredImage, Description: "Releases cored docker images"},
	"test":           {Fn: coreum.Test, Description: "Runs unit tests"},
	"test-fuzz":      {Fn: coreum.TestFuzz, Description: "Runs fuzz tests"},
	"tidy":           {Fn: golang.Tidy, Description: "Runs go mod tidy"},
	"wasm": {
		Fn:          coreum.CompileAllSmartContracts,
		Description: "Builds smart contracts required by integration tests",
	},
}
