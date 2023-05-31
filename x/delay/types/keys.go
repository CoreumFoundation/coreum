package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

const (
	// ModuleName defines the module name.
	ModuleName = "delay"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key.
	RouterKey = ModuleName
)

// CreateDelayedItemKey creates key for delayed item.
func CreateDelayedItemKey(ctx sdk.Context, id string, delay time.Duration) ([]byte, error) {
	execTime := ctx.BlockTime().Add(delay).Unix()
	if execTime < 0 {
		return nil, errors.New("there were no blockchains before 1970-01-01")
	}

	key := make([]byte, 8, 8+len(id))
	// big endian is used to be sure that results are sortable lexicographically when stored messages are iterated
	binary.BigEndian.PutUint64(key, uint64(execTime))

	return append(key, []byte(id)...), nil
}

// ExtractUnixTimestampFromDelayedItemKey extracts from the key the timestamp of delayed message execution.
func ExtractUnixTimestampFromDelayedItemKey(key []byte) (uint64, error) {
	if len(key) < 8 {
		return 0, errors.New("key is too short")
	}

	return binary.BigEndian.Uint64(key[:8]), nil
}
