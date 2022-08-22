package keeper

import (
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

type Keeper interface {
	Logger(ctx sdk.Context) log.Logger

	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params)

	FreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin)
	UnfreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error
	GetFrozenCoin(ctx sdk.Context, holder sdk.AccAddress, denom string) sdk.Coin
	ListAccountFrozenCoins(ctx sdk.Context, holder sdk.AccAddress) (sdk.Coins, error)
	ListFrozenCoins(ctx sdk.Context) (map[string]sdk.Coins, error)

	Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error)
}

type BaseKeeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	memKey     sdk.StoreKey
	paramstore paramtypes.Subspace
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
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
	}
}

func (k BaseKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *BaseKeeper) FreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

	key := []byte(coin.Denom)

	if store.Has(key) {
		amount := big.NewInt(0).SetBytes(store.Get(key))
		coin.Amount = coin.Amount.Add(sdk.NewIntFromBigInt(amount))
	}

	store.Set(key, coin.Amount.BigInt().Bytes())
}

func (k *BaseKeeper) UnfreezeCoin(ctx sdk.Context, holder sdk.AccAddress, coin sdk.Coin) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

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

func (k *BaseKeeper) GetFrozenCoin(ctx sdk.Context, holder sdk.AccAddress, denom string) sdk.Coin {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

	key := []byte(denom)

	if !store.Has(key) {
		return sdk.NewInt64Coin(denom, 0)
	}

	amount := sdk.NewIntFromBigInt(big.NewInt(0).SetBytes(store.Get(key)))
	return sdk.NewCoin(denom, amount)
}

func (k *BaseKeeper) ListAccountFrozenCoins(ctx sdk.Context, holder sdk.AccAddress) (sdk.Coins, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

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
	frozenCoins := make(map[string]sdk.Coins)

	accIter := ctx.KVStore(k.storeKey).Iterator(nil, nil)
	defer accIter.Close()

	for ; accIter.Valid(); accIter.Next() {
		acc := accIter.Key()
		store := prefix.NewStore(ctx.KVStore(k.storeKey), acc)

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
