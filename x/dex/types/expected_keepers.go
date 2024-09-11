package types

import (
	context "context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
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
	DEXLock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	DEXUnlock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	DEXUnlockAndSend(ctx sdk.Context, from, to sdk.AccAddress, coin sdk.Coin) error
	GetSpendableBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDEXSettings(ctx sdk.Context, denom string) (types.DEXSettings, error)
}

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	ExecuteAfterBlock(ctx sdk.Context, id string, data proto.Message, height uint64) error
	ExecuteAfter(ctx sdk.Context, id string, data proto.Message, time time.Time) error
	RemoveExecuteAtBlock(ctx sdk.Context, id string, height uint64) error
	RemoveExecuteAfter(ctx sdk.Context, id string, time time.Time) error
}
