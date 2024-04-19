package testcontracts

import (
	_ "embed"
)

// Built artifacts of smart contracts.
var (
	//go:embed asset-extension/artifacts/asset_extension.wasm
	AssetExtensionWasm []byte
)
