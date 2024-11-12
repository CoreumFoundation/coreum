package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) increaseFTLimits(
	ctx sdk.Context,
	addr sdk.AccAddress,
	lockedCoin, expectedToReceiveCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Increasing DEX FT limits.",
		"addr", addr,
		"lockedCoin", lockedCoin.String(),
		"expectedToReceiveCoin", expectedToReceiveCoin.String(),
	)

	if err := k.assetFTKeeper.DEXIncreaseLimits(ctx, addr, lockedCoin, expectedToReceiveCoin); err != nil {
		return sdkerrors.Wrap(err, "failed to increase DEX FT limits")
	}

	return nil
}

func (k Keeper) decreaseFTLimits(
	ctx sdk.Context,
	addr sdk.AccAddress,
	lockedCoin, expectedToReceiveCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Decreasing DEX FT limits.",
		"addr", addr,
		"lockedCoin", lockedCoin.String(),
		"expectedToReceiveCoin", expectedToReceiveCoin.String(),
	)

	if err := k.assetFTKeeper.DEXDecreaseLimits(ctx, addr, lockedCoin, expectedToReceiveCoin); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to decrease DEX FT limits, err: %s", err)
	}

	return nil
}

func (k Keeper) decreaseFTLimitsAndSend(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	unlockAndSendCoin, expectedToReceiveCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Decreasing DEX FT limits and sending.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"unlockAndSendCoin", unlockAndSendCoin.String(),
		"expectedToReceiveCoin", expectedToReceiveCoin.String(),
	)

	if err := k.assetFTKeeper.DEXDecreaseLimitsAndSend(
		ctx, fromAddr, toAddr, unlockAndSendCoin, expectedToReceiveCoin,
	); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to decrease DEX FT limits and send, err: %s", err)
	}

	return nil
}

func (k Keeper) checkFTLimitsAndSend(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	sendCoin, checkExpectedToReceiveCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Checking DEX FT limits and sending.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"sendCoin", sendCoin.String(),
		"checkExpectedToReceiveCoin", checkExpectedToReceiveCoin.String(),
	)

	if err := k.assetFTKeeper.DEXCheckLimitsAndSend(
		ctx,
		fromAddr, toAddr,
		sendCoin, checkExpectedToReceiveCoin,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to check DEX FT limits and send")
	}

	return nil
}

func (k Keeper) lockFT(
	ctx sdk.Context,
	addr sdk.AccAddress,
	lockCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Locking FT coin.",
		"addr", addr,
		"lockCoin", lockCoin.String(),
	)

	if err := k.assetFTKeeper.DEXLock(ctx, addr, lockCoin); err != nil {
		return sdkerrors.Wrap(err, "failed to lock DEX FT coin")
	}

	return nil
}

func (k Keeper) unlockFT(
	ctx sdk.Context,
	addr sdk.AccAddress,
	unlockCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Unlocking FT coin.",
		"addr", addr,
		"unlockCoin", unlockCoin.String(),
	)

	if err := k.assetFTKeeper.DEXUnlock(ctx, addr, unlockCoin); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to unlock DEX FT coin, err: %s", err)
	}

	return nil
}
