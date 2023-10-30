package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// WasmKeeper defines methods required from the WASM keeper.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}
