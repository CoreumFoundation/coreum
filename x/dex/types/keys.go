package types

import (
	"crypto/sha256"
	"fmt"

	sdkerrors "cosmossdk.io/errors"

	"github.com/CoreumFoundation/coreum/v4/pkg/store"
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
	// OrderBookSeqKey defines the key for the order book sequence.
	OrderBookSeqKey = []byte{0x02}
	// OrderBookDataKeyPrefix defines the key prefix for the order book data.
	OrderBookDataKeyPrefix = []byte{0x03}
	// OrderSeqKey defines the key for the order sequence.
	OrderSeqKey = []byte{0x04}
	// OrderKeyPrefix defines the key prefix for the order.
	OrderKeyPrefix = []byte{0x05}
	// OrderIDToSeqKeyPrefix defines the key prefix for the order ID to sequence.
	OrderIDToSeqKeyPrefix = []byte{0x06}
	// OrderBookRecordKeyPrefix defines the key prefix for the order book record.
	OrderBookRecordKeyPrefix = []byte{0x07}
	// ParamsKey defines the key to store parameters of the module, set via governance.
	ParamsKey = []byte{0x09}
	// AccountDenomOrdersCountKeyPrefix defines the key prefix for the account denom orders count.
	AccountDenomOrdersCountKeyPrefix = []byte{0x10}
	// AccountDenomOrderSeqKeyPrefix defines the key prefix for the account denom order seq.
	AccountDenomOrderSeqKeyPrefix = []byte{0x11}
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
func CreateOrderKey(orderSeq uint64) []byte {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, orderSeq)
	return store.JoinKeys(OrderKeyPrefix, key)
}

// CreateOrderIDToSeqKey creates order ID to sequence key.
func CreateOrderIDToSeqKey(accountNumber uint64, orderID string) []byte {
	return store.JoinKeys(CreateOrderIDToSeqKeyPrefix(accountNumber), []byte(orderID))
}

// CreateOrderIDToSeqKeyPrefix creates order ID to sequence key prefix.
func CreateOrderIDToSeqKeyPrefix(accountNumber uint64) []byte {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, accountNumber)
	return store.JoinKeys(OrderIDToSeqKeyPrefix, key)
}

// DecodeOrderIDToSeqKey decodes order ID to sequence key and returns the account number and order ID.
func DecodeOrderIDToSeqKey(key []byte) (uint64, string, error) {
	accNumber, orderID, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, "", err
	}

	return accNumber, string(orderID), nil
}

// CreateOrderBookRecordKey creates order book key record with fixed key length to support the correct ordering
// and be able to decode the key into the values.
func CreateOrderBookRecordKey(orderBookID uint32, side Side, price Price, orderSeq uint64) ([]byte, error) {
	key := CreateOrderBookSideKey(orderBookID, side)
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
func CreateOrderBookSideKey(orderBookID uint32, side Side) []byte {
	key := make([]byte, 0)
	key = store.AppendUint32ToOrderedBytes(key, orderBookID)
	key = store.AppendUint8ToOrderedBytes(key, uint8(side))

	return store.JoinKeys(OrderBookRecordKeyPrefix, key)
}

// BuildGoodTilBlockHeightDelayKey builds the key for the good til block height delay store.
func BuildGoodTilBlockHeightDelayKey(orderSeq uint64) string {
	// the string will be store the delay store and must be unique for the app
	return fmt.Sprintf("%stlh%d", ModuleName, orderSeq)
}

// BuildGoodTilBlockTimeDelayKey builds the key for the good til block time delay store.
func BuildGoodTilBlockTimeDelayKey(orderSeq uint64) string {
	// the string will be store the delay store and must be unique for the app
	return fmt.Sprintf("%stlt%d", ModuleName, orderSeq)
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

// CreateAccountDenomOrderSeqKey creates account denom order seq key.
func CreateAccountDenomOrderSeqKey(accNumber uint64, denom string, orderSeq uint64) ([]byte, error) {
	key, err := CreateAccountDenomKeyPrefix(accNumber, denom)
	if err != nil {
		return nil, err
	}
	key = store.AppendUint64ToOrderedBytes(key, orderSeq)

	return key, nil
}

// CreateAccountDenomKeyPrefix creates account denom key prefix.
func CreateAccountDenomKeyPrefix(accNumber uint64, denom string) ([]byte, error) {
	key := make([]byte, 0)
	key = store.AppendUint64ToOrderedBytes(key, accNumber)
	denomKey, err := store.JoinKeysWithLength([]byte(denom))
	if err != nil {
		return key, err
	}
	return store.JoinKeys(AccountDenomOrderSeqKeyPrefix, key, denomKey), nil
}

// DecodeAccountDenomKeyOrderSeq decodes the order seq from account denom key.
func DecodeAccountDenomKeyOrderSeq(key []byte) (uint64, error) {
	orderSeq, _, err := store.ReadOrderedBytesToUint64(key)
	if err != nil {
		return 0, err
	}

	return orderSeq, nil
}
