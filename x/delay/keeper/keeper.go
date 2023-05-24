package keeper

import (
	"encoding/binary"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// MessageToExecute contains info about delayed message to execute.
type MessageToExecute struct {
	Key     []byte
	Message proto.Message
}

// Keeper is delay module Keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

// DelayMessage stores a message to be executed later.
func (k Keeper) DelayMessage(ctx sdk.Context, id string, msg proto.Message, delay time.Duration) error {
	execTime := ctx.BlockTime().Add(delay).Unix()
	if execTime < 0 {
		panic("there were no blockchains before 1970-01-01")
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

// MessagesToExecute returns delayed messages to be executed at the end of the current block.
func (k Keeper) MessagesToExecute(ctx sdk.Context) ([]MessageToExecute, error) {
	store := ctx.KVStore(k.storeKey)

	// messages will be returned from this iterator in the execution time ascending order
	iter := store.Iterator(nil, nil)

	blockTime := ctx.BlockTime().Unix()
	if blockTime < 0 {
		panic("there were no blockchains before 1970-01-01")
	}
	blockTimeUnsigned := uint64(blockTime)

	msgs := []MessageToExecute{}
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) < 8 {
			panic("key is too short")
		}

		// due to the order of messages returned by the iterator, if we find that execution time of the message is after
		// the current block time, then there is no reason to iterate further
		execTime := binary.BigEndian.Uint64(key[:8])
		if execTime > blockTimeUnsigned {
			return msgs, nil
		}

		var msg sdk.Msg
		if err := k.cdc.UnmarshalInterface(iter.Value(), &msg); err != nil {
			return nil, errors.Wrap(err, "decoding delayed message failed")
		}

		msgs = append(msgs, MessageToExecute{
			Key:     key,
			Message: msg,
		})
	}
	return msgs, nil
}

// DeleteMessage deletes messages from the store.
func (k Keeper) DeleteMessage(ctx sdk.Context, msgs []MessageToExecute) {
	store := ctx.KVStore(k.storeKey)
	for _, m := range msgs {
		store.Delete(m.Key)
	}
}
