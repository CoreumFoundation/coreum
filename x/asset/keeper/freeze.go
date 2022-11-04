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
func (k Keeper) FreezeFungibleToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	err := k.isFreezeAllowed(ctx, issuer, coin)
	if err != nil {
		return err
	}

	frozenStore := k.frozenBalanceStore(ctx, addr)
	frozenBalance := frozenStore.getFrozenBalance(coin.Denom)
	newFrozenBalance := frozenBalance.Add(coin)
	bankBalance := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
	if bankBalance.IsLT(newFrozenBalance) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds,
			"account balance %s is less that desired frozen balance %s is not available",
			bankBalance.String(),
			newFrozenBalance.String(),
		)
	}

	frozenStore.setFrozenBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenFrozen{
		Account: addr.String(),
		Coin:    coin,
	})
}

// UnfreezeFungibleToken unfreezes specified tokens from the specified account
func (k Keeper) UnfreezeFungibleToken(ctx sdk.Context, issuer sdk.AccAddress, addr sdk.AccAddress, coin sdk.Coin) error {
	err := k.isFreezeAllowed(ctx, issuer, coin)
	if err != nil {
		return err
	}

	frozenStore := k.frozenBalanceStore(ctx, addr)
	frozenBalance := frozenStore.getFrozenBalance(coin.Denom)
	if frozenBalance.IsLT(coin) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "unfreeze amount is more the frozen balance %s", coin.String())
	}

	newFrozenBalance := frozenBalance.Sub(coin)
	frozenStore.setFrozenBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenUnfrozen{
		Account: addr.String(),
		Coin:    coin,
	})
}

func (k Keeper) isFreezeAllowed(ctx sdk.Context, issuer sdk.AccAddress, coin sdk.Coin) error {
	ft, err := k.getFungibleTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if ft.Issuer != issuer.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "address is unauthorized to perform this operation")
	}

	if err := isFeatureEnabled(ft.Features, types.FungibleTokenFeature_freezable); err != nil { //nolint:nosnakecase
		return sdkerrors.Wrapf(err, "denom:%s, feature:%s", coin.Denom, types.FungibleTokenFeature_freezable) //nolint:nosnakecase
	}

	return nil
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
	return balance.Sub(frozenBalance)
}

// GetFrozenBalance returns the frozen balance of a denom on an account
func (k Keeper) GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.frozenBalanceStore(ctx, addr).getFrozenBalance(denom)
}

// GetFrozenBalances returns the frozen balance on an account
func (k Keeper) GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.frozenBalanceStore(ctx, addr).getFrozenBalances(pagination)
}

type frozenStore struct {
	store prefix.Store
	cdc   codec.BinaryCodec
}

func (s frozenStore) getFrozenBalance(denom string) sdk.Coin {
	frozenBalance := sdk.NewCoin(denom, sdk.NewInt(0))
	if bz := s.store.Get([]byte(denom)); bz != nil {
		s.cdc.MustUnmarshal(bz, &frozenBalance)
	}

	return frozenBalance
}

func (s frozenStore) getFrozenBalances(pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	coins := sdk.NewCoins()
	pageRes, err := query.Paginate(s.store, pagination, func(key, value []byte) error {
		var coin sdk.Coin
		s.cdc.MustUnmarshal(value, &coin)
		coins = append(coins, coin)
		return nil
	})
	return coins, pageRes, err
}

func (s frozenStore) setFrozenBalance(coin sdk.Coin) {
	if coin.Amount.IsZero() {
		s.store.Delete([]byte(coin.Denom))
	} else {
		bz := s.cdc.MustMarshal(&coin)
		s.store.Set([]byte(coin.Denom), bz)
	}
}

// frozenBalanceStore get the store for the frozen balances of an account
func (k Keeper) frozenBalanceStore(ctx sdk.Context, addr sdk.AccAddress) frozenStore {
	store := ctx.KVStore(k.storeKey)
	return frozenStore{
		store: prefix.NewStore(store, types.CreateFrozenBalancesPrefix(addr)),
		cdc:   k.cdc,
	}
}

// isFeatureEnabled checks weather a feature is present on a list of token features
func isFeatureEnabled(features []types.FungibleTokenFeature, feature types.FungibleTokenFeature) error {
	for _, o := range features {
		if o == feature {
			return nil
		}
	}
	return types.ErrFeatureNotActive
}
