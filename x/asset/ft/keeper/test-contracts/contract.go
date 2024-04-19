package testcontracts

import (
	_ "embed"
)

var (
	//go:embed asset-extension/artifacts/asset_extension.wasm
	AssetExtensionWasm []byte
)
