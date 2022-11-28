package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func (k Keeper) SetGlobalFreezeEnabled(ctx sdk.Context, sender sdk.AccAddress, denom string, enabled bool) error {
	// FIXME: Fungible token is read from store twice. Refactor checkFeatureAllowed
	// checkFeatureAllowed should be moved to FungibleTokenDefinition or at lest func should receive it.
	// same for available balance.
	if err := k.checkFeatureAllowed(ctx, sender, denom, types.FungibleTokenFeature_freeze); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	// Global freeze implemented in an idempotent way, so it is allowed to freeze/unfreeze multiple times without effect.
	if enabled {
		store.Set(types.CreateGlobalFreezePrefix(denom), []byte{0x01})
	} else {
		store.Delete(types.CreateGlobalFreezePrefix(denom))
	}

	return nil
}
