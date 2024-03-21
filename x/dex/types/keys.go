package types

import (
	"encoding/binary"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

const uint64ByteSize = 8

// Store keys.
var (
	OrderSequenceKey       = []byte{0x00}
	OrderTransientQueueKey = []byte{0x01}
	DenomSequenceKey       = []byte{0x02}
	DenomMappingKey        = []byte{0x03}
	OrderQueueKey          = []byte{0x04}
	OrderOwnerKey          = []byte{0x05}
	OrderKey               = []byte{0x06}
)

// StoreTrue keeps a value used by stores to indicate that key is present.
var StoreTrue = []byte{0x01}

// CreateDenomPairKeyPrefix creates the key prefix from two denom sequence numbers.
func CreateDenomPairKeyPrefix(prefix []byte, denom1Seq, denom2Seq uint64) []byte {
	key := make([]byte, 0, 2*uint64ByteSize)
	binary.BigEndian.AppendUint64(key, denom1Seq)
	binary.BigEndian.AppendUint64(key, denom2Seq)
	return store.JoinKeys(prefix, key)
}

// CreateDenomMappingKey creates the key for the denom-uint64 mapping.
func CreateDenomMappingKey(denom string) []byte {
	return store.JoinKeys(DenomMappingKey, []byte(denom))
}

// CreateOrderTransientQueueKey creates the key for an order inside transient queue.
func CreateOrderTransientQueueKey(denom1Seq, denom2Seq, orderID uint64) []byte {
	if denom1Seq > denom2Seq {
		denom1Seq, denom2Seq = denom2Seq, denom1Seq
	}
	key := make([]byte, 0, uint64ByteSize)
	binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(CreateDenomPairKeyPrefix(OrderTransientQueueKey, denom1Seq, denom2Seq), key)
}

// DecomposeOrderTransientQueueKey decomposes transient order key.
func DecomposeOrderTransientQueueKey(key []byte) (uint64, uint64, uint64) {
	return binary.BigEndian.Uint64(key[:uint64ByteSize]),
		binary.BigEndian.Uint64(key[uint64ByteSize : 2*uint64ByteSize]),
		binary.BigEndian.Uint64(key[2*uint64ByteSize:])
}

// CreateOrderQueueKey creates the key for an order inside transient queue.
func CreateOrderQueueKey(denomOfferedSeq, denomRequestedSeq, orderID uint64, price sdk.Dec) []byte {

	wholePart := price.TruncateInt()
	decPart := price.Sub(wholePart.ToLegacyDec()).Mul(sdk.NewDecFromInt(sdk.NewIntFromUint64(1000000000000000000))).TruncateInt()

	key := make([]byte, 0, 3*uint64ByteSize)
	binary.BigEndian.AppendUint64(key, math.MaxUint64-wholePart.Uint64())
	binary.BigEndian.AppendUint64(key, math.MaxUint64-decPart.Uint64())
	binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(CreateDenomPairKeyPrefix(OrderQueueKey, denomOfferedSeq, denomRequestedSeq), key)
}

// DecomposeOrderQueueKey decomposes order queue key.
func DecomposeOrderQueueKey(key []byte) uint64 {
	return binary.BigEndian.Uint64(key[2*uint64ByteSize:])
}

// CreateOrderOwnerKey creates the key for an order assigned to an account.
func CreateOrderOwnerKey(accountNumber, orderID uint64) []byte {
	key := make([]byte, 0, 2*uint64ByteSize)
	binary.BigEndian.AppendUint64(key, accountNumber)
	binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(OrderOwnerKey, key)
}

// CreateOrderKey creates the key for an order.
func CreateOrderKey(orderID uint64) []byte {
	key := make([]byte, 0, uint64ByteSize)
	binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(OrderOwnerKey, key)
}
