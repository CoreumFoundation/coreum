package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var ibcEnablePrefix = []byte("upgradev1")

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	DelayMessage(ctx sdk.Context, id string, msg proto.Message, delay time.Duration) error
}

// StoreDelayedUpgradeV1 stores request for upgrading token to V1.
func (k Keeper) StoreDelayedUpgradeV1(ctx sdk.Context, sender sdk.AccAddress, denom string, ibcEnabled bool) error {
	params := k.GetParams(ctx)
	if ctx.BlockTime().After(params.TokenUpgradeDecisionTimeout) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "it is no longer possible IBC")
	}

	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if !def.IsIssuer(sender) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only issuer may enable IBC")
	}

	if def.IsFeatureEnabled(types.Feature_ibc) {
		return errors.Errorf("ibc has been already enabled for denom: %s", denom)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), ibcEnablePrefix)
	key := []byte(denom)
	if store.Has(key) {
		return errors.Errorf("pending request for enabling IBC already exists for denom: %s", denom)
	}

	delayedData := &types.DelayedTokenUpgradeV1{
		Denom:      denom,
		IbcEnabled: ibcEnabled,
	}
	err = k.delayKeeper.DelayMessage(ctx, "assetft-ibcenable-"+denom, delayedData, params.TokenUpgradeGracePeriod)
	if err != nil {
		return err
	}

	store.Set(key, asset.StoreTrue)

	return nil
}

// UpgradeTokenToV1 upgrades token to version V1.
func (k Keeper) UpgradeTokenToV1(ctx sdk.Context, data *types.DelayedTokenUpgradeV1) error {
	def, err := k.GetDefinition(ctx, data.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", data.Denom)
	}

	if def.IsFeatureEnabled(types.Feature_ibc) {
		return errors.Errorf("ibc has been already enabled for denom: %s", data.Denom)
	}

	subunit, issuer, err := types.DeconstructDenom(data.Denom)
	if err != nil {
		return err
	}

	def.Features = append(def.Features, types.Feature_ibc)
	k.SetDefinition(ctx, issuer, subunit, def)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), ibcEnablePrefix)
	key := []byte(data.Denom)
	store.Delete(key)

	return nil
}
