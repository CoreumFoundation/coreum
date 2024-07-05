package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// Keeper is the dex module keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// SaveOrderBookRecord saves order book record to the store.
func (k Keeper) SaveOrderBookRecord(ctx sdk.Context, record types.OrderBookRecord) error {
	// TODO(dzmitryhil) don't forget to add unspecified side validation
	key, err := types.CreateOrderBookRecordKey(record.PairID, record.Side, record.Price, record.OrderSeq)
	if err != nil {
		return err
	}
	storeRecord := types.OrderBookStoreRecord{
		OrderID:           record.OrderID,
		AccountID:         record.AccountID,
		RemainingQuantity: record.RemainingQuantity,
		RemainingBalance:  record.RemainingBalance,
	}
	ctx.KVStore(k.storeKey).Set(key, k.cdc.MustMarshal(&storeRecord))

	return nil
}

// IterateOrderBook iterates an order book.
func (k Keeper) IterateOrderBook(
	ctx sdk.Context,
	pairID uint64,
	side types.Side,
	reverse bool,
	cb func(types.OrderBookRecord) (bool, error),
) error {
	var iterator storetypes.Iterator
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateOrderBookSideKey(pairID, side))
	if reverse {
		iterator = store.ReverseIterator(nil, nil)
	} else {
		iterator = store.Iterator(nil, nil)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var storedRecord types.OrderBookStoreRecord
		k.cdc.MustUnmarshal(iterator.Value(), &storedRecord)

		price, orderSeq, err := types.DecodeOrderBookSideRecordKey(iterator.Key())
		if err != nil {
			return err
		}

		stop, err := cb(types.OrderBookRecord{
			// key attributes
			PairID:   pairID,
			Side:     side,
			Price:    price,
			OrderSeq: orderSeq,
			// value attributes
			OrderID:           storedRecord.OrderID,
			AccountID:         storedRecord.AccountID,
			RemainingQuantity: storedRecord.RemainingQuantity,
			RemainingBalance:  storedRecord.RemainingBalance,
		})
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return nil
}
