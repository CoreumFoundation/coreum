package build

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/coreum/build/coreum"
	"github.com/CoreumFoundation/crust/build/crust"
)

// Commands is a definition of commands available in build system.
var Commands = map[string]build.CommandFunc{
	"build/me":                        crust.BuildBuilder,
	"build":                           coreum.BuildCored,
	"build/integration-tests":         coreum.BuildAllIntegrationTests,
	"build/integration-tests/ibc":     coreum.BuildIntegrationTests(coreum.TestIBC),
	"build/integration-tests/modules": coreum.BuildIntegrationTests(coreum.TestModules),
	"build/integration-tests/upgrade": coreum.BuildIntegrationTests(coreum.TestUpgrade),
	"generate":                        coreum.Generate,
	"images":                          coreum.BuildCoredDockerImage,
	"lint":                            coreum.Lint,
	"release":                         coreum.ReleaseCored,
	"release/images":                  coreum.ReleaseCoredImage,
	"test":                            coreum.Test,
	"tidy":                            coreum.Tidy,
	"wasm":                            coreum.CompileAllSmartContracts,
}
