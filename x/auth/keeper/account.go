package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func NewInfiniteAccountKeeper(ak ante.AccountKeeper) InfiniteAccountKeeper {
	return InfiniteAccountKeeper{
		ak: ak,
	}
}

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
