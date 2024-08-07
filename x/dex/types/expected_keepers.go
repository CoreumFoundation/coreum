package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper interface.
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, address sdk.AccAddress) authtypes.AccountI
	GetAccountAddressByID(ctx sdk.Context, id uint64) string
}

// AssetFTKeeper represents required methods of asset ft keeper.
type AssetFTKeeper interface {
	DEXLock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	DEXUnlock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error
	DEXUnlockAndSend(ctx sdk.Context, from, to sdk.AccAddress, coin sdk.Coin) error
}
