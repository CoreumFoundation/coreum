package keeper

import (
	"bytes"
	"sort"

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
	types.Order
	orderID      uint64
	isPersistent bool
	isDirty      bool
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
	tStore.Set(types.CreateOrderTransientQueueKey(
		k.createDenomSequence(ctx, order.DenomOffered()),
		k.createDenomSequence(ctx, order.DenomRequested()),
		k.nextOrderID(ctx),
	), orderBytes)

	return nil
}

// ProcessTransientQueue processes orders stored in the transient queue and matches them.
func (k Keeper) ProcessTransientQueue(ctx sdk.Context) error {
	iterator := prefix.NewStore(ctx.TransientStore(k.transientStoreKey), types.OrderTransientQueueKey).Iterator(nil, nil)
	defer iterator.Close()

	for iterator.Valid() {
		if err := k.processTransientOrderBook(ctx, iterator); err != nil {
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

func (k Keeper) processTransientOrderBook(ctx sdk.Context, iterator storetypes.Iterator) error {
	denom1Sequence, denom2Sequence, _ := types.DecomposeOrderTransientQueueKey(iterator.Key())

	store := ctx.KVStore(k.storeKey)
	persistentIteratorA := prefix.NewStore(store,
		types.CreateDenomPairKeyPrefix(types.OrderQueueKey, denom1Sequence, denom2Sequence)).Iterator(nil, nil)
	defer persistentIteratorA.Close()

	persistentIteratorB := prefix.NewStore(store,
		types.CreateDenomPairKeyPrefix(types.OrderQueueKey, denom2Sequence, denom1Sequence)).Iterator(nil, nil)
	defer persistentIteratorB.Close()

	sideA, err := k.loadPersistentOrder(ctx, []*orderWrapper{}, persistentIteratorA)
	if err != nil {
		return err
	}
	sideB, err := k.loadPersistentOrder(ctx, []*orderWrapper{}, persistentIteratorB)
	if err != nil {
		return err
	}

	var orderBookPrefix []byte
loop:
	for ; iterator.Valid(); iterator.Next() {
		prefix := iterator.Key()[:16]
		switch {
		case orderBookPrefix == nil:
			orderBookPrefix = prefix
		case !bytes.Equal(prefix, orderBookPrefix):
			break loop
		}

		order, err := k.decodeOrder(iterator.Value())
		if err != nil {
			return err
		}

		_, _, orderID := types.DecomposeOrderTransientQueueKey(iterator.Key())
		offeredDenomSequence := k.createDenomSequence(ctx, order.DenomOffered())

		if denom1Sequence == offeredDenomSequence {
			sideA = k.appendOrder(sideA, order, orderID)
			var err error
			sideA, sideB, err = k.matchOrder(ctx, orderID, sideA, sideB, persistentIteratorA, persistentIteratorB)
			if err != nil {
				return err
			}
		} else {
			sideB = k.appendOrder(sideB, order, orderID)
			var err error
			sideB, sideA, err = k.matchOrder(ctx, orderID, sideB, sideA, persistentIteratorB, persistentIteratorA)
			if err != nil {
				return err
			}
		}
	}

	if err := k.persistOrders(ctx, sideA); err != nil {
		return err
	}
	return k.persistOrders(ctx, sideB)
}

func (k Keeper) appendOrder(side []*orderWrapper, order types.Order, orderID uint64) []*orderWrapper {
	// TODO: Because `side` is always sorted, the algorithm below might be optimised to put the new element
	// at the specific index, which would give the complexity of O(n) instead of O(nlogn).
	side = append(side, &orderWrapper{Order: order, orderID: orderID, isDirty: true})
	if len(side) > 1 {
		sort.Slice(side, func(i, j int) bool {
			oA, oB := side[i], side[j]
			return oA.Price().LT(oB.Price()) || (oA.Price().Equal(oB.Price()) && oA.orderID > oB.orderID)
		})
	}
	return side
}

func (k Keeper) matchOrder(
	ctx sdk.Context,
	orderID uint64,
	sideA, sideB []*orderWrapper,
	persistentIteratorA, persistentIteratorB storetypes.Iterator,
) ([]*orderWrapper, []*orderWrapper, error) {
	orderA := sideA[len(sideA)-1]
	if orderA.orderID != orderID {
		return sideA, sideB, nil
	}
	for {
		if len(sideB) == 0 {
			return sideA, sideB, nil
		}
		orderB := sideB[len(sideB)-1]
		if orderA.Price().LT(sdk.OneDec().Quo(orderB.Price())) {
			return sideA, sideB, nil
		}

		amountA := orderA.AmountOffered()
		amountB := amountA.ToLegacyDec().Quo(orderB.Price()).RoundInt()
		if amountB.GT(orderB.AmountOffered()) {
			amountB = orderB.AmountOffered()
			amountA = amountB.ToLegacyDec().Mul(orderB.Price()).RoundInt()
			if amountA.GT(orderA.AmountOffered()) {
				amountA = orderA.AmountOffered()
			}
		}

		var err error
		var fullyMatched bool
		sideA, fullyMatched, err = k.reduceOrder(ctx, orderA, amountA, sideA, persistentIteratorA)
		if err != nil {
			return nil, nil, err
		}
		sideB, _, err = k.reduceOrder(ctx, orderB, amountB, sideB, persistentIteratorB)
		if err != nil {
			return nil, nil, err
		}

		// TODO: execute bank transfers

		if fullyMatched {
			return sideA, sideB, nil
		}
	}
}

func (k Keeper) reduceOrder(
	ctx sdk.Context,
	order *orderWrapper,
	amount sdkmath.Int,
	side []*orderWrapper,
	persistentIterator storetypes.Iterator,
) ([]*orderWrapper, bool, error) {
	order.ReduceOfferedAmount(amount)
	order.isDirty = true

	if !order.AmountOffered().ToLegacyDec().Mul(order.Price()).RoundInt().IsZero() {
		return side, false, nil
	}

	side = side[:len(side)-1]

	if !order.isPersistent {
		return side, true, nil
	}

	if err := k.dropOrder(ctx, order); err != nil {
		return nil, false, err
	}

	side, err := k.loadPersistentOrder(ctx, side, persistentIterator)
	if err != nil {
		return nil, false, err
	}
	return side, true, nil
}

func (k Keeper) loadPersistentOrder(
	ctx sdk.Context,
	side []*orderWrapper,
	iterator storetypes.Iterator,
) ([]*orderWrapper, error) {
	if !iterator.Valid() {
		return side, nil
	}
	orderID := types.DecomposeOrderQueueKey(iterator.Key())
	iterator.Next()

	order, err := k.orderByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return k.appendOrder(side, order, orderID), nil
}

func (k Keeper) persistOrders(ctx sdk.Context, orders []*orderWrapper) error {
	store := ctx.KVStore(k.storeKey)

	for _, order := range orders {
		if !order.isDirty {
			continue
		}

		orderBytes, err := k.encodeOrder(order.Order)
		if err != nil {
			return err
		}

		store.Set(types.CreateOrderKey(order.orderID), orderBytes)

		if order.isPersistent {
			continue
		}

		account, err := sdk.AccAddressFromBech32(order.Account())
		if err != nil {
			return err
		}
		acc := k.accountKeeper.GetAccount(ctx, account)

		store.Set(types.CreateOrderQueueKey(
			k.createDenomSequence(ctx, order.DenomOffered()),
			k.createDenomSequence(ctx, order.DenomRequested()),
			order.orderID,
			order.Price(),
		), types.StoreTrue)

		store.Set(types.CreateOrderOwnerKey(
			acc.GetAccountNumber(),
			order.orderID,
		), types.StoreTrue)
	}

	return nil
}

func (k Keeper) dropOrder(ctx sdk.Context, order *orderWrapper) error {
	account, err := sdk.AccAddressFromBech32(order.Account())
	if err != nil {
		return err
	}
	acc := k.accountKeeper.GetAccount(ctx, account)

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.CreateOrderKey(order.orderID))
	store.Delete(types.CreateOrderQueueKey(
		k.createDenomSequence(ctx, order.DenomOffered()),
		k.createDenomSequence(ctx, order.DenomRequested()),
		order.orderID,
		order.Price(),
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
	orderAny, err := codectypes.NewAnyWithValue(order)
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

	return &orderWrapper{Order: order, isPersistent: true}, nil
}
