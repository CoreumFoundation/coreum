package modules

import (
	_ "embed"
)

var (
	//go:embed testdata/wasm/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
	//go:embed testdata/wasm/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
	//go:embed testdata/wasm/nft/artifacts/nft.wasm
	nftWASM []byte
	//go:embed testdata/wasm/authz/artifacts/authz.wasm
	authzWASM []byte
	//nolint
	//go:embed testdata/wasm/ft/artifacts/ft.wasm
	FTWASM []byte
)
