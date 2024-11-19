package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) incrementUint64Counter(ctx sdk.Context, key []byte) (uint64, error) {
	return k.genNextUint64Sequence(ctx, key)
}

func (k Keeper) decrementUint64Counter(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		return 0, err
	}
	val.Value--

	return val.Value, k.setDataToStore(ctx, key, &val)
}

func (k Keeper) getUint64Value(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		return 0, err
	}

	return val.Value, nil
}

func (k Keeper) setUint64Value(ctx sdk.Context, key []byte, sequence uint64) error {
	val := gogotypes.UInt64Value{
		Value: sequence,
	}

	return k.setDataToStore(ctx, key, &val)
}

func (k Keeper) genNextUint64Sequence(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
	}
	// start with 1
	val.Value++

	return val.Value, k.setDataToStore(ctx, key, &val)
}

func (k Keeper) setUint32Value(ctx sdk.Context, key []byte, sequence uint32) error {
	val := gogotypes.UInt32Value{
		Value: sequence,
	}

	return k.setDataToStore(ctx, key, &val)
}

func (k Keeper) genNextUint32Sequence(ctx sdk.Context, key []byte) (uint32, error) {
	var val gogotypes.UInt32Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
	}
	// start with 1
	val.Value++

	return val.Value, k.setDataToStore(ctx, key, &val)
}

func (k Keeper) setDataToStore(
	ctx sdk.Context,
	key []byte,
	val proto.Message,
) error {
	bz, err := k.cdc.Marshal(val)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to marshal %T, err: %s", err, val)
	}
	return k.storeService.OpenKVStore(ctx).Set(key, bz)
}

func (k Keeper) getDataFromStore(
	ctx sdk.Context,
	key []byte,
	val proto.Message,
) error {
	bz, _ := k.storeService.OpenKVStore(ctx).Get(key)
	if bz == nil {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "store type %T", val)
	}

	if err := k.cdc.Unmarshal(bz, val); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal %T, err: %s", err, val)
	}

	return nil
}
