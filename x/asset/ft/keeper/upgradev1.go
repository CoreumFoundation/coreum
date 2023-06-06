package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

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

	if err := k.SetPendingVersion(ctx, denom, upgradeV1Version); err != nil {
		return err
	}

	if !ibcEnabled {
		// if issuer does not want to enable IBC we may upgrade the token immediately
		// because it's behaviour is not changed
		version.Version = upgradeV1Version
		k.SetVersion(ctx, denom, version)
		k.ClearPendingVersion(ctx, denom)
		return nil
	}

	data := &types.DelayedTokenUpgradeV1{
		Denom: denom,
	}

	err = k.delayKeeper.DelayExecution(ctx, tokenUpgradeID(upgradeV1Version, data.Denom), data, params.TokenUpgradeGracePeriod)
	if err != nil {
		return err
	}
	return nil
}

// UpgradeTokenToV1 upgrades token to version V1.
func (k Keeper) UpgradeTokenToV1(ctx sdk.Context, data *types.DelayedTokenUpgradeV1) error {
	def, err := k.GetDefinition(ctx, data.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", data.Denom)
	}

	subunit, issuer, err := types.DeconstructDenom(data.Denom)
	if err != nil {
		return err
	}

	def.Features = append(def.Features, types.Feature_ibc)
	k.SetDefinition(ctx, issuer, subunit, def)

	version := k.GetVersion(ctx, data.Denom)
	version.Version = upgradeV1Version
	k.SetVersion(ctx, data.Denom, version)
	k.ClearPendingVersion(ctx, data.Denom)

	return nil
}

func tokenUpgradeID(version int, denom string) string {
	return fmt.Sprintf("%s-upgrade-%d-%s", types.ModuleName, version, denom)
}
