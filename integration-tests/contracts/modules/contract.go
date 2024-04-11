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
	//go:embed authz-transfer/artifacts/authz_transfer.wasm
	AuthzTransferWASM []byte
	//go:embed authz-nft-trade/artifacts/authz_nft_trade.wasm
	AuthzNftTradeWASM []byte
	//go:embed authz-stargate/artifacts/authz_stargate.wasm
	AuthzStargateWASM []byte
	//go:embed asset-extension/artifacts/asset_extension.wasm
	AssetExtensionWasm []byte
)
