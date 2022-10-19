package types

import (
	"strconv"
)

const (
	// ModuleName defines the module name
	ModuleName = "asset"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	// FungibleTokenKeyPrefix defines the key prefix for the fungible token.
	FungibleTokenKeyPrefix = []byte{0x01}
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return JoinKeysWithLength(FungibleTokenKeyPrefix, []byte(denom))
}

// JoinKeysWithLength joins the keys with the length separation to protect from the intersecting keys
// in case the length is not fixed.
//
// Example of such behavior:
// prefix + ab + c = prefixabc
// prefix + a + bc = prefixabc
//
// Example with the usage of the func
// prefix + ab + c = prefix2ab1c
// prefix + a + ab = prefix1a2bc
func JoinKeysWithLength(prefix []byte, keys ...[]byte) []byte {
	compositeKey := make([]byte, 0)
	compositeKey = append(compositeKey, prefix...)
	for _, key := range keys {
		if len(key) == 0 {
			continue
		}
		byteLen := []byte(strconv.Itoa(len(key)))
		compositeKey = append(compositeKey, byteLen...)
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey
}
