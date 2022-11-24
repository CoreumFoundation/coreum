package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerr "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func (k Keeper) SetGlobalFreezeEnabled(ctx sdk.Context, sender sdk.AccAddress, denom string, enabled bool) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return err
	}

	// FIXME: Fungible token is read from store twice.
	// checkFeatureAllowed should be moved to FungibleTokenDefinition or at lest func should receive it.
	// same for available balance.
	if err := k.checkFeatureAllowed(ctx, sender, denom, types.FungibleTokenFeature_freeze); err != nil {
		return err
	}

	if ft.GlobalFreezeEnabled == enabled {
		if enabled {
			return sdkerr.Wrap(types.ErrGlobalFreezeEnabled, "already enabled")
		} else {
			return sdkerr.Wrap(types.ErrGlobalFreezeDisabled, "already disabled")
		}
	}

	ft.GlobalFreezeEnabled = enabled
	k.SetFungibleTokenDefinition(ctx, ft)

	return nil
}
