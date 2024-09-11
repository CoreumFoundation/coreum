package keeper

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) delayGoodTilCancellation(
	ctx sdk.Context,
	goodTil types.GoodTil,
	orderSeq uint64,
	creator sdk.AccAddress,
	orderID string,
) error {
	if goodTil.GoodTilBlockHeight > 0 {
		return k.delayGoodTilBlockHeightCancellation(ctx, goodTil.GoodTilBlockHeight, orderSeq, creator, orderID)
	}
	if goodTil.GoodTilBlockTime != nil {
		return k.delayGoodTilBlockTimeCancellation(ctx, *goodTil.GoodTilBlockTime, orderSeq, creator, orderID)
	}

	return nil
}

func (k Keeper) delayGoodTilBlockHeightCancellation(
	ctx sdk.Context,
	height uint64,
	orderSeq uint64,
	creator sdk.AccAddress,
	orderID string,
) error {
	k.logger(ctx).Debug(
		"Delaying good til height cancellation.",
		"orderSeq", orderSeq,
		"height", height,
		"creator", creator.String(),
		"orderID", orderID,
	)
	if err := k.delayKeeper.ExecuteAfterBlock(
		ctx,
		types.BuildGoodTilBlockHeightDelayKey(orderSeq),
		&types.CancelGoodTil{
			Creator: creator.String(),
			OrderID: orderID,
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
	orderSeq uint64,
	creator sdk.AccAddress,
	orderID string,
) error {
	k.logger(ctx).Debug(
		"Delaying good til time cancellation.",
		"orderSeq", orderSeq,
		"time", time,
		"creator", creator.String(),
		"orderID", orderID,
	)
	if err := k.delayKeeper.ExecuteAfter(
		ctx,
		types.BuildGoodTilBlockTimeDelayKey(orderSeq),
		&types.CancelGoodTil{
			Creator: creator.String(),
			OrderID: orderID,
		},
		time,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to create good til time delayed cancellation")
	}

	return nil
}

func (k Keeper) removeGoodTilDelay(ctx sdk.Context, goodTil types.GoodTil, orderSeq uint64) error {
	if goodTil.GoodTilBlockHeight > 0 {
		if err := k.removeGoodTilBlockHeightCancellation(ctx, goodTil.GoodTilBlockHeight, orderSeq); err != nil {
			return err
		}
	}
	if goodTil.GoodTilBlockTime != nil {
		if err := k.removeGoodTilBlockTimeCancellation(ctx, *goodTil.GoodTilBlockTime, orderSeq); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) removeGoodTilBlockHeightCancellation(
	ctx sdk.Context,
	height uint64,
	orderSeq uint64,
) error {
	k.logger(ctx).Debug("Removing good til height delayed cancellation.", "orderSeq", orderSeq, "height", height)
	if err := k.delayKeeper.RemoveExecuteAtBlock(
		ctx,
		types.BuildGoodTilBlockHeightDelayKey(orderSeq),
		height,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to remove good til height delayed cancellation")
	}

	return nil
}

func (k Keeper) removeGoodTilBlockTimeCancellation(
	ctx sdk.Context,
	time time.Time,
	orderSeq uint64,
) error {
	k.logger(ctx).Debug("Removing good til time delayed cancellation.", "orderSeq", orderSeq, "time", time)
	if err := k.delayKeeper.RemoveExecuteAfter(
		ctx,
		types.BuildGoodTilBlockTimeDelayKey(orderSeq),
		time,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to remove good til time delayed cancellation")
	}

	return nil
}
