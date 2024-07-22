package keeper

import (
	"context"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewInfiniteAccountKeeper returns new InfiniteAccountKeeper.
func NewInfiniteAccountKeeper(ak ante.AccountKeeper) InfiniteAccountKeeper {
	return InfiniteAccountKeeper{
		ak: ak,
	}
}

// InfiniteAccountKeeper replaces the original gas meter with the infinite one before calling an
// underlying method of real keeper. Gas consumed by the real keeper is non-deterministic. To use
// some ante decorators at the stage where deterministic gas must be delivered we use this wrapper
// to ignore gas consumed by keeper calls required there.
type InfiniteAccountKeeper struct {
	ak ante.AccountKeeper
}

// GetParams returns params.
func (iak InfiniteAccountKeeper) GetParams(ctx context.Context) (params types.Params) {
	return iak.ak.GetParams(sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter()))
}

// GetAccount returns account info by address.
func (iak InfiniteAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return iak.ak.GetAccount(sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter()), addr)
}

// SetAccount sets account info.
func (iak InfiniteAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
	iak.ak.SetAccount(sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter()), acc)
}

// GetModuleAddress returns address of a module.
func (iak InfiniteAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return iak.ak.GetModuleAddress(moduleName)
}

// AddressCodec returns the AddressCodec.
func (iak InfiniteAccountKeeper) AddressCodec() address.Codec {
	return iak.AddressCodec()
}
