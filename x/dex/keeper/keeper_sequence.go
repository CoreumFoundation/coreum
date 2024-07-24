package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

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
