package types

import (
	"crypto/sha256"
	"fmt"

	sdkerrors "cosmossdk.io/errors"

	"github.com/CoreumFoundation/coreum/v5/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "dex"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName
)

// Store key prefixes.
var (
	// OrderBookKeyPrefix defines the key prefix for the order book.
	OrderBookKeyPrefix = []byte{0x01}
	// OrderBookSequenceKey defines the key for the order book sequence.
	OrderBookSequenceKey = []byte{0x02}
	// OrderBookDataKeyPrefix defines the key prefix for the order book data.
	OrderBookDataKeyPrefix = []byte{0x03}
	// OrderSequenceKey defines the key for the order sequence.
	OrderSequenceKey = []byte{0x04}
	// OrderKeyPrefix defines the key prefix for the order.
	OrderKeyPrefix = []byte{0x05}
	// OrderIDToSequenceKeyPrefix defines the key prefix for the order ID to sequence.
	OrderIDToSequenceKeyPrefix = []byte{0x06}
	// OrderBookRecordKeyPrefix defines the key prefix for the order book record.
	OrderBookRecordKeyPrefix = []byte{0x07}
	// ParamsKey defines the key to store parameters of the module, set via governance.
	ParamsKey = []byte{0x08}
	// AccountDenomOrdersCountKeyPrefix defines the key prefix for the account denom orders count.
	AccountDenomOrdersCountKeyPrefix = []byte{0x09}
	// AccountDenomOrderSequenceKeyPrefix defines the key prefix for the account denom order sequence.
	AccountDenomOrderSequenceKeyPrefix = []byte{0x10}
)

// CreateOrderBookKey creates order book key.
func CreateOrderBookKey(baseDenom, quoteDenom string) ([]byte, error) {
	// join with length here to prevent the issue described in the `JoinKeysWithLength` comment.
	denomsKey, err := store.JoinKeysWithLength([]byte(baseDenom), []byte(quoteDenom))
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidKey, "failed to join keys, err: %s", err)
	}
	hash := sha256.New()
	_, err = hash.Write(denomsKey)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidKey, "failed write denoms hash, err: %s", err)
	}

	return store.JoinKeys(OrderBookKeyPrefix, hash.Sum(nil)), nil
}

// CreateOrderBookDataKey creates order book data key.
func CreateOrderBookDataKey(orderBookID uint32) []byte {
	key := make([]byte, 0)
	key = store.AppendUint32ToOrderedBytes(key, orderBookID)
	return store.JoinKeys(OrderBookDataKeyPrefix, key)
}

// DecodeOrderBookDataKey decodes order book data key and returns the order book ID.
func DecodeOrderBookDataKey(key []byte) (uint32, error) {
	orderBookID, _, err := store.ReadOrderedBytesToUint32(key)
	if err != nil {
		return 0, err
	}
	return orderBookID, nil
}

// CreateOrderKey creates order key.
func CreateOrderKey(orderSequence uint64) []byte {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, orderSequence)
	return store.JoinKeys(OrderKeyPrefix, key)
}

// CreateOrderIDToSequenceKey creates order ID to sequence key.
func CreateOrderIDToSequenceKey(accountNumber uint64, orderID string) []byte {
	return store.JoinKeys(CreateOrderIDToSequenceKeyPrefix(accountNumber), []byte(orderID))
}

// CreateOrderIDToSequenceKeyPrefix creates order ID to sequence key prefix.
func CreateOrderIDToSequenceKeyPrefix(accountNumber uint64) []byte {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, accountNumber)
	return store.JoinKeys(OrderIDToSequenceKeyPrefix, key)
}

// DecodeOrderIDToSequenceKey decodes order ID to sequence key and returns the account number and order ID.
func DecodeOrderIDToSequenceKey(key []byte) (uint64, string, error) {
	accNumber, orderID, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, "", err
	}

	return accNumber, string(orderID), nil
}

