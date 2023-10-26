//go:build integrationtests

package modules

import (
	_ "embed"
)

// Smart contracts bytecode.
var (
	//go:embed bank-send/artifacts/bank_send.wasm
	BankSendWASM []byte
	//go:embed simple-state/artifacts/simple_state.wasm
	SimpleStateWASM []byte
	//go:embed ft/artifacts/ft.wasm
	FTWASM []byte
	//go:embed nft/artifacts/nft.wasm
	NftWASM []byte
	//go:embed authz/artifacts/authz.wasm
	AuthzWASM []byte
)
