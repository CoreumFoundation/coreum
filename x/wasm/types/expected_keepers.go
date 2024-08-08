package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmKeeper defines methods required from the WASM keeper.
type WasmKeeper interface {
	HasContractInfo(ctx context.Context, contractAddress sdk.AccAddress) bool
}