// CreateOrderBookRecordKey creates order book key record with fixed key length to support the correct ordering
// and be able to decode the key into the values.
func CreateOrderBookRecordKey(orderBookID uint32, side Side, price Price, orderSequence uint64) ([]byte, error) {
	key := CreateOrderBookSideKey(orderBookID, side)
	var err error
	key, err = CreateOrderBookSideRecordKey(key, price, orderSequence)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// CreateOrderBookSideRecordKey creates order book side record key.
func CreateOrderBookSideRecordKey(key []byte, price Price, orderSequence uint64) ([]byte, error) {
	priceKey, err := price.MarshallToOrderedBytes()
	if err != nil {
		return nil, err
	}
	key = store.JoinKeys(key, priceKey)
	key = store.AppendUint64ToOrderedBytes(key, orderSequence)

	return key, nil
}

// DecodeOrderBookSideRecordKey decodes order book side record key into values.
func DecodeOrderBookSideRecordKey(key []byte) (Price, uint64, error) {
	var p Price
	nextKeyPart, err := p.UnmarshallFromOrderedBytes(key)
	if err != nil {
		return Price{}, 0, err
	}
	orderSequence, _, err := store.ReadOrderedBytesToUint64(nextKeyPart)
	if err != nil {
		return Price{}, 0, err
	}

	return p, orderSequence, nil
}

// CreateOrderBookSideKey creates order book side key.
func CreateOrderBookSideKey(orderBookID uint32, side Side) []byte {
	key := make([]byte, 0)
	key = store.AppendUint32ToOrderedBytes(key, orderBookID)
	key = store.AppendUint8ToOrderedBytes(key, uint8(side))

	return store.JoinKeys(OrderBookRecordKeyPrefix, key)
}

// BuildGoodTilBlockHeightDelayKey builds the key for the good til block height delay store.
func BuildGoodTilBlockHeightDelayKey(orderSequence uint64) string {
	// the string will be store the delay store and must be unique for the app
	return fmt.Sprintf("%stlh%d", ModuleName, orderSequence)
}

// BuildGoodTilBlockTimeDelayKey builds the key for the good til block time delay store.
func BuildGoodTilBlockTimeDelayKey(orderSequence uint64) string {
	// the string will be store the delay store and must be unique for the app
	return fmt.Sprintf("%stlt%d", ModuleName, orderSequence)
}

// CreateAccountDenomOrdersCountKey creates account denom orders count key.
func CreateAccountDenomOrdersCountKey(accNumber uint64, denom string) ([]byte, error) {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, accNumber)
	denomKey, err := store.JoinKeysWithLength([]byte(denom))
	if err != nil {
		return key, err
	}

	return store.JoinKeys(AccountDenomOrdersCountKeyPrefix, key, denomKey), nil
}

// DecodeAccountDenomOrdersCountKey decodes account denom orders count key and returns the account number and denom.
func DecodeAccountDenomOrdersCountKey(key []byte) (uint64, string, error) {
	accNumber, denomWithLength, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, "", err
	}
	decodedDenomWithLength, err := store.ParseLengthPrefixedKeys(denomWithLength)
	if err != nil {
		return 0, "", err
	}
	if len(decodedDenomWithLength) != 1 {
		return 0, "", fmt.Errorf("expected decoded denom keys length is 1  got %d", len(decodedDenomWithLength))
	}
	denom := string(decodedDenomWithLength[0])

	return accNumber, denom, nil
}

// CreateAccountDenomOrderSequenceKey creates account denom order sequence key.
func CreateAccountDenomOrderSequenceKey(accNumber uint64, denom string, orderSequence uint64) ([]byte, error) {
	key, err := CreateAccountDenomKeyPrefix(accNumber, denom)
	if err != nil {
		return nil, err
	}
	key = store.AppendUint64ToOrderedBytes(key, orderSequence)

	return key, nil
}

// CreateAccountDenomKeyPrefix creates account denom key prefix.
func CreateAccountDenomKeyPrefix(accNumber uint64, denom string) ([]byte, error) {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, accNumber) // same as in method CreateAccountDenomOrdersCountKey and references the same functionality.
	denomKey, err := store.JoinKeysWithLength([]byte(denom))
	if err != nil {
		return key, err
	}
	return store.JoinKeys(AccountDenomOrderSequenceKeyPrefix, key, denomKey), nil
}

// DecodeAccountDenomKeyOrderSequence decodes the order sequence from account denom key.
func DecodeAccountDenomKeyOrderSequence(key []byte) (uint64, error) {
	orderSequence, _, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, err
	}

	return orderSequence, nil
}
