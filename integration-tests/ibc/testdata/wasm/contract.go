//go:build integrationtests

package modules

import (
	_ "embed"
)

var (
	//nolint
	//go:embed ibc-transfer/artifacts/ibc_transfer.wasm
	IbcTransferWASM []byte
	//nolint
	//go:embed ibc-call/artifacts/ibc_call.wasm
	IbcClassWASM []byte
)
