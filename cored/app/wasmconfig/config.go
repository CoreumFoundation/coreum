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

// Config is the extra config required for wasm
type Config struct {
	// SimulationGasLimit is the max gas to be used in a smart query contract call
	ContractQueryGasLimit uint64 `mapstructure:"contract-query-gas-limit"`

	// SimulationGasLimit is the max gas to be used in a tx simulation call.
	// When not set the consensus max block gas is used instead
	ContractSimulationGasLimit uint64 `mapstructure:"contract-query-gas-limit"`

	// ContractDebugMode log what contract print
	ContractDebugMode bool `mapstructure:"contract-debug-mode"`

	// MemoryCacheSize in MiB not bytes
	ContractMemoryCacheSize uint32 `mapstructure:"contract-memory-cache-size"`
}

// ToWasmConfig convert config to wasmd's config
func (c Config) ToWasmConfig() wasmtypes.WasmConfig {
	return wasmtypes.WasmConfig{
		SimulationGasLimit: &c.ContractSimulationGasLimit,
		SmartQueryGasLimit: c.ContractQueryGasLimit,
		MemoryCacheSize:    c.ContractMemoryCacheSize,
		ContractDebugMode:  c.ContractDebugMode,
	}
}

// DefaultConfig returns the default settings for WasmConfig
func DefaultConfig() *Config {
	return &Config{
		ContractQueryGasLimit:      DefaultContractQueryGasLimit,
		ContractSimulationGasLimit: DefaultContractSimulationGasLimit,
		ContractDebugMode:          DefaultContractDebugMode,
		ContractMemoryCacheSize:    DefaultContractMemoryCacheSize,
	}
}
