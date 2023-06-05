package types

import (
	"encoding/binary"
	"time"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/store"
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

const uint64Length = 8

// CreateDelayedItemKey creates key for delayed item.
func CreateDelayedItemKey(id string, t time.Time) ([]byte, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	execTime := t.Unix()
	if execTime < 0 {
		return nil, errors.New("there were no blockchains before 1970-01-01")
	}

	key := make([]byte, uint64Length)
	// big endian is used to be sure that results are sortable lexicographically when stored messages are iterated
	binary.BigEndian.PutUint64(key, uint64(execTime))

	return store.JoinKeys(DelayedItemKeyPrefix, key, []byte(id)), nil
}

// ExtractTimeAndIDFromDelayedItemKey extracts from the key the timestamp and ID of delayed message execution.
func ExtractTimeAndIDFromDelayedItemKey(key []byte) (time.Time, string, error) {
	if len(key) < uint64Length+1 {
		return time.Time{}, "", errors.New("key is too short")
	}

	return time.Unix(int64(binary.BigEndian.Uint64(key[:uint64Length])), 0).UTC(), string(key[uint64Length:]), nil
}
