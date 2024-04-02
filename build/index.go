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
	"build/integration-tests": {Fn: coreum.BuildAllIntegrationTests,
		Description: "Builds all the integration tests"},
	"build/integration-tests/ibc": {Fn: coreum.BuildIntegrationTests(coreum.TestIBC),
		Description: "Builds IBC integration tests"},
	"build/integration-tests/modules": {Fn: coreum.BuildIntegrationTests(coreum.TestModules),
		Description: "Builds modules integration tests"},
	"build/integration-tests/upgrade": {Fn: coreum.BuildIntegrationTests(coreum.TestUpgrade),
		Description: "Builds upgrade integration tests"},
	"download":       {Fn: coreum.DownloadDependencies, Description: "Downloads go dependencies"},
	"generate":       {Fn: coreum.Generate, Description: "Generates artifacts"},
	"images":         {Fn: coreum.BuildCoredDockerImage, Description: "Builds cored docker image"},
	"lint":           {Fn: coreum.Lint, Description: "Lints code"},
	"release":        {Fn: coreum.ReleaseCored, Description: "Releases cored binary"},
	"release/images": {Fn: coreum.ReleaseCoredImage, Description: "Releases cored docker images"},
	"test":           {Fn: coreum.Test, Description: "Runs unit tests"},
	"tidy":           {Fn: coreum.Tidy, Description: "Runs go mod tidy"},
	"wasm": {Fn: coreum.CompileAllSmartContracts,
		Description: "Builds smart contracts required by integration tests"},
}
