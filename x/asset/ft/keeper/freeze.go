package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// Freeze freezes specified token from the specified account
func (k Keeper) Freeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	ft, err := k.GetTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	newFrozenBalance := frozenBalance.Add(coin)
	frozenStore.SetBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		PreviousAmount: frozenBalance,
		CurrentAmount:  newFrozenBalance,
	})
}

// Unfreeze unfreezes specified tokens from the specified account
func (k Keeper) Unfreeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	ft, err := k.GetTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	if !frozenBalance.IsGTE(coin) {
		return sdkerrors.Wrapf(types.ErrNotEnoughBalance,
			"unfreeze request %s is greater than the available frozen balance %s",
			coin.String(),
			frozenBalance.String(),
		)
	}

	newFrozenBalance := frozenBalance.Sub(coin)
	frozenStore.SetBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		PreviousAmount: frozenBalance,
		CurrentAmount:  newFrozenBalance,
	})
}

// SetFrozenBalances sets the frozen balances of a specified account
func (k Keeper) SetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		frozenStore.SetBalance(coin)
	}
}

// areCoinsSpendable returns an error if there are not enough coins balances to be spent
func (k Keeper) isCoinSpendable(ctx sdk.Context, addr sdk.AccAddress, ft types.FTDefinition, amount sdk.Int) error {
	if k.isGloballyFrozen(ctx, ft.Denom) {
		return sdkerrors.Wrapf(types.ErrGloballyFrozen, "%s is globally frozen", ft.Denom)
	}

	availableBalance := k.availableBalance(ctx, addr, ft.Denom)
	if !availableBalance.Amount.GTE(amount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is not available, available %s",
			sdk.NewCoin(ft.Denom, amount), availableBalance)
	}
	return nil
}

func (k Keeper) availableBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balance := k.bankKeeper.GetBalance(ctx, addr, denom)
	if balance.IsZero() {
		return balance
	}

	frozenBalance := k.GetFrozenBalance(ctx, addr, denom)
	if frozenBalance.IsGTE(balance) {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}
	return balance.Sub(frozenBalance)
}

// GetFrozenBalance returns the frozen balance of a denom and account
func (k Keeper) GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.frozenAccountBalanceStore(ctx, addr).Balance(denom)
}

// GetFrozenBalances returns the frozen balance of an account
func (k Keeper) GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.frozenAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetAccountsFrozenBalances returns the frozen balance on all the account
func (k Keeper) GetAccountsFrozenBalances(ctx sdk.Context, pagination *query.PageRequest) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.frozenBalancesStore(ctx), pagination)
}

// frozenBalancesStore get the store for the frozen balances of all accounts
func (k Keeper) frozenBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.FrozenBalancesKeyPrefix)
}

// frozenAccountBalanceStore gets the store for the frozen balances of an account
func (k Keeper) frozenAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	store := ctx.KVStore(k.storeKey)
	return newBalanceStore(k.cdc, store, types.CreateFrozenBalancesPrefix(addr))
}
