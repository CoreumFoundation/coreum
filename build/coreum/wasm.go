package coreum

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/rust"
)

// Smart contract names.
const (
	WasmModulesDir = repoPath + "/integration-tests/contracts/modules"
	WasmIBCDir     = repoPath + "/integration-tests/contracts/ibc"
)

// CompileModulesSmartContracts compiles modules smart contracts.
func CompileModulesSmartContracts(ctx context.Context, deps build.DepsFunc) error {
	return compileWasmDir(WasmModulesDir, deps)
}

// CompileIBCSmartContracts compiles ibc smart contracts.
func CompileIBCSmartContracts(ctx context.Context, deps build.DepsFunc) error {
	return compileWasmDir(WasmIBCDir, deps)
}

// CompileAllSmartContracts compiles all th smart contracts.
func CompileAllSmartContracts(ctx context.Context, deps build.DepsFunc) error {
	if err := compileWasmDir(WasmModulesDir, deps); err != nil {
		return err
	}

	return compileWasmDir(WasmIBCDir, deps)
}

func compileWasmDir(dirPath string, deps build.DepsFunc) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return errors.WithStack(err)
	}

	actions := make([]build.CommandFunc, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		actions = append(actions, rust.CompileSmartContract(filepath.Join(dirPath, e.Name())))
	}
	deps(actions...)

	return nil
}
