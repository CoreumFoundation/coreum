package store

import "github.com/gogo/protobuf/proto"

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
func JoinKeysWithLength(prefix, key []byte) []byte {
	compositeKey := make([]byte, 0)
	compositeKey = append(compositeKey, prefix...)
	if len(key) == 0 {
		return compositeKey
	}
	byteLen := proto.EncodeVarint(uint64(len(key)))
	compositeKey = append(compositeKey, byteLen...)
	compositeKey = append(compositeKey, key...)

	return compositeKey
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
