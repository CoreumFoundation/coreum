package store_test

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

func TestJoinKeysWithLength(t *testing.T) {
	prefix := []byte{0x01}
	require.Equal(t, 1, len(prefix))
	require.Equal(t, 1, cap(prefix))
	// clone key to be sure it's not updated
	keyClone := append(make([]byte, 0, len(prefix)), prefix...)
	// gen new key
	denom := []byte("denom")
	key := store.JoinKeysWithLength(prefix, denom)
	exp := make([]byte, 0)
	exp = append(exp, prefix...)
	exp = append(exp, proto.EncodeVarint(uint64(len(denom)))...)
	exp = append(exp, denom...)
	require.Equal(t, exp, key)
	require.Equal(t, keyClone, prefix)
}

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
