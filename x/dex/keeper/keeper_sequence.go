package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"
)

func (k Keeper) genNextUint64Seq(ctx sdk.Context, key []byte) (uint64, error) {
	var val gogotypes.UInt64Value
	found, err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		return 0, err
	}
	// start with zero
	if found {
		val.Value++
	}
	return val.Value, k.setDataToStore(ctx, key, &val)
}

func (k Keeper) genNextUint32Seq(ctx sdk.Context, key []byte) (uint32, error) {
	var val gogotypes.UInt32Value
	found, err := k.getDataFromStore(ctx, key, &val)
	if err != nil {
		return 0, err
	}
	// start with zero
	if found {
		val.Value++
	}
	return val.Value, k.setDataToStore(ctx, key, &val)
}
