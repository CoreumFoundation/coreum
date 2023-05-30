package keeper

import (
	"encoding/binary"
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

// DelayMessage stores a message to be executed later.
func (k Keeper) DelayMessage(ctx sdk.Context, id string, msg proto.Message, delay time.Duration) error {
	execTime := ctx.BlockTime().Add(delay).Unix()
	if execTime < 0 {
		return errors.New("there were no blockchains before 1970-01-01")
	}

	key := make([]byte, 8, 8+len(id))
	// big endian is used to be sure that results are sortable lexicographically when stored messages are iterated
	binary.BigEndian.PutUint64(key, uint64(execTime))
	key = append(key, []byte(id)...)

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return errors.Errorf("delayed message is already stored under the key, id: %s", id)
	}

	b, err := k.cdc.MarshalInterface(msg)
	if err != nil {
		return errors.Wrap(err, "marshaling delayed message failed")
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

		// due to the order of items returned by the iterator, if we find that execution time is after
		// the current block time, then there is no reason to iterate further
		execTime := binary.BigEndian.Uint64(key[:8])
		if execTime > blockTimeUnsigned {
			return nil
		}

		var data proto.Message
		if err := k.cdc.UnmarshalInterface(iter.Value(), &data); err != nil {
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
