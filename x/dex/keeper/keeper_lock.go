package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) lockOrderBalance(
	ctx sdk.Context,
	order types.Order,
) (sdkmath.Int, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdkmath.Int{}, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	var lockedBalance sdk.Coin
	switch order.Type {
	case types.ORDER_TYPE_LIMIT:
		var err error
		lockedBalance, err = order.ComputeLimitOrderLockedBalance()
		if err != nil {
			return sdkmath.Int{}, err
		}
	case types.ORDER_TYPE_MARKET:
		if order.Side == types.SIDE_BUY {
			// for the buy market order we lock the entire spendable amount
			lockedBalance = k.assetFTKeeper.GetSpendableBalance(ctx, creatorAddr, order.QuoteDenom)
		} else {
			lockedBalance = sdk.NewCoin(order.BaseDenom, order.Quantity)
		}
	default:
		return sdkmath.Int{}, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpect order type : %s", order.Type.String(),
		)
	}

	if !lockedBalance.IsPositive() {
		return sdkmath.Int{}, sdkerrors.Wrapf(
			cosmoserrors.ErrInsufficientFunds, "no funds of denom: %s to lock", lockedBalance.Denom,
		)
	}

	if err := k.lockCoin(ctx, creatorAddr, lockedBalance); err != nil {
		return sdkmath.Int{}, err
	}

	k.logger(ctx).Debug("Locked order balance.", "lockedBalance", lockedBalance)

	return lockedBalance.Amount, nil
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
