package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var globalFreezeEnabledStoreVal = []byte{0x00}

// GloballyFreeze enables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyFreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	ft, err := k.GetTokenDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	k.SetGlobalFreeze(ctx, denom, true)
	return nil
}

// GloballyUnfreeze disables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyUnfreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	ft, err := k.GetTokenDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	k.SetGlobalFreeze(ctx, denom, false)
	return nil
}

// SetGlobalFreeze enables/disables global freeze on a fungible token depending on frozen arg.
func (k Keeper) SetGlobalFreeze(ctx sdk.Context, denom string, frozen bool) {
	if frozen {
		ctx.KVStore(k.storeKey).Set(types.CreateGlobalFreezeKey(denom), globalFreezeEnabledStoreVal)
		return
	}
	ctx.KVStore(k.storeKey).Delete(types.CreateGlobalFreezeKey(denom))
}

func (k Keeper) isGloballyFrozen(ctx sdk.Context, denom string) bool {
	globFreezeVal := ctx.KVStore(k.storeKey).Get(types.CreateGlobalFreezeKey(denom))
	return bytes.Equal(globFreezeVal, globalFreezeEnabledStoreVal)
}
