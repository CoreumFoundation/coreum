package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewInfiniteAccountKeeper returns new InfiniteAccountKeeper
func NewInfiniteAccountKeeper(ak ante.AccountKeeper) InfiniteAccountKeeper {
	return InfiniteAccountKeeper{
		ak: ak,
	}
}

// InfiniteAccountKeeper replaces the original gas meter with the infinite one before calling an underlying method of real keeper.
// Gas consumed by the real keeper is non-deterministic. To use some ante decorators at the stage where deterministic gas must be
// delivered we use this wrapper to ignore gas consumed by keeper calls required there.
type InfiniteAccountKeeper struct {
	ak ante.AccountKeeper
}

func (iak InfiniteAccountKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	return iak.ak.GetParams(ctx)
}

func (iak InfiniteAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI {
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	return iak.ak.GetAccount(ctx, addr)
}

func (iak InfiniteAccountKeeper) SetAccount(ctx sdk.Context, acc types.AccountI) {
	ctx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	iak.ak.SetAccount(ctx, acc)
}

func (iak InfiniteAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return iak.ak.GetModuleAddress(moduleName)
}
