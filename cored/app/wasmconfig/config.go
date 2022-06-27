package wasmconfig

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// config default values
const (
	DefaultContractQueryGasLimit      = uint64(3_000_000)
	DefaultContractSimulationGasLimit = uint64(50_000_000)
	DefaultContractDebugMode          = false
	DefaultContractMemoryCacheSize    = uint32(2048)
)

func DefaultWASMConfig() wasmtypes.WasmConfig {
	defaultContractSimulationGasLimit := DefaultContractSimulationGasLimit

	return wasmtypes.WasmConfig{
		SimulationGasLimit: &defaultContractSimulationGasLimit,
		SmartQueryGasLimit: DefaultContractQueryGasLimit,
		MemoryCacheSize:    DefaultContractMemoryCacheSize,
		ContractDebugMode:  DefaultContractDebugMode,
	}
}
