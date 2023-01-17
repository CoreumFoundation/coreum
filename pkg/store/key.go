package store

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// maxKeyLen is the maximum allowed length (in bytes) for a key to be length-prefixed.
const maxKeyLen = 255

// JoinKeysWithLength joins the keys with the length separation to protect from the intersecting keys
// in case the length is not fixed.
//
// Example of such behavior:
// prefix + ab + c = prefixabc
// prefix + a + bc = prefixabc
//
// Example with the usage of the func
// prefix + ab + c = prefix2ab1c
// prefix + a + bc = prefix1a2bc
func JoinKeysWithLength(prefix []byte, key []byte) []byte {
	compositeKey := make([]byte, 0)
	compositeKey = append(compositeKey, prefix...)
	bzLen := len(key)
	if bzLen == 0 {
		return compositeKey
	}
	if bzLen > maxKeyLen {
		panic(errors.Errorf("key length should be max %d bytes, got %d", maxKeyLen, bzLen))
	}
	byteLen := proto.EncodeVarint(uint64(len(key)))
	compositeKey = append(compositeKey, byteLen...)
	compositeKey = append(compositeKey, key...)

	return compositeKey
}

// JoinKeysWithLengthMany is similar to JoinKeysWithLength but gets a list of keys and prefixes all of them
func JoinKeysWithLengthMany(keys ...[]byte) []byte {
	compositeKey := make([]byte, 0)
	for _, key := range keys {
		bzLen := len(key)
		if bzLen == 0 {
			return compositeKey
		}
		if bzLen > maxKeyLen {
			panic(errors.Errorf("key length should be max %d bytes, got %d", maxKeyLen, bzLen))
		}
		byteLen := proto.EncodeVarint(uint64(len(key)))
		compositeKey = append(compositeKey, byteLen...)
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey
}

// ParseJoinKeysWithLengthMany parses all the length prefixed keys, put together by JoinKeysWithLengthMany
func ParseJoinKeysWithLengthMany(key []byte) [][]byte {
	bzLen := len(key)
	if bzLen == 0 {
		return nil
	}
	keys := make([][]byte, 0)
	startBound := 1
	for {
		addrLen := key[startBound-1]
		endBound := startBound + int(addrLen)
		if bzLen < endBound {
			return nil
		}
		keySection := key[startBound:endBound]
		keys = append(keys, keySection)

		if endBound == bzLen {
			break
		}
		startBound = endBound + 1
	}
	return keys
}

// JoinKeys joins the keys protecting the prefixes from the modification.
func JoinKeys(keys ...[]byte) []byte {
	var length int
	for _, key := range keys {
		length += len(key)
	}

	compositeKey := make([]byte, 0, length)
	for _, key := range keys {
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey
}
