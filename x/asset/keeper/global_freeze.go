package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

var (
	globalFreezeEnabledStoreVal = []byte{0x00}
)

// GloballyFreezeFungibleToken enables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyFreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.FungibleTokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Set(types.CreateGlobalFreezePrefix(denom), globalFreezeEnabledStoreVal)
	return nil
}

// GloballyUnfreezeFungibleToken disables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyUnfreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.FungibleTokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	ctx.KVStore(k.storeKey).Delete(types.CreateGlobalFreezePrefix(denom))
	return nil
}

func (k Keeper) isGloballyFrozen(ctx sdk.Context, denom string) bool {
	globFreezeVal := ctx.KVStore(k.storeKey).Get(types.CreateGlobalFreezePrefix(denom))
	return bytes.Equal(globFreezeVal, globalFreezeEnabledStoreVal)
}
