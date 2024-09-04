package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) setUint64Seq(ctx sdk.Context, key []byte, seq uint64) error {
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

func (k Keeper) setUint32Seq(ctx sdk.Context, key []byte, seq uint32) error {
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
