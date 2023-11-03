package wasm

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/wasm/types"
)

// IsSmartContract checks if provided address is the address of smart contract.
func IsSmartContract(ctx sdk.Context, addr sdk.AccAddress, wasmKeeper types.WasmKeeper) bool {
	return len(addr) == wasmtypes.ContractAddrLen && wasmKeeper.HasContractInfo(ctx, addr)
}
