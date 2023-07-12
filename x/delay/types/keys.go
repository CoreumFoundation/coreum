package types

import (
	"encoding/binary"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/v2/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "delay"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key.
	RouterKey = ModuleName
)

// DelayedItemKeyPrefix defines the key prefix for the delayed item.
var DelayedItemKeyPrefix = []byte{0x01}

const timestampLength = 8

// CreateDelayedItemKey creates key for delayed item.
func CreateDelayedItemKey(id string, t time.Time) ([]byte, error) {
	if id == "" {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "id cannot be empty")
	}

	execTime := t.Unix()
	if execTime < 0 {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "unix timestamp of the execution time must be non-negative")
	}

	key := make([]byte, timestampLength)
	// big endian is used to be sure that results are sortable lexicographically when stored messages are iterated
	binary.BigEndian.PutUint64(key, uint64(execTime))

	return store.JoinKeys(DelayedItemKeyPrefix, key, []byte(id)), nil
}

// ExtractTimeAndIDFromDelayedItemKey extracts from the key the timestamp and ID of delayed message execution.
func ExtractTimeAndIDFromDelayedItemKey(key []byte) (time.Time, string, error) {
	if len(key) < timestampLength+1 {
		return time.Time{}, "", sdkerrors.Wrap(ErrInvalidInput, "key is too short")
	}

	return time.Unix(int64(binary.BigEndian.Uint64(key[:timestampLength])), 0).UTC(), string(key[timestampLength:]), nil
}
