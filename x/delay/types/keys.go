package types

import (
	"time"

	sdkerrors "cosmossdk.io/errors"

	"github.com/CoreumFoundation/coreum/v4/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "delay"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key.
	RouterKey = ModuleName
)

var (
	// DelayedItemKeyPrefix defines the key prefix for the delayed item.
	DelayedItemKeyPrefix = []byte{0x01}
	// BlockItemKeyPrefix defines the key prefix for the block item.
	BlockItemKeyPrefix = []byte{0x02}
)

// CreateDelayedItemKey creates key for delayed item.
func CreateDelayedItemKey(id string, t time.Time) ([]byte, error) {
	if id == "" {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "id cannot be empty")
	}

	execTime := t.Unix()
	if execTime < 0 {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "unix timestamp of the execution time must be non-negative")
	}

	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, uint64(execTime))
	return store.JoinKeys(DelayedItemKeyPrefix, key, []byte(id)), nil
}

// DecodeDelayedItemKey extracts from the key the timestamp and ID of delayed message execution.
func DecodeDelayedItemKey(key []byte) (time.Time, string, error) {
	if len(key) < store.Uint64OrderedBytesSize+1 {
		return time.Time{}, "", sdkerrors.Wrap(ErrInvalidInput, "key is too short")
	}
	// first part is the timestamp, the rest is ID
	execTime, id, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return time.Time{}, "", sdkerrors.Wrapf(ErrInvalidInput, "invalid key, err:%s", err.Error())
	}
	return time.Unix(int64(execTime), 0).UTC(), string(id), nil
}

// CreateBlockItemKey creates key for block item.
func CreateBlockItemKey(id string, height uint64) ([]byte, error) {
	if id == "" {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "id cannot be empty")
	}

	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, height)
	return store.JoinKeys(BlockItemKeyPrefix, key, []byte(id)), nil
}

// DecodeBlockItemKey extracts from the key the height and ID of the message execution.
func DecodeBlockItemKey(key []byte) (uint64, string, error) {
	if len(key) < store.Uint64OrderedBytesSize+1 {
		return 0, "", sdkerrors.Wrap(ErrInvalidInput, "key is too short")
	}
	// first part is the timestamp, the rest is ID
	height, id, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, "", sdkerrors.Wrapf(ErrInvalidInput, "invalid key, err:%s", err.Error())
	}
	return height, string(id), nil
}
