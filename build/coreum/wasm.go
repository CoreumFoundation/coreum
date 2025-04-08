package coreum

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/crust/build/rust"
	"github.com/CoreumFoundation/crust/build/types"
)

// Smart contract names.
const (
	WasmModulesDir        = repoPath + "/integration-tests/contracts/modules"
	WasmIBCDir            = repoPath + "/integration-tests/contracts/ibc"
	WasmAssetExtensionDir = repoPath + "/x/asset/ft/keeper/test-contracts"
	WasmDexDir            = repoPath + "/x/dex/keeper/test-contracts"
)

// CompileModulesSmartContracts compiles modules smart contracts.
func CompileModulesSmartContracts(ctx context.Context, deps types.DepsFunc) error {
	return compileWasmDir(WasmModulesDir, deps)
}

// CompileIBCSmartContracts compiles ibc smart contracts.
func CompileIBCSmartContracts(ctx context.Context, deps types.DepsFunc) error {
	return compileWasmDir(WasmIBCDir, deps)
}

// CompileAssetExtensionSmartContracts compiles asset smart contracts.
func CompileAssetExtensionSmartContracts(ctx context.Context, deps types.DepsFunc) error {
	return compileWasmDir(WasmAssetExtensionDir, deps)
}

// CompileDEXSmartContracts compiles asset smart contracts.
func CompileDEXSmartContracts(ctx context.Context, deps types.DepsFunc) error {
	return compileWasmDir(WasmDexDir, deps)
}

// CompileAllSmartContracts compiles all th smart contracts.
func CompileAllSmartContracts(ctx context.Context, deps types.DepsFunc) error {
	allWasmDirectories := []string{
		WasmModulesDir,
		WasmIBCDir,
		WasmAssetExtensionDir,
		WasmDexDir,
	}
	for _, dir := range allWasmDirectories {
		if err := compileWasmDir(dir, deps); err != nil {
			return err
		}
	}
	return nil
}

func compileWasmDir(dirPath string, deps types.DepsFunc) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return errors.WithStack(err)
	}

	actions := make([]types.CommandFunc, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		actions = append(actions, rust.CompileSmartContract(filepath.Join(dirPath, e.Name())))
	}
	deps(actions...)

	return nil
}
