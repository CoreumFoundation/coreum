package store_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/pkg/store"
)

func TestJoinKeys(t *testing.T) {
	prefix := []byte{0x01}
	require.Equal(t, 1, len(prefix))
	require.Equal(t, 1, cap(prefix))
	// clone key to be sure it's not updated
	keyClone := append(make([]byte, 0, len(prefix)), prefix...)
	// gen new key
	denom := []byte("denom")
	key := store.JoinKeys(prefix, denom)
	require.Equal(t, append(prefix, denom...), key)
	require.Equal(t, keyClone, prefix)
}

func TestJoinKeysWithLength(t *testing.T) {
	testCases := []struct {
		keys        [][]byte
		expectError bool
	}{
		{
			keys: [][]byte{
				[]byte("key1"),
				[]byte("key2"),
				[]byte("key3"),
			},
		},
		{
			keys: [][]byte{
				[]byte("key1"),
			},
		},
		{
			keys: [][]byte{
				[]byte("key1"),
				[]byte("key2"),
				[]byte(""),
			},
			expectError: true,
		},
		{
			keys: [][]byte{
				[]byte("key1"),
				bytes.Repeat([]byte{0x1}, 256),
				[]byte("key2"),
			},
			expectError: true,
		},
		{
			keys: [][]byte{
				[]byte("1key1"),
				[]byte("2key2"),
			},
		},
		{
			keys: [][]byte{
				bytes.Repeat([]byte{0x01}, 255),
				bytes.Repeat([]byte{0x02}, 255),
				bytes.Repeat([]byte{0x03}, 255),
			},
		},
	}
	for _, tc := range testCases {
		keys := tc.keys
		compositeKey, err := store.JoinKeysWithLength(keys...)
		if tc.expectError {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		parsedKeys, err := store.ParseLengthPrefixedKeys(compositeKey)
		require.NoError(t, err)
		require.Equal(t, keys, parsedKeys)
	}
}
