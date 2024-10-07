package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) increaseFTLimits(
	ctx sdk.Context,
	addr sdk.AccAddress,
	lockCoin, reserveWhitelistingCoinCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Increasing DEX FT limits.",
		"addr", addr,
		"lockCoin", lockCoin.String(),
		"reserveWhitelistingCoinCoin", reserveWhitelistingCoinCoin.String(),
	)

	if err := k.assetFTKeeper.DEXIncreaseLimits(ctx, addr, lockCoin, reserveWhitelistingCoinCoin); err != nil {
		return sdkerrors.Wrap(err, "failed to increase DEX FT limits")
	}

	return nil
}

func (k Keeper) decreaseFTLimits(
	ctx sdk.Context,
	addr sdk.AccAddress,
	unlockCoin, releaseWhitelistingCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Decreasing DEX FT limits.",
		"addr", addr,
		"unlockCoin", unlockCoin.String(),
		"releaseWhitelistingCoin", releaseWhitelistingCoin.String(),
	)

	if err := k.assetFTKeeper.DEXDecreaseLimits(ctx, addr, unlockCoin, releaseWhitelistingCoin); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to decrease DEX FT limits, err: %s", err)
	}

	return nil
}

func (k Keeper) decreaseFTLimitsAndSend(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	unlockAndSendCoin, releaseWhitelistingCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Decreasing DEX FT limits and sending.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"unlockAndSendCoin", unlockAndSendCoin.String(),
		"releaseWhitelistingCoin", releaseWhitelistingCoin.String(),
	)

	if err := k.assetFTKeeper.DEXDecreaseLimitsAndSend(
		ctx, fromAddr, toAddr, unlockAndSendCoin, releaseWhitelistingCoin,
	); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to decrease DEX FT limits and send, err: %s", err)
	}

	return nil
}

func (k Keeper) checksFTLimitsAndSend(
	ctx sdk.Context,
	fromAddr, toAddr sdk.AccAddress,
	sendCoin, checkReserveWhitelistingCoin sdk.Coin,
) error {
	k.logger(ctx).Debug(
		"Checking DEX FT limits and sending.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"sendCoin", sendCoin.String(),
		"checkReserveWhitelistingCoin", checkReserveWhitelistingCoin.String(),
	)

	if err := k.assetFTKeeper.DEXChecksLimitsAndSend(
		ctx,
		fromAddr, toAddr,
		sendCoin, checkReserveWhitelistingCoin,
	); err != nil {
		return sdkerrors.Wrap(err, "failed to check DEX FT limits and send")
	}

	return nil
}
