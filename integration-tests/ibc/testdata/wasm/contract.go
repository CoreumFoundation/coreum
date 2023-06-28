//go:build integrationtests

package ibc

import (
	_ "embed"
)

var (
	//nolint
	//go:embed ibc-transfer/artifacts/ibc_transfer.wasm
	IBCTransferWASM []byte
	//nolint
	//go:embed ibc-call/artifacts/ibc_call.wasm
	IBCClassWASM []byte
)
