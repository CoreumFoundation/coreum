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
	//go:embed ft-legacy/artifacts/ft_legacy.wasm
	FTLegacyWASM []byte
	//go:embed nft-legacy/artifacts/nft_legacy.wasm
	NftLegacyWASM []byte
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
)
