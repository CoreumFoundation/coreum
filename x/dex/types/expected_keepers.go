package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"

	dextypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
)

// AccountKeeper defines the expected account keeper interface.
type AccountKeeper interface {
	GetAccount(ctx context.Context, address sdk.AccAddress) sdk.AccountI
}

// AccountQueryServer defines the expected account query server interface.
type AccountQueryServer interface {
	AccountAddressByID(
		ctx context.Context, req *authtypes.QueryAccountAddressByIDRequest,
	) (*authtypes.QueryAccountAddressByIDResponse, error)
}

// AssetFTKeeper represents required methods of asset ft keeper.
type AssetFTKeeper interface {
	DEXExecuteActions(ctx sdk.Context, actions dextypes.DEXActions) error
	DEXDecreaseLimits(ctx sdk.Context, addr sdk.AccAddress, lockedCoin sdk.Coins, expectedToReceiveCoin sdk.Coin) error
	GetSpendableBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error)
	GetDEXSettings(ctx sdk.Context, denom string) (dextypes.DEXSettings, error)
	ValidateDEXCancelOrdersByDenomIsAllowed(ctx sdk.Context, addr sdk.AccAddress, denom string) error
	HasSupply(ctx sdk.Context, denom string) bool
}

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	ExecuteAfterBlock(ctx sdk.Context, id string, data proto.Message, height uint64) error
	ExecuteAfter(ctx sdk.Context, id string, data proto.Message, time time.Time) error
	RemoveExecuteAtBlock(ctx sdk.Context, id string, height uint64) error
	RemoveExecuteAfter(ctx sdk.Context, id string, time time.Time) error
}
