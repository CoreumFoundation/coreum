package build

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum/build/coreum"
	"github.com/CoreumFoundation/crust/build/crust"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]build.Command{
	"build/me": {Fn: crust.BuildBuilder, Description: "Builds the builder"},
	"build":    {Fn: coreum.BuildCored, Description: "Builds cored binary"},
	"download": {Fn: coreum.DownloadDependencies, Description: "Downloads go dependencies"},
	"generate": {Fn: coreum.Generate, Description: "Generates artifacts"},
	"images":   {Fn: coreum.BuildCoredDockerImage, Description: "Builds cored docker image"},
	"integration-tests": {
		Fn:          coreum.RunAllIntegrationTests(false),
		Description: "Runs all safe integration tests",
	},
	"integration-tests-unsafe": {
		Fn:          coreum.RunAllIntegrationTests(true),
		Description: "Runs all the integration tests including unsafe",
	},
	"integration-tests/ibc": {
		Fn:          coreum.RunIntegrationTests(coreum.TestIBC, false),
		Description: "Runs safe IBC integration tests",
	},
	"integration-tests-unsafe/ibc": {
		Fn:          coreum.RunIntegrationTests(coreum.TestIBC, true),
		Description: "Runs all IBC integration tests including unsafe",
	},
	"integration-tests/modules": {
		Fn:          coreum.RunIntegrationTests(coreum.TestModules, false),
		Description: "Runs safe modules integration tests",
	},
	"integration-tests-unsafe/modules": {
		Fn:          coreum.RunIntegrationTests(coreum.TestModules, true),
		Description: "Runs all modules integration tests including unsafe",
	},
	"integration-tests/upgrade": {
		Fn:          coreum.RunIntegrationTests(coreum.TestUpgrade, false),
		Description: "Runs safe upgrade integration tests",
	},
	"integration-tests-unsafe/upgrade": {
		Fn:          coreum.RunIntegrationTests(coreum.TestUpgrade, true),
		Description: "Runs all upgrade integration tests including unsafe",
	},
	"lint":           {Fn: coreum.Lint, Description: "Lints code"},
	"release":        {Fn: coreum.ReleaseCored, Description: "Releases cored binary"},
	"release/images": {Fn: coreum.ReleaseCoredImage, Description: "Releases cored docker images"},
	"test":           {Fn: coreum.Test, Description: "Runs unit tests"},
	"tidy":           {Fn: coreum.Tidy, Description: "Runs go mod tidy"},
	"wasm": {Fn: coreum.CompileAllSmartContracts,
		Description: "Builds smart contracts required by integration tests"},
}
