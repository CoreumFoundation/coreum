package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

const tokenUpgradeV1Version = 1

// AddDelayedTokenUpgradeV1 stores request for upgrading token to V1.
func (k Keeper) AddDelayedTokenUpgradeV1(ctx sdk.Context, sender sdk.AccAddress, denom string, ibcEnabled bool) error {
	params := k.GetParams(ctx)
	if ctx.BlockTime().After(params.TokenUpgradeDecisionTimeout) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "it is no longer possible to upgrade the token")
	}

	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if !def.IsIssuer(sender) {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only issuer may upgrade the token")
	}

	if def.Version >= tokenUpgradeV1Version {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "denom %s has been already upgraded to v1", denom)
	}

	if err := k.SetPendingVersion(ctx, denom, tokenUpgradeV1Version); err != nil {
		return err
	}

	// we don't read the current TokenUpgradeStatuses because we know that this is the initial state
	tokenUpgradeStatuses := types.TokenUpgradeStatuses{
		V1: &types.TokenUpgradeV1Status{
			IbcEnabled: ibcEnabled,
			StartTime:  ctx.BlockTime(),
			EndTime:    ctx.BlockTime().Add(params.TokenUpgradeGracePeriod),
		},
	}

	if !ibcEnabled {
		// if issuer does not want to enable IBC we may upgrade the token immediately
		// because it's behaviour is not changed
		def.Version = tokenUpgradeV1Version
		subunit, issuer, err := types.DeconstructDenom(denom)
		if err != nil {
			return err
		}
		k.SetDefinition(ctx, issuer, subunit, def)
		k.ClearPendingVersion(ctx, denom)
		tokenUpgradeStatuses.V1.EndTime = tokenUpgradeStatuses.V1.StartTime
		k.SetTokenUpgradeStatuses(ctx, denom, tokenUpgradeStatuses)
		return nil
	}

	k.SetTokenUpgradeStatuses(ctx, denom, tokenUpgradeStatuses)

	data := &types.DelayedTokenUpgradeV1{
		Denom: denom,
	}

	return k.delayKeeper.DelayExecution(ctx, tokenUpgradeID(tokenUpgradeV1Version, data.Denom), data, params.TokenUpgradeGracePeriod)
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
	def.Version = tokenUpgradeV1Version
	k.SetDefinition(ctx, issuer, subunit, def)
	k.ClearPendingVersion(ctx, data.Denom)

	return nil
}

func tokenUpgradeID(version int, denom string) string {
	return fmt.Sprintf("%s-upgrade-%d-%s", types.ModuleName, version, denom)
}
