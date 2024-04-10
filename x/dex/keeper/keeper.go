package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

type orderWrapper struct {
	Order        types.Order
	orderID      uint64
	isPersistent bool
}

// Keeper is the dex module keeper.
type Keeper struct {
	cdc               codec.BinaryCodec
	storeKey          storetypes.StoreKey
	transientStoreKey storetypes.StoreKey
	accountKeeper     types.AccountKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	transientStoreKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
) Keeper {
	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
		accountKeeper:     accountKeeper,
	}
}

// StoreTransientOrder stores order in the transient queue to be processed by the end blocker.
func (k Keeper) StoreTransientOrder(ctx sdk.Context, order types.Order) error {
	// TODO: Lock funds

	orderBytes, err := k.encodeOrder(order)
	if err != nil {
		return err
	}

	tStore := ctx.TransientStore(k.transientStoreKey)
	tStore.Set(types.CreateOrderKey(k.nextOrderID(ctx)), orderBytes)

	return nil
}

// ProcessTransientQueue processes orders stored in the transient queue and matches them.
func (k Keeper) ProcessTransientQueue(ctx sdk.Context) error {
	iterator := prefix.NewStore(ctx.TransientStore(k.transientStoreKey), types.OrderKey).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		if err := k.processTransientOrder(ctx, iterator); err != nil {
			return err
		}
	}

	return nil
}

// ExportOrders exports persistent orders.
func (k Keeper) ExportOrders(ctx sdk.Context) ([]types.Order, error) {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.OrderKey).Iterator(nil, nil)
	defer iterator.Close()

	orders := []types.Order{}
	for ; iterator.Valid(); iterator.Next() {
		order, err := k.decodeOrder(iterator.Value())
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (k Keeper) processTransientOrder(ctx sdk.Context, iterator storetypes.Iterator) error {
	order, err := k.decodeOrder(iterator.Value())
	if err != nil {
		return err
	}

	denom1Sequence := k.createDenomSequence(ctx, order.DenomOffered())
	denom2Sequence := k.createDenomSequence(ctx, order.DenomRequested())

	store := ctx.KVStore(k.storeKey)
	iteratorA := prefix.NewStore(store,
		types.CreateDenomPairKeyPrefix(types.OrderQueueKey, denom1Sequence, denom2Sequence)).Iterator(nil, nil)
	defer iteratorA.Close()

	iteratorB := prefix.NewStore(store,
		types.CreateDenomPairKeyPrefix(types.OrderQueueKey, denom2Sequence, denom1Sequence)).Iterator(nil, nil)
	defer iteratorB.Close()

	orderID := types.DecomposeOrderKey(iterator.Key())
	wrappedOrder := &orderWrapper{Order: order, orderID: orderID}

	orderA, err := k.loadPersistentOrder(ctx, iteratorA)
	if err != nil {
		return err
	}

	if orderA != nil && orderA.Order.Price().LTE(order.Price()) {
		return k.persistOrder(ctx, wrappedOrder)
	}

	return k.matchOrder(ctx, wrappedOrder, iteratorB)
}

func (k Keeper) matchOrder(ctx sdk.Context, orderA *orderWrapper, iteratorB storetypes.Iterator) error {
	var fullyMatchedA bool
	for {
		orderB, err := k.loadPersistentOrder(ctx, iteratorB)
		if err != nil {
			return err
		}
		if orderB == nil || orderA.Order.Price().GT(sdk.OneDec().Quo(orderB.Order.Price())) {
			break
		}

		amountA := orderA.Order.AmountOffered()
		amountB := amountA.ToLegacyDec().Quo(orderB.Order.Price()).RoundInt()
		if amountB.GT(orderB.Order.AmountOffered()) {
			amountB = orderB.Order.AmountOffered()
			amountA = amountB.ToLegacyDec().Mul(orderB.Order.Price()).RoundInt()
			if amountA.GT(orderA.Order.AmountOffered()) {
				amountA = orderA.Order.AmountOffered()
			}
		}

		// TODO: execute bank transfers

		fullyMatched, err := k.reduceOrder(orderB, amountB)
		if err != nil {
			return err
		}
		if fullyMatched {
			if err := k.dropOrder(ctx, orderB); err != nil {
				return err
			}
		} else {
			if err := k.persistOrder(ctx, orderB); err != nil {
				return err
			}
		}

		fullyMatchedA, err = k.reduceOrder(orderA, amountA)
		if err != nil {
			return err
		}
		if fullyMatchedA {
			return nil
		}
	}

	return k.persistOrder(ctx, orderA)
}

func (k Keeper) reduceOrder(order *orderWrapper, amount sdkmath.Int) (bool, error) {
	order.Order.ReduceOfferedAmount(amount)

	if !order.Order.AmountOffered().ToLegacyDec().Mul(order.Order.Price()).RoundInt().IsZero() {
		return false, nil
	}

	return true, nil
}

func (k Keeper) loadPersistentOrder(
	ctx sdk.Context,
	iterator storetypes.Iterator,
) (*orderWrapper, error) {
	if !iterator.Valid() {
		return nil, nil
	}
	orderID := types.DecomposeOrderQueueKey(iterator.Key())
	iterator.Next()

	return k.orderByID(ctx, orderID)
}

func (k Keeper) persistOrder(ctx sdk.Context, order *orderWrapper) error {
	orderBytes, err := k.encodeOrder(order.Order)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Set(types.CreateOrderKey(order.orderID), orderBytes)

	if order.isPersistent {
		return nil
	}

	account, err := sdk.AccAddressFromBech32(order.Order.Account())
	if err != nil {
		return err
	}
	acc := k.accountKeeper.GetAccount(ctx, account)

	store.Set(types.CreateOrderQueueKey(
		k.createDenomSequence(ctx, order.Order.DenomOffered()),
		k.createDenomSequence(ctx, order.Order.DenomRequested()),
		order.orderID,
		order.Order.Price(),
	), types.StoreTrue)

	store.Set(types.CreateOrderOwnerKey(
		acc.GetAccountNumber(),
		order.orderID,
	), types.StoreTrue)

	return nil
}

func (k Keeper) dropOrder(ctx sdk.Context, order *orderWrapper) error {
	account, err := sdk.AccAddressFromBech32(order.Order.Account())
	if err != nil {
		return err
	}
	acc := k.accountKeeper.GetAccount(ctx, account)

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.CreateOrderKey(order.orderID))
	store.Delete(types.CreateOrderQueueKey(
		k.createDenomSequence(ctx, order.Order.DenomOffered()),
		k.createDenomSequence(ctx, order.Order.DenomRequested()),
		order.orderID,
		order.Order.Price(),
	))
	store.Delete(types.CreateOrderOwnerKey(
		acc.GetAccountNumber(),
		order.orderID,
	))

	return nil
}

