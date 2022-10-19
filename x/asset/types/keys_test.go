package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestJoinKeys(t *testing.T) {
	prefix := []byte{0x01}
	require.Equal(t, 1, len(prefix))
	require.Equal(t, 1, cap(prefix))
	// clone key to be sure it's not updated
	keyClone := append(make([]byte, 0, len(prefix)), prefix...)
	// gen new key
	denom := []byte("denom")
	key := types.JoinKeys(prefix, denom)
	require.Equal(t, append(prefix, denom...), key)
	require.Equal(t, keyClone, prefix)
}
