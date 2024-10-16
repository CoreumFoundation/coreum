package types

import (
	context "context"
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
	DEXIncreaseLimits(ctx sdk.Context, addr sdk.AccAddress, lockCoin, reserveWhitelistingCoin sdk.Coin) error
	DEXDecreaseLimits(ctx sdk.Context, addr sdk.AccAddress, unlockCoin, releaseWhitelistingCoin sdk.Coin) error
	DEXDecreaseLimitsAndSend(
		ctx sdk.Context,
		fromAddr, toAddr sdk.AccAddress,
		unlockAndSendCoin, releaseWhitelistingCoin sdk.Coin,
	) error
	DEXChecksLimitsAndSend(
		ctx sdk.Context,
		fromAddr, toAddr sdk.AccAddress,
		sendCoin, checkReserveWhitelistingCoin sdk.Coin,
	) error
	DEXLock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	DEXUnlock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	GetSpendableBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetDEXSettings(ctx sdk.Context, denom string) (dextypes.DEXSettings, error)
	ValidateDEXCancelOrdersByDenomIsAllowed(ctx sdk.Context, addr sdk.AccAddress, denom string) error
}

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	ExecuteAfterBlock(ctx sdk.Context, id string, data proto.Message, height uint64) error
	ExecuteAfter(ctx sdk.Context, id string, data proto.Message, time time.Time) error
	RemoveExecuteAtBlock(ctx sdk.Context, id string, height uint64) error
	RemoveExecuteAfter(ctx sdk.Context, id string, time time.Time) error
}
