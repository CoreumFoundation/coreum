package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// FreezeKeeper defines an interface which can be used to freeze/unfreeze balances
type FreezeKeeper interface {
	FreezeToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error
	UnfreezeToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error
	GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// FreezeToken freezes specified token from the specified account
func (k Keeper) FreezeToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	frozenStore := k.getFrozenBalanceStore(ctx, addr)

	ft, err := k.GetFungibleToken(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if ft.Issuer != issuer.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only issuer is authorized to perform this operation")
	}

	if err := types.FTHasOption(ft.Options, types.FungibleTokenOption_Freezable); err != nil { //nolint:nosnakecase
		return sdkerrors.Wrapf(err, "denom:%s, option:%s", coin.Denom, types.FungibleTokenOption_Freezable) //nolint:nosnakecase
	}

	if err := k.areCoinsSpendable(ctx, addr, sdk.NewCoins(coin)); err != nil {
		return err
	}

	bz := k.cdc.MustMarshal(&coin)

	frozenStore.Set([]byte(coin.Denom), bz)

	return nil
}

// areCoinsSpendable returns an error is there are not enough coins balances to be spent
func (k Keeper) areCoinsSpendable(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		frozenBalance := k.GetFrozenBalance(ctx, addr, coin.Denom)
		balance := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
		if !balance.IsGTE(frozenBalance.Add(coin)) {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is not available", coin)
		}
	}
	return nil
}

// GetFrozenBalance returns the frozen balance of a denom on an account
func (k Keeper) GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	frozenStore := k.getFrozenBalanceStore(ctx, addr)
	frozenBalance := sdk.NewCoin(denom, sdk.NewInt(0))
	if bz := frozenStore.Get([]byte(denom)); bz != nil {
		k.cdc.MustUnmarshal(bz, &frozenBalance)
	}

	return frozenBalance
}

// GetFrozenBalances returns the frozen balance on an account
func (k Keeper) GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	frozenStore := k.getFrozenBalanceStore(ctx, addr)
	iterator := frozenStore.Iterator(nil, nil)
	coins := sdk.NewCoins()
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var balance sdk.Coin
		k.cdc.MustUnmarshal(iterator.Value(), &balance)
		coins = append(coins, balance)
	}
	return coins
}

// UnfreezeToken unfreezes specified tokens from the specified account
func (k Keeper) UnfreezeToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	frozenStore := k.getFrozenBalanceStore(ctx, addr)

	frozenBalance := k.GetFrozenBalance(ctx, addr, coin.Denom)
	if frozenBalance.IsLT(coin) {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, "not enough frozen coins")
	}

	ft, err := k.GetFungibleToken(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if ft.Issuer != issuer.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only issuer is authorized to perform this operation")
	}

	newFrozenBalance := frozenBalance.Sub(coin)
	if newFrozenBalance.IsZero() {
		frozenStore.Delete([]byte(coin.Denom))
	} else {
		bz := k.cdc.MustMarshal(&newFrozenBalance)
		frozenStore.Set([]byte(coin.Denom), bz)
	}

	return nil
}

// getFrozenBalanceStore get the store for the frozen balances of an account
func (k Keeper) getFrozenBalanceStore(ctx sdk.Context, addr sdk.AccAddress) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.CreateFrozenBalancesPrefix(addr))
}
