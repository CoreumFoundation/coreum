package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) lockOrderBalance(ctx sdk.Context, order types.Order) (sdk.Coin, error) {
	lockedBalance, err := order.ComputeLockedBalance()
	if err != nil {
		return sdk.Coin{}, err
	}

	creatorAddr, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	if err := k.lockCoin(ctx, creatorAddr, lockedBalance); err != nil {
		return sdk.Coin{}, err
	}

	k.logger(ctx).Debug("Locked order balance.", "lockedBalance", lockedBalance)

	return lockedBalance, nil
}

func (k Keeper) lockCoin(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	// don't check for empty coin because we don't expect the coin here
	k.logger(ctx).Debug(
		"Locking DEX coin.",
		"addr", addr,
		"coin", coin.String(),
	)

	return k.assetFTKeeper.DEXLock(ctx, addr, coin)
}

func (k Keeper) unlockCoin(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsZero() {
		k.logger(ctx).Debug(
			"Nothing to unlock.",
			"addr", addr,
			"coin", coin.String(),
		)
		return nil
	}

	k.logger(ctx).Debug(
		"Unlocking DEX coin.",
		"addr", addr,
		"coin", coin.String(),
	)

	return k.assetFTKeeper.DEXUnlock(ctx, addr, coin)
}

func (k Keeper) unlockAndSendCoin(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsZero() {
		k.logger(ctx).Debug(
			"Nothing to unlock and send.",
			"fromAddr", fromAddr,
			"toAddr", toAddr,
			"coin", coin.String(),
		)
		return nil
	}

	k.logger(ctx).Debug(
		"Unlocking and sending DEX coin.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"coin", coin.String(),
	)

	return k.assetFTKeeper.DEXUnlockAndSend(ctx, fromAddr, toAddr, coin)
}
