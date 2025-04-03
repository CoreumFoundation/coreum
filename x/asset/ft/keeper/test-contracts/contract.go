package testcontracts

import (
	_ "embed"
)

// Built artifacts of smart contracts.
var (
	//go:embed asset-extension/artifacts/asset_extension.wasm
	AssetExtensionWasm []byte
)

// Check contract.rs for constant values defined in contract.
const (
	AmountDisallowedTrigger               = 7
	AmountBurningTrigger                  = 101
	AmountMintingTrigger                  = 105
	AmountIgnoreBurnRateTrigger           = 108
	AmountIgnoreSendCommissionRateTrigger = 109
	AmountBlockIBCTrigger                 = 110
	AmountBlockSmartContractTrigger       = 111
	IDDEXOrderSuffixTrigger               = "blocked"
	AmountDEXExpectToSpendTrigger         = 103_000_000
	AmountDEXExpectToReceiveTrigger       = 104_000_000
)
