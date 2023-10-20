package wasm

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper defines methods required from the WASM keeper.
type Keeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}

// IsSmartContract checks if provided address is the address of smart contract.
func IsSmartContract(ctx sdk.Context, addr sdk.AccAddress, wasmKeeper Keeper) bool {
	return len(addr) == wasmtypes.ContractAddrLen && wasmKeeper.HasContractInfo(ctx, addr)
}
