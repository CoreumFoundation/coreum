package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) incrementUint64Counter(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
		val.Value = 1 // start with 1
	} else {
		val.Value++
	}

	return val.Value, k.setDataToStore(ctx, key, &val)
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

func (k Keeper) setUint64Value(ctx sdk.Context, key []byte, seq uint64) error {
	val := gogotypes.UInt64Value{
		Value: seq,
	}

	return k.setDataToStore(ctx, key, &val)
}

func (k Keeper) genNextUint64Seq(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
	} else {
		val.Value++
	}

	return val.Value, k.setDataToStore(ctx, key, &val)
}

func (k Keeper) setUint32Value(ctx sdk.Context, key []byte, seq uint32) error {
	val := gogotypes.UInt32Value{
		Value: seq,
	}

	return k.setDataToStore(ctx, key, &val)
}

func (k Keeper) genNextUint32Seq(ctx sdk.Context, key []byte) (uint32, error) {
	var val gogotypes.UInt32Value
	err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
	} else {
		val.Value++
	}

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
	ctx.KVStore(k.storeKey).Set(key, bz)
	return nil
}

func (k Keeper) getDataFromStore(
	ctx sdk.Context,
	key []byte,
	val proto.Message,
) error {
	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "store type %T", val)
	}

	if err := k.cdc.Unmarshal(bz, val); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal %T, err: %s", err, val)
	}

	return nil
}
