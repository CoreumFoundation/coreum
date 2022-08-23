package keeper

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type Keeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)

	FreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error
	UnfreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error
	ListAccountFrozenCoins(ctx sdk.Context, holder sdk.AccAddress) (sdk.Coins, error)
	ListFrozenCoins(ctx sdk.Context) (map[string]sdk.Coins, error)

	Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error)
}

type BaseKeeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace
	bankKeeper BankKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper BankKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &BaseKeeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		bankKeeper: bankKeeper,
	}
}

func (k BaseKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *BaseKeeper) FreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error {
	balance := k.bankKeeper.GetBalance(ctx, holder, coin.Denom)
	if balance.Amount.IsZero() {
		return fmt.Errorf("the given account does not hold the given coin")
	}

	store := k.getFreezeCoinStore(ctx, holder)

	key := []byte(coin.Denom)

	if store.Has(key) {
		amount := big.NewInt(0).SetBytes(store.Get(key))
		coin.Amount = coin.Amount.Add(sdk.NewIntFromBigInt(amount))
	}

	store.Set(key, coin.Amount.BigInt().Bytes())

	return nil
}

func (k *BaseKeeper) UnfreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error {
	balance := k.bankKeeper.GetBalance(ctx, holder, coin.Denom)
	if balance.Amount.IsZero() {
		return fmt.Errorf("the given account does not hold the given coin")
	}

	store := k.getFreezeCoinStore(ctx, holder)

	key := []byte(coin.Denom)

	if !store.Has(key) {
		return fmt.Errorf("%s is not frozen on the given account", coin)
	}

	amount := sdk.NewIntFromBigInt(big.NewInt(0).SetBytes(store.Get(key)))
	if coin.Amount.GT(amount) {
		return fmt.Errorf("only %s%s is frozen on the given account", amount, key)
	}

	coin.Amount = coin.Amount.Sub(amount)

	store.Set(key, coin.Amount.BigInt().Bytes())

	return nil
}

func (k *BaseKeeper) ListAccountFrozenCoins(ctx sdk.Context, holder sdk.AccAddress) (sdk.Coins, error) {
	store := k.getFreezeCoinStore(ctx, holder)

	coinIter := store.Iterator(nil, nil)
	defer coinIter.Close()

	var frozenCoins sdk.Coins
	for ; coinIter.Valid(); coinIter.Next() {
		denom, amountRaw := coinIter.Key(), coinIter.Value()
		amount := sdk.NewIntFromBigInt(big.NewInt(0).SetBytes(amountRaw))
		frozenCoins = frozenCoins.Add(sdk.NewCoin(string(denom), amount))
	}

	if err := coinIter.Error(); err != nil {
		return nil, err
	}

	return frozenCoins, nil
}

func (k *BaseKeeper) ListFrozenCoins(ctx sdk.Context) (map[string]sdk.Coins, error) {
	baseStore := ctx.KVStore(k.storeKey)

	frozenCoins := make(map[string]sdk.Coins)

	accIter := baseStore.Iterator(nil, nil)
	defer accIter.Close()

	for ; accIter.Valid(); accIter.Next() {
		acc := accIter.Key()
		if !bytes.HasPrefix(acc, types.KeyPrefix(types.FrozenCoinKey)) {
			continue
		}

		store := prefix.NewStore(baseStore, acc)

		coinIter := store.Iterator(nil, nil)
		defer coinIter.Close()

		coins := sdk.NewCoins()
		for ; coinIter.Valid(); coinIter.Next() {
			denom, amountRaw := coinIter.Key(), coinIter.Value()
			amount := sdk.NewIntFromBigInt(big.NewInt(0).SetBytes(amountRaw))
			coins = coins.Add(sdk.NewCoin(string(denom), amount))
		}

		if err := coinIter.Error(); err != nil {
			return nil, err
		}

		frozenCoins[string(acc)] = coins
	}

	if err := accIter.Error(); err != nil {
		return nil, err
	}

	return frozenCoins, nil
}

func (k *BaseKeeper) getFreezeCoinStore(ctx sdk.Context, holder sdk.AccAddress) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.FrozenCoinKey+holder.String()))
}
