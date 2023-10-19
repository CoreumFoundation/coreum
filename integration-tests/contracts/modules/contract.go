//go:build integrationtests

package modules

import (
	_ "embed"
	"encoding/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
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

// EmptyPayload represents empty payload.
var EmptyPayload = must.Bytes(json.Marshal(struct{}{}))
