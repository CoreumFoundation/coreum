package testcontracts

import (
	_ "embed"
)

// Built artifacts of dex.
var (
	//go:embed dex-reentrancy-poc/artifacts/dex_reentrancy_poc.wasm
	DexReentrancyPocWasm []byte
)
