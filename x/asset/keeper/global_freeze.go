package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// SetGlobalFreezeEnabled enables or disables global freeze on a fungible token depending on enabled arg.
func (k Keeper) SetGlobalFreezeEnabled(ctx sdk.Context, sender sdk.AccAddress, denom string, enabled bool) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.FungibleTokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
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
