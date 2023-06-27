//go:build integrationtests

package modules

import (
	_ "embed"
)

var (
	//nolint
	//go:embed testdata/wasm/bank-send/artifacts/bank_send.wasm
	BankSendWASM []byte
	//nolint
	//go:embed testdata/wasm/simple-state/artifacts/simple_state.wasm
	SimpleStateWASM []byte
	//nolint
	//go:embed testdata/wasm/ft/artifacts/ft.wasm
	FTWASM []byte
	//nolint
	//go:embed testdata/wasm/nft/artifacts/nft.wasm
	NftWASM []byte
	//nolint
	//go:embed testdata/wasm/authz/artifacts/authz.wasm
	AuthzWASM []byte
)
