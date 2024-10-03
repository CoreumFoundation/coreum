package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) lockCoin(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin, receiveDenom string) error {
	k.logger(ctx).Debug(
		"Locking DEX coin.",
		"addr", addr,
		"coin", coin.String(),
		"receiveDenom", receiveDenom,
	)

	if err := k.assetFTKeeper.DEXLock(ctx, addr, coin, receiveDenom); err != nil {
		return sdkerrors.Wrapf(types.ErrFailedToLockCoin, "failed to lock order coins: %s", err)
	}

	return nil
}

func (k Keeper) unlockCoin(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug(
		"Unlocking DEX coin.",
		"addr", addr,
		"coin", coin.String(),
	)

	return k.assetFTKeeper.DEXUnlock(ctx, addr, coin)
}

func (k Keeper) unlockAndSendCoin(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug(
		"Unlocking and sending DEX coin.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"coin", coin.String(),
	)

	return k.assetFTKeeper.DEXUnlockAndSend(ctx, fromAddr, toAddr, coin)
}

func (k Keeper) sendCoinWithLockCheck(
	ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, coin sdk.Coin, receiveDenom string,
) error {
	k.logger(ctx).Debug(
		"Sending DEX coin with lock check.",
		"fromAddr", fromAddr,
		"toAddr", toAddr,
		"coin", coin.String(),
	)

	if err := k.assetFTKeeper.DEXSendWithLockCheck(ctx, fromAddr, toAddr, coin, receiveDenom); err != nil {
		return sdkerrors.Wrapf(types.ErrFailedToSendCoinWithLockCheck, "failed to send coins with lock check: %s", err)
	}

	return nil
}
