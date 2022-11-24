package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// FreezeFungibleToken freezes specified token from the specified account
func (k Keeper) FreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	err := k.checkFeatureAllowed(ctx, sender, coin, types.FungibleTokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.getFrozenBalance(coin.Denom)
	newFrozenBalance := frozenBalance.Add(coin)
	frozenStore.setFrozenBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenFrozen{
		Account: addr.String(),
		Coin:    coin,
	})
}

// UnfreezeFungibleToken unfreezes specified tokens from the specified account
func (k Keeper) UnfreezeFungibleToken(ctx sdk.Context, sender sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	err := k.checkFeatureAllowed(ctx, sender, coin, types.FungibleTokenFeature_freeze) //nolint:nosnakecase
	if err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.getFrozenBalance(coin.Denom)
	if frozenBalance.IsGTE(coin) {
		newFrozenBalance := frozenBalance.Sub(coin)
		frozenStore.setFrozenBalance(newFrozenBalance)
	} else {
		return sdkerrors.Wrapf(types.ErrNotEnoughBalance,
			"unfreeze request %s is greater than the available frozen balance %s",
			coin.String(),
			frozenBalance.String(),
		)
	}

	return ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenUnfrozen{
		Account: addr.String(),
		Coin:    coin,
	})
}

// SetFrozenBalances sets the frozen balances of a specified account
func (k Keeper) SetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		frozenStore.setFrozenBalance(coin)
	}
}

// areCoinsSpendable returns an error is there are not enough coins balances to be spent
func (k Keeper) areCoinsSpendable(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		availableBalance := k.availableBalance(ctx, addr, coin.Denom)
		if !availableBalance.IsGTE(coin) {
			return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is not available", coin)
		}
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
	return k.frozenAccountBalanceStore(ctx, addr).getFrozenBalance(denom)
}

// GetFrozenBalances returns the frozen balance of an account
func (k Keeper) GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.frozenAccountBalanceStore(ctx, addr).getFrozenBalances(pagination)
}

// GetAccountsFrozenBalances returns the frozen balance on all of the account
func (k Keeper) GetAccountsFrozenBalances(ctx sdk.Context, pagination *query.PageRequest) ([]types.Balance, *query.PageResponse, error) {
	frozenStore := k.frozenBalancesStore(ctx)
	var balances []types.Balance
	mapAddressToBalancesIdx := make(map[string]int)
	pageRes, err := query.Paginate(frozenStore, pagination, func(key, value []byte) error {
		address, err := types.AddressFromBalancesStore(key)
		if err != nil {
			return err
		}

		var coin sdk.Coin
		k.cdc.MustUnmarshal(value, &coin)

		idx, ok := mapAddressToBalancesIdx[address.String()]
		if ok {
			// address is already on the set of accounts balances
			balances[idx].Coins = balances[idx].Coins.Add(coin)
			balances[idx].Coins.Sort()
			return nil
		}

		accountBalance := types.Balance{
			Address: address.String(),
			Coins:   sdk.NewCoins(coin),
		}
		balances = append(balances, accountBalance)
		mapAddressToBalancesIdx[address.String()] = len(balances) - 1
		return nil
	})

	return balances, pageRes, err
}

// frozenBalancesStore get the store for the frozen balances of all accounts
func (k Keeper) frozenBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.FrozenBalancesKeyPrefix)
}

type frozenAccountBalanceStore struct {
	store prefix.Store
	cdc   codec.BinaryCodec
}

func (s frozenAccountBalanceStore) getFrozenBalance(denom string) sdk.Coin {
	frozenBalance := sdk.NewCoin(denom, sdk.ZeroInt())
	if bz := s.store.Get([]byte(denom)); bz != nil {
		s.cdc.MustUnmarshal(bz, &frozenBalance)
	}

	return frozenBalance
}

func (s frozenAccountBalanceStore) getFrozenBalances(pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	coins := sdk.NewCoins()
	pageRes, err := query.Paginate(s.store, pagination, func(key, value []byte) error {
		var coin sdk.Coin
		s.cdc.MustUnmarshal(value, &coin)
		coins = append(coins, coin)
		return nil
	})
	return coins, pageRes, err
}

func (s frozenAccountBalanceStore) setFrozenBalance(coin sdk.Coin) {
	if coin.Amount.IsZero() {
		s.store.Delete([]byte(coin.Denom))
	} else {
		bz := s.cdc.MustMarshal(&coin)
		s.store.Set([]byte(coin.Denom), bz)
	}
}

// frozenAccountBalanceStore get the store for the frozen balances of an account
func (k Keeper) frozenAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) frozenAccountBalanceStore {
	store := ctx.KVStore(k.storeKey)
	return frozenAccountBalanceStore{
		store: prefix.NewStore(store, types.CreateFrozenBalancesPrefix(addr)),
		cdc:   k.cdc,
	}
}

// isFeatureEnabled checks weather a feature is present on a list of token features
func isFeatureEnabled(features []types.FungibleTokenFeature, feature types.FungibleTokenFeature) bool {
	for _, o := range features {
		if o == feature {
			return true
		}
	}
	return false
}
