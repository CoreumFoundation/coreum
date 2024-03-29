package types

import (
	"encoding/binary"

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
	OrderLastIDKey            = []byte{0x00}
	DenomLastSequenceKey      = []byte{0x01}
	DenomToSequenceMappingKey = []byte{0x02}
	OrderQueueKey             = []byte{0x03}
	OrderOwnerKey             = []byte{0x04}
	OrderKey                  = []byte{0x05}
)

// StoreTrue keeps a value used by stores to indicate that key is present.
var StoreTrue = []byte{0x01}

// CreateDenomPairKeyPrefix creates the key prefix from two denom sequence numbers.
func CreateDenomPairKeyPrefix(prefix []byte, denom1Seq, denom2Seq uint64) []byte {
	key := make([]byte, 0, 2*uint64ByteSize)
	key = binary.BigEndian.AppendUint64(key, denom1Seq)
	key = binary.BigEndian.AppendUint64(key, denom2Seq)
	return store.JoinKeys(prefix, key)
}

// CreateDenomMappingKey creates the key for the denom-uint64 mapping.
func CreateDenomMappingKey(denom string) []byte {
	return store.JoinKeys(DenomToSequenceMappingKey, []byte(denom))
}

// CreateOrderQueueKey creates the key for an order inside transient queue.
func CreateOrderQueueKey(denomOfferedSeq, denomRequestedSeq, orderID uint64, price sdk.Dec) []byte {
	wholePart, decPart := priceToUint64s(price)

	key := make([]byte, 0, 3*uint64ByteSize)
	key = binary.BigEndian.AppendUint64(key, wholePart)
	key = binary.BigEndian.AppendUint64(key, decPart)
	key = binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(CreateDenomPairKeyPrefix(OrderQueueKey, denomOfferedSeq, denomRequestedSeq), key)
}

// DecomposeOrderQueueKey decomposes order queue key.
func DecomposeOrderQueueKey(key []byte) uint64 {
	return binary.BigEndian.Uint64(key[2*uint64ByteSize:])
}

// CreateOrderOwnerKey creates the key for an order assigned to an account.
func CreateOrderOwnerKey(accountNumber, orderID uint64) []byte {
	key := make([]byte, 0, 2*uint64ByteSize)
	key = binary.BigEndian.AppendUint64(key, accountNumber)
	key = binary.BigEndian.AppendUint64(key, orderID)

	return store.JoinKeys(OrderOwnerKey, key)
}

// CreateOrderKey creates the key for an order.
func CreateOrderKey(orderID uint64) []byte {
	return store.JoinKeys(OrderKey, binary.BigEndian.AppendUint64(make([]byte, 0, uint64ByteSize), orderID))
}

// DecomposeOrderKey decomposes order key.
func DecomposeOrderKey(key []byte) uint64 {
	return binary.BigEndian.Uint64(key)
}

func priceToUint64s(price sdk.Dec) (uint64, uint64) {
	wholePart := price.TruncateInt()
	decPart := price.Sub(wholePart.ToLegacyDec()).Mul(sdk.NewDecFromInt(sdk.NewIntFromUint64(1000000000000000000))).
		TruncateInt()

	return wholePart.Uint64(), decPart.Uint64()
}
