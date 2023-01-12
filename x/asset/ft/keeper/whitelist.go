package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// SetWhitelistedBalance sets whitelisted limit for the account
func (k Keeper) SetWhitelistedBalance(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsNil() || coin.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "whitelisted limit amount should be greater than or equal to 0")
	}

	ft, err := k.GetTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if err = ft.CheckFeatureAllowed(sender, types.TokenFeature_whitelist); err != nil { //nolint:nosnakecase
		return err
	}

	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	previousWhitelistedBalance := whitelistedStore.Balance(coin.Denom)
	whitelistedStore.SetBalance(coin)

	return ctx.EventManager().EmitTypedEvent(&types.EventWhitelistedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: previousWhitelistedBalance.Amount,
		CurrentAmount:  coin.Amount,
	})
}

// SetWhitelistedBalances sets the whitelisted balances of a specified account
func (k Keeper) SetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		whitelistedStore.SetBalance(coin)
	}
}

// GetWhitelistedBalance returns the whitelisted balance of a denom and account
func (k Keeper) GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balance(denom)
}

// GetWhitelistedBalances returns the whitelisted balance of an account
func (k Keeper) GetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetAccountsWhitelistedBalances returns the whitelisted balance of all the account
func (k Keeper) GetAccountsWhitelistedBalances(ctx sdk.Context, pagination *query.PageRequest) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.whitelistedBalancesStore(ctx), pagination)
}

// IterateAllWhitelistedBalances iterates over all whitelisted balances of all accounts and applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllWhitelistedBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) error {
	return k.whitelistedAccountBalancesStore(ctx).IterateAllBalances(cb)
}

// whitelistedBalancesStore get the store for the whitelisted balances of all accounts
func (k Keeper) whitelistedBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.WhitelistedBalancesKeyPrefix)
}

// whitelistedAccountBalanceStore gets the store for the whitelisted balances of an account
func (k Keeper) whitelistedAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	store := ctx.KVStore(k.storeKey)
	return newBalanceStore(k.cdc, store, types.CreateWhitelistedBalancesKey(addr))
}

// whitelistedAccountBalancesStore gets the store for the whitelisted balances
func (k Keeper) whitelistedAccountBalancesStore(ctx sdk.Context) balanceStore {
	store := ctx.KVStore(k.storeKey)
	return newBalanceStore(k.cdc, store, types.WhitelistedBalancesKeyPrefix)
}

// areCoinsReceivable returns an error if whitelisted amount is too low to receive coins
func (k Keeper) isCoinReceivable(ctx sdk.Context, addr sdk.AccAddress, ft types.FTDefinition, amount sdk.Int) error {
	if !ft.IsFeatureEnabled(types.TokenFeature_whitelist) || ft.IsIssuer(addr) { //nolint:nosnakecase
		return nil
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, ft.Denom)
	whitelistedBalance := k.GetWhitelistedBalance(ctx, addr, ft.Denom)

	finalBalance := balance.Amount.Add(amount)
	if finalBalance.GT(whitelistedBalance.Amount) {
		return sdkerrors.Wrapf(types.ErrWhitelistedLimitExceeded, "balance whitelisted for %s is not enough to receive %s, current whitelisted balance: %s",
			addr, sdk.NewCoin(ft.Denom, amount), whitelistedBalance)
	}
	return nil
}
