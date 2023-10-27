package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Keeper defines methods required from the WASM keeper.
type WasmKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
}
