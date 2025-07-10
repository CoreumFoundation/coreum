package keeper

import (
	sdkstore "cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// OrderBookIterator is order book iterator.
type OrderBookIterator struct {
	cdc          codec.BinaryCodec
	readFromTail bool
	iterator     storetypes.Iterator
	orderBookID  uint32
	side         types.Side
	// pool of orders which were loaded already but not returned yet
	pool []types.OrderBookRecord
}

// NewOrderBookIterator returns new instance of the OrderBookIterator.
func NewOrderBookIterator(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	storeService sdkstore.KVStoreService,
	orderBookID uint32,
	side types.Side,
	readFromTail bool,
) *OrderBookIterator {
	moduleStore := storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.CreateOrderBookSideKey(orderBookID, side))
	var iterator storetypes.Iterator
	if readFromTail {
		iterator = store.ReverseIterator(nil, nil)
	} else {
		iterator = store.Iterator(nil, nil)
	}

	return &OrderBookIterator{
		cdc:          cdc,
		readFromTail: readFromTail,
		iterator:     iterator,
		orderBookID:  orderBookID,
		side:         side,
		pool:         make([]types.OrderBookRecord, 0),
	}
}

// Close closes the iterator, releasing allocated resources.
func (i *OrderBookIterator) Close() error {
	return i.iterator.Close()
}

// Next returns the order book record with the lowest price and lowest order sequence (for the same price) if read forms
// head, or order book record with the highest price and lowest order sequence (for the same price) if reads from tail.
func (i *OrderBookIterator) Next() (types.OrderBookRecord, bool, error) {
	if i.readFromTail {
		return i.nextFromTail()
	}
	return i.nextFromHead()
}

// nextFromTail returns next from the tail order book record sorted by price descending, and order sequence ascending.
func (i *OrderBookIterator) nextFromTail() (types.OrderBookRecord, bool, error) {
	if len(i.pool) > 1 {
		return i.removeAndReturnCurrentTailRecord()
	}
	// load records to the pool, we load at least 2 until next price is different from current
	for {
		r, found, err := i.nextFromIterator()
		if err != nil {
			return types.OrderBookRecord{}, false, err
		}
		if !found {
			break
		}
		i.pool = append(i.pool, r)
		if len(i.pool) < 2 {
			continue
		}
		// check if records have same price
		if i.pool[len(i.pool)-2].Price.Rat().Cmp(i.pool[len(i.pool)-1].Price.Rat()) == 0 {
			continue
		}
		break
	}
	if len(i.pool) > 1 {
		return i.removeAndReturnCurrentTailRecord()
	}
	if len(i.pool) == 1 {
		r := i.pool[0]
		i.pool = make([]types.OrderBookRecord, 0)
		return r, true, nil
	}

	return types.OrderBookRecord{}, false, nil
}

func (i *OrderBookIterator) removeAndReturnCurrentTailRecord() (types.OrderBookRecord, bool, error) {
	lastIndex := len(i.pool) - 1
	prevFromLastIndex := len(i.pool) - 2

	lastRecord := i.pool[lastIndex]
	prevFromLastRecord := i.pool[prevFromLastIndex]

	var (
		indexToRemove  int
		recordToReturn types.OrderBookRecord
	)
	// if prices are the same we have reached the end with the same price
	if lastRecord.Price.Rat().Cmp(prevFromLastRecord.Price.Rat()) == 0 {
		indexToRemove = lastIndex
		recordToReturn = lastRecord
	} else {
		indexToRemove = prevFromLastIndex
		recordToReturn = prevFromLastRecord
	}

	i.pool = append(i.pool[:indexToRemove], i.pool[indexToRemove+1:]...)
	return recordToReturn, true, nil
}

// nextFromTail returns next from the head order book record sorted by price ascending, and order sequence ascending.
func (i *OrderBookIterator) nextFromHead() (types.OrderBookRecord, bool, error) {
	return i.nextFromIterator()
}

func (i *OrderBookIterator) nextFromIterator() (types.OrderBookRecord, bool, error) {
	if !i.iterator.Valid() {
		return types.OrderBookRecord{}, false, nil
	}

	r, err := i.readOrderBookRecordFromIterator()
	if err != nil {
		return types.OrderBookRecord{}, false, err
	}

	i.iterator.Next()

	return r, true, nil
}

func (i *OrderBookIterator) readOrderBookRecordFromIterator() (types.OrderBookRecord, error) {
	// decode key to values
	price, orderSequence, err := types.DecodeOrderBookSideRecordKey(i.iterator.Key())
	if err != nil {
		return types.OrderBookRecord{}, err
	}
	// decode value
	var storedRecord types.OrderBookRecordData
	if err := i.cdc.Unmarshal(i.iterator.Value(), &storedRecord); err != nil {
		return types.OrderBookRecord{},
			sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal OrderBookRecordData, err: %s", err)
	}

	return types.OrderBookRecord{
		// key attributes
		OrderBookID:   i.orderBookID,
		Side:          i.side,
		Price:         price,
		OrderSequence: orderSequence,
		// value attributes
		OrderID:                   storedRecord.OrderID,
		AccountNumber:             storedRecord.AccountNumber,
		RemainingBaseQuantity:     storedRecord.RemainingBaseQuantity,
		RemainingSpendableBalance: storedRecord.RemainingSpendableBalance,
	}, nil
}

// NewOrderBookSideIterator returns order book iterator with the reading based on side (buy - tail, sell head).
func (k Keeper) NewOrderBookSideIterator(ctx sdk.Context, orderBookID uint32, side types.Side) *OrderBookIterator {
	readFromTail := side == types.SIDE_BUY

	return NewOrderBookIterator(ctx, k.cdc, k.storeService, orderBookID, side, readFromTail)
}
