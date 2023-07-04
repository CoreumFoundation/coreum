//go:build integrationtests

package ibc

import (
	_ "embed"
)

// Smart contracts bytecode.
var (
	//go:embed ibc-transfer/artifacts/ibc_transfer.wasm
	IBCTransferWASM []byte
	//go:embed ibc-call/artifacts/ibc_call.wasm
	IBCClassWASM []byte
)
