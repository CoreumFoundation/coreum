package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var upgradeV1Prefix = []byte("upgradev1")

const (
	upgradePlanIntroducingTokenV1 = "v2"
	upgradeV1Version              = 1
)

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	DelayExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, delay time.Duration) error
}

// StoreDelayedUpgradeV1 stores request for upgrading token to V1.
func (k Keeper) StoreDelayedUpgradeV1(ctx sdk.Context, sender sdk.AccAddress, denom string, ibcEnabled bool) error {
	params := k.GetParams(ctx)
	if ctx.BlockTime().After(params.TokenUpgradeDecisionTimeout) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "it is no longer possible to upgrade the token")
	}

	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if !def.IsIssuer(sender) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only issuer may upgrade the token")
	}

	version := k.GetVersion(ctx, denom)
	if version.Version >= upgradeV1Version {
		return errors.Errorf("denom %s has been already upgraded to v1", denom)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), upgradeV1Prefix)
	key := []byte(denom)
	if store.Has(key) {
		return errors.Errorf("pending request for v1 upgrade already exists for denom: %s", denom)
	}

	delayedData := &types.DelayedTokenUpgradeV1{
		Denom:      denom,
		IbcEnabled: ibcEnabled,
	}
	err = k.delayKeeper.DelayExecution(ctx, "assetft-ibcenable-"+denom, delayedData, params.TokenUpgradeGracePeriod)
	if err != nil {
		return err
	}

	store.Set(key, asset.StoreTrue)

	return nil
}

// UpgradeTokenToV1 upgrades token to version V1.
func (k Keeper) UpgradeTokenToV1(ctx sdk.Context, data *types.DelayedTokenUpgradeV1) error {
	subunit, issuer, err := types.DeconstructDenom(data.Denom)
	if err != nil {
		return err
	}

	if data.IbcEnabled {
		def, err := k.GetDefinition(ctx, data.Denom)
		if err != nil {
			return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", data.Denom)
		}

		def.Features = append(def.Features, types.Feature_ibc)
		k.SetDefinition(ctx, issuer, subunit, def)
	}

	version := k.GetVersion(ctx, data.Denom)
	if version.Version < upgradeV1Version {
		version.Version = upgradeV1Version
		k.SetVersion(ctx, data.Denom, version)
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), upgradeV1Prefix)
	key := []byte(data.Denom)
	store.Delete(key)

	return nil
}
