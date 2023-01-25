package store

import (
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
// prefix + a + bc = prefix1a2bc.
func JoinKeysWithLength(keys ...[]byte) ([]byte, error) {
	compositeKey := make([]byte, 0)
	for index, key := range keys {
		keyLen := len(key)
		if keyLen == 0 {
			return nil, errors.Errorf("received empty key on index %d", index)
		}
		if keyLen > maxKeyLen {
			return nil, errors.Errorf("key length should be max %d bytes, got %d", maxKeyLen, keyLen)
		}

		compositeKey = append(compositeKey, byte(keyLen))
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey, nil
}

// ParseLengthPrefixedKeys parses all the length prefixed keys, put together by JoinKeysWithLength.
func ParseLengthPrefixedKeys(key []byte) ([][]byte, error) {
	inputKeyLen := len(key)
	if inputKeyLen == 0 {
		return nil, errors.New("empty key")
	}
	keys := make([][]byte, 0)
	startBound := 1
	for {
		keyLen := key[startBound-1]
		endBound := startBound + int(keyLen)
		if inputKeyLen < endBound {
			return nil, errors.New("length prefix does not match the key")
		}
		keySection := key[startBound:endBound]
		keys = append(keys, keySection)

		if endBound == inputKeyLen {
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
