package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
}
