//go:build integrationtests

package modules

import (
	_ "embed"
)

var (
	//nolint
	//go:embed bank-send/artifacts/bank_send.wasm
	BankSendWASM []byte
	//nolint
	//go:embed simple-state/artifacts/simple_state.wasm
	SimpleStateWASM []byte
	//nolint
	//go:embed ft/artifacts/ft.wasm
	FTWASM []byte
	//nolint
	//go:embed nft/artifacts/nft.wasm
	NftWASM []byte
	//nolint
	//go:embed authz/artifacts/authz.wasm
	AuthzWASM []byte
)
