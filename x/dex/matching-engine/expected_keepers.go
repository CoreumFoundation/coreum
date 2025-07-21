package matchingengine

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// OrderBookQueue is an interface which returns the next order with the best price.
type OrderBookQueue interface {
	Next() (types.OrderBookRecord, bool, error)
}

// DEXKeeper exposes methods of dex module needed by the matching engine.
type DEXKeeper interface {
	GetOrderData(ctx sdk.Context, orderSequence uint64) (types.OrderData, error)
}

// AccountKeeper is the interface to the auth module.
type AccountKeeper interface {
	GetAccountAddress(ctx sdk.Context, accNumber uint64) (sdk.AccAddress, error)
}