func (k Keeper) nextOrderID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	orderID := sdkmath.ZeroUint()
	bz := store.Get(types.OrderLastIDKey)

	if bz != nil {
		if err := orderID.Unmarshal(bz); err != nil {
			panic(err)
		}
	}
	orderID = orderID.Incr()
	bz, err := orderID.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.OrderLastIDKey, bz)

	return orderID.Uint64()
}

func (k Keeper) nextDenomSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	seq := sdkmath.ZeroUint()
	bz := store.Get(types.DenomLastSequenceKey)

	if bz != nil {
		if err := seq.Unmarshal(bz); err != nil {
			panic(err)
		}
	}
	seq = seq.Incr()
	bz, err := seq.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.DenomLastSequenceKey, bz)

	return seq.Uint64()
}

func (k Keeper) createDenomSequence(ctx sdk.Context, denom string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateDenomMappingKey(denom)
	bz := store.Get(key)
	if bz == nil {
		seqRaw := k.nextDenomSequence(ctx)
		seq := sdkmath.NewUint(seqRaw)

		bz, err := seq.Marshal()
		if err != nil {
			panic(err)
		}
		store.Set(key, bz)

		return seqRaw
	}

	seq := sdkmath.ZeroUint()
	if err := seq.Unmarshal(bz); err != nil {
		panic(err)
	}
	return seq.Uint64()
}

func (k Keeper) encodeOrder(order types.Order) ([]byte, error) {
	orderAny, err := codectypes.NewAnyWithValue(order.(*types.OrderLimit))
	if err != nil {
		return nil, err
	}

	bz, err := orderAny.Marshal()
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func (k Keeper) decodeOrder(bz []byte) (types.Order, error) {
	orderAny := &codectypes.Any{}
	if err := k.cdc.Unmarshal(bz, orderAny); err != nil {
		return nil, errors.Wrapf(err, "decoding order failed")
	}

	var order types.Order
	if err := k.cdc.UnpackAny(orderAny, &order); err != nil {
		return nil, errors.Wrapf(err, "unpacking order failed")
	}

	return order, nil
}

func (k Keeper) orderByID(ctx sdk.Context, orderID uint64) (*orderWrapper, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CreateOrderKey(orderID))
	if bz == nil {
		return nil, errors.New("order does not exist")
	}

	order, err := k.decodeOrder(bz)
	if err != nil {
		return nil, err
	}

	return &orderWrapper{Order: order, orderID: orderID, isPersistent: true}, nil
}
