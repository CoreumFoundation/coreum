package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

var ibcEnablePrefix = []byte("upgradev3ibcenable")

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	DelayMessage(ctx sdk.Context, id string, msg proto.Message, delay time.Duration) error
}

// EnableIBCKeeper provides functionality required by v3 upgrade.
type EnableIBCKeeper struct {
	cdc         codec.BinaryCodec
	keeper      Keeper
	storeKey    sdk.StoreKey
	delayKeeper DelayKeeper
}

// NewEnableIBCKeeper returns EnableIBCKeeper keeper.
func NewEnableIBCKeeper(
	cdc codec.BinaryCodec,
	keeper Keeper,
	storeKey sdk.StoreKey,
	delayKeeper DelayKeeper,
) EnableIBCKeeper {
	return EnableIBCKeeper{
		cdc:         cdc,
		keeper:      keeper,
		storeKey:    storeKey,
		delayKeeper: delayKeeper,
	}
}

// StoreEnableIBCRequest stores request for enabling IBC.
func (k EnableIBCKeeper) StoreEnableIBCRequest(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	params := k.keeper.GetParams(ctx)
	if ctx.BlockTime().After(params.IbcDecisionTimeout) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "it is no longer possible IBC")
	}

	def, err := k.keeper.GetDefinition(ctx, denom)
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

	err = k.delayKeeper.DelayMessage(ctx, "assetft-ibcenable-"+denom, &types.MsgEnableIBCExecutor{Denom: denom}, params.IbcGracePeriod)
	if err != nil {
		return err
	}

	store.Set(key, asset.StoreTrue)

	return nil
}

// EnableIBC enables IBC.
func (k EnableIBCKeeper) EnableIBC(ctx sdk.Context, denom string) error {
	def, err := k.keeper.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if def.IsFeatureEnabled(types.Feature_ibc) {
		return errors.Errorf("ibc has been already enabled for denom: %s", denom)
	}

	subunit, issuer, err := types.DeconstructDenom(denom)
	if err != nil {
		return err
	}

	def.Features = append(def.Features, types.Feature_ibc)
	k.keeper.SetDefinition(ctx, issuer, subunit, def)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), ibcEnablePrefix)
	key := []byte(denom)
	store.Delete(key)

	return nil
}
