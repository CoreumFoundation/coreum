package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) sendCoinToDEX(ctx sdk.Context, fromAddr string, coin sdk.Coin) error {
	acc, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", fromAddr)
	}

	k.logger(ctx).Debug(
		"Sending coin to DEX.",
		"fromAddr", fromAddr,
		"coin", coin.String(),
	)

	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, acc, types.ModuleName, sdk.NewCoins(coin))
}

func (k Keeper) sendCoinsFromDEX(ctx sdk.Context, toNumber uint64, coins sdk.Coins) error {
	if coins.IsZero() {
		k.logger(ctx).Debug(
			"Skipping sending coin from DEX, nothing to send",
			"coins", coins,
		)
		return nil
	}
	toAddr, err := k.getAccountAddress(ctx, toNumber)
	if err != nil {
		return err
	}

	k.logger(ctx).Debug(
		"Sending coin from DEX.",
		"toAddr", toAddr.String(),
		"coin", coins.String(),
	)

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, coins)
}
