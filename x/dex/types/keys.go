package types

import (
	"encoding/binary"

	"github.com/CoreumFoundation/coreum/v4/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "dex"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// TransientStoreKey defines the transient module store key.
	TransientStoreKey = "transient_" + ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName
)

// Store keys.
var (
	OrderTransientSequenceKey = []byte{0x00}
	OrderTransientQueueKey    = []byte{0x01}
	DenomSequenceKey          = []byte{0x02}
	DenomMappingKey           = []byte{0x03}
)

// CreateDenomMappingKey creates the key for the denom-uint64 mapping.
func CreateDenomMappingKey(denom string) []byte {
	return store.JoinKeys(DenomMappingKey, []byte(denom))
}

// CreateOrderTransientQueueKey creates the key for an order inside transient queue.
func CreateOrderTransientQueueKey(denom1Seq, denom2Seq, orderSeq uint64) []byte {
	key := make([]byte, 0, 3*8)
	binary.BigEndian.AppendUint64(key, denom1Seq)
	binary.BigEndian.AppendUint64(key, denom2Seq)
	binary.BigEndian.AppendUint64(key, orderSeq)

	return store.JoinKeys(OrderTransientQueueKey, key)
}
