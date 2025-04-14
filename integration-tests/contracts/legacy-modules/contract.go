//go:build integrationtests

package modules

import (
	_ "embed"
)

// Smart contracts bytecode.
var (
	// TODO(v6): remove contract once we upgrade to v6
	//go:embed asset-extension-legacy/artifacts/asset_extension_legacy.wasm
	AssetFTExtensionLegacyWASM []byte
	//go:embed ft-legacy/artifacts/ft_legacy.wasm
	FTLegacyWASM []byte
	//go:embed nft-legacy/artifacts/nft_legacy.wasm
	NFTLegacyWASM []byte
)
