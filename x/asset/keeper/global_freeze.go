package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func (k Keeper) SetGlobalFreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	if err := k.checkFeatureAllowed(ctx, sender, denom, types.FungibleTokenFeature_freeze); err != nil {
		return err
	}

	globalFrozenStore := k.globalFrozenBalancesStore(ctx)
	globalFrozenStore.Set([]byte(denom), []byte("true"))

	return nil
}

// globalFrozenBalancesStore get the store for the frozen balances of all accounts
func (k Keeper) globalFrozenBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.GlobalFrozenBalancesKeyPrefix)
}
