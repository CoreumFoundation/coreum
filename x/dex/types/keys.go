package types

import (
	"github.com/CoreumFoundation/coreum/v4/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "dex"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey is the message route for slashing.
	RouterKey = ModuleName
)

// Store key prefixes.
var (
	// OrderBookKeyPrefix defines the key prefix for the order book.
	OrderBookKeyPrefix = []byte{0x01}
)

// CreateOrderBookRecordKey creates order book key record with fixed key length to support the correct ordering
// and be able to decode the key into the values.
func CreateOrderBookRecordKey(pairID uint64, side Side, price Price, orderSeq uint64) ([]byte, error) {
	key := CreateOrderBookSideKey(pairID, side)
	var err error
	key, err = CreateOrderBookSideRecordKey(key, price, orderSeq)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// CreateOrderBookSideRecordKey creates order book side record key.
func CreateOrderBookSideRecordKey(key []byte, price Price, orderSeq uint64) ([]byte, error) {
	priceKey, err := price.MarshallToOrderedBytes()
	if err != nil {
		return nil, err
	}
	key = store.JoinKeys(key, priceKey)
	key = store.AppendUint64ToOrderedBytes(key, orderSeq)

	return key, nil
}

// DecodeOrderBookSideRecordKey decodes order book side record key into values.
func DecodeOrderBookSideRecordKey(key []byte) (Price, uint64, error) {
	var p Price
	nextKeyPart, err := p.UnmarshallFromOrderedBytes(key)
	if err != nil {
		return Price{}, 0, err
	}
	orderSeq, _, err := store.ReadOrderedBytesToUint64(nextKeyPart)
	if err != nil {
		return Price{}, 0, err
	}

	return p, orderSeq, nil
}

// CreateOrderBookSideKey creates order book side key.
func CreateOrderBookSideKey(pairID uint64, side Side) []byte {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, pairID)
	key = store.AppendUint8ToOrderedBytes(key, uint8(side))

	return store.JoinKeys(OrderBookKeyPrefix, key)
}
