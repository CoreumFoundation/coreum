package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/delay/types"
)

// Keeper is delay module Keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
	router   types.Router
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	router types.Router,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		router:   router,
	}
}

// DelayExecution stores an item to be executed later.
func (k Keeper) DelayExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, delay time.Duration) error {
	key, err := types.CreateDelayedItemKey(ctx, id, delay)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return errors.Errorf("delayed item is already stored under the key, id: %s", id)
	}

	b, err := k.cdc.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "marshaling delayed item failed")
	}
	store.Set(key, b)
	return nil
}

// ExecuteDelayedItems executes delayed logic.
func (k Keeper) ExecuteDelayedItems(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// messages will be returned from this iterator in the execution time ascending order
	iter := store.Iterator(nil, nil)

	blockTime := ctx.BlockTime().Unix()
	if blockTime < 0 {
		return errors.New("there were no blockchains before 1970-01-01")
	}
	blockTimeUnsigned := uint64(blockTime)

	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 8 {
			return errors.New("key is too short")
		}

		execTime, err := types.ExtractUnixTimestampFromDelayedItemKey(key)
		if err != nil {
			return err
		}

		// due to the order of items returned by the iterator, if we find that execution time is after
		// the current block time, then there is no reason to iterate further
		if execTime > blockTimeUnsigned {
			return nil
		}

		var data codec.ProtoMarshaler
		if err := k.cdc.Unmarshal(iter.Value(), data); err != nil {
			return errors.Wrap(err, "decoding delayed message failed")
		}

		handler, err := k.router.Handler(data)
		if err != nil {
			return err
		}
		if handler == nil {
			return errors.Errorf("no handler for %s found", proto.MessageName(data))
		}
		if err := handler(ctx, data); err != nil {
			return err
		}

		store.Delete(key)
	}
	return nil
}
