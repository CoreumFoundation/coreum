package keeper

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func validateGoodTil(ctx sdk.Context, order types.Order) error {
	if order.GoodTil.GoodTilBlockHeight > 0 {
		currentHeight := ctx.BlockHeight()
		if order.GoodTil.GoodTilBlockHeight <= uint64(currentHeight) {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"good til block height %d must be greater than current block height %d",
				order.GoodTil.GoodTilBlockHeight, currentHeight,
			)
		}
	}
	if order.GoodTil.GoodTilBlockTime != nil {
		currentTime := ctx.BlockTime()
		if !order.GoodTil.GoodTilBlockTime.After(currentTime) {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"good til block time %s must be greater than current block time %s",
				order.GoodTil.GoodTilBlockTime, currentTime,
			)
		}
	}

	return nil
}

func (k Keeper) delayGoodTilCancellation(
	ctx sdk.Context,
	goodTil types.GoodTil,
	orderSequence uint64,
	creator sdk.AccAddress,
) error {
	if goodTil.GoodTilBlockHeight > 0 {
		return k.delayGoodTilBlockHeightCancellation(ctx, goodTil.GoodTilBlockHeight, orderSequence, creator)
	}
	if goodTil.GoodTilBlockTime != nil {
		return k.delayGoodTilBlockTimeCancellation(ctx, *goodTil.GoodTilBlockTime, orderSequence, creator)
	}

	return nil
}

func (k Keeper) delayGoodTilBlockHeightCancellation(
	ctx sdk.Context,
	height uint64,
	orderSequence uint64,
	creator sdk.AccAddress,
) error {
	k.logger(ctx).Debug(
		"Delaying good til height cancellation.",
		"orderSequence", orderSequence,
		"height", height,
		"creator", creator.String(),
	)
	if err := k.delayKeeper.ExecuteAfterBlock(
		ctx,
		types.BuildGoodTilBlockHeightDelayKey(orderSequence),
		&types.CancelGoodTil{
			Creator:       creator.String(),
			OrderSequence: orderSequence,
		},
		height,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to create good til height delayed cancellation")
	}

	return nil
}

func (k Keeper) delayGoodTilBlockTimeCancellation(
	ctx sdk.Context,
	time time.Time,
	orderSequence uint64,
	creator sdk.AccAddress,
) error {
	k.logger(ctx).Debug(
		"Delaying good til time cancellation.",
		"orderSequence", orderSequence,
		"time", time,
		"creator", creator.String(),
	)
	if err := k.delayKeeper.ExecuteAfter(
		ctx,
		types.BuildGoodTilBlockTimeDelayKey(orderSequence),
		&types.CancelGoodTil{
			Creator:       creator.String(),
			OrderSequence: orderSequence,
		},
		time,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to create good til time delayed cancellation")
	}

	return nil
}

func (k Keeper) removeGoodTilDelay(ctx sdk.Context, goodTil types.GoodTil, orderSequence uint64) error {
	if goodTil.GoodTilBlockHeight > 0 {
		if err := k.removeGoodTilBlockHeightCancellation(ctx, goodTil.GoodTilBlockHeight, orderSequence); err != nil {
			return err
		}
	}
	if goodTil.GoodTilBlockTime != nil {
		if err := k.removeGoodTilBlockTimeCancellation(ctx, *goodTil.GoodTilBlockTime, orderSequence); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) removeGoodTilBlockHeightCancellation(
	ctx sdk.Context,
	height uint64,
	orderSequence uint64,
) error {
	k.logger(ctx).Debug("Removing good til height delayed cancellation.", "orderSequence", orderSequence, "height", height)
	if err := k.delayKeeper.RemoveExecuteAtBlock(
		ctx,
		types.BuildGoodTilBlockHeightDelayKey(orderSequence),
		height,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to remove good til height delayed cancellation")
	}

	return nil
}

func (k Keeper) removeGoodTilBlockTimeCancellation(
	ctx sdk.Context,
	time time.Time,
	orderSequence uint64,
) error {
	k.logger(ctx).Debug("Removing good til time delayed cancellation.", "orderSequence", orderSequence, "time", time)
	if err := k.delayKeeper.RemoveExecuteAfter(
		ctx,
		types.BuildGoodTilBlockTimeDelayKey(orderSequence),
		time,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to remove good til time delayed cancellation")
	}

	return nil
}
