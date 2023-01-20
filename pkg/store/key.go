package store

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// maxKeyLen is the maximum allowed length (in bytes) for a key to be length-prefixed.
const maxKeyLen = 255

// JoinKeysWithLength joins the keys with the length separation to allow to parse back the original keys
// in case the length is not fixed.
//
// Example of such behavior:
// prefix + ab + c = prefixabc
// prefix + a + bc = prefixabc
//
// Example with the usage of the func
// prefix + ab + c = prefix2ab1c
// prefix + a + bc = prefix1a2bc
func JoinKeysWithLength(keys ...[]byte) ([]byte, error) {
	compositeKey := make([]byte, 0)
	for _, key := range keys {
		bzLen := len(key)
		if bzLen == 0 {
			return compositeKey, errors.New("received empty key")
		}
		if bzLen > maxKeyLen {
			return nil, errors.Errorf("key length should be max %d bytes, got %d", maxKeyLen, bzLen)
		}
		byteLen := proto.EncodeVarint(uint64(len(key)))
		compositeKey = append(compositeKey, byteLen...)
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey, nil
}

// ParseLengthPrefixedKeys parses all the length prefixed keys, put together by JoinKeysWithLength
func ParseLengthPrefixedKeys(key []byte) ([][]byte, error) {
	bzLen := len(key)
	if bzLen == 0 {
		return nil, errors.New("empty key")
	}
	keys := make([][]byte, 0)
	startBound := 1
	for {
		keyLen := key[startBound-1]
		endBound := startBound + int(keyLen)
		if bzLen < endBound {
			return nil, errors.New("length prefix does not match the key")
		}
		keySection := key[startBound:endBound]
		keys = append(keys, keySection)

		if endBound == bzLen {
			break
		}
		startBound = endBound + 1
	}
	return keys, nil
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
