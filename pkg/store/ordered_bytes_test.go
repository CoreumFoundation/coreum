package store_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/pkg/store"
)

func TestInt8ToAndFromOrderedBytes(t *testing.T) {
	for i := math.MinInt8; i <= math.MaxUint8; i++ {
		iBytes := make([]byte, 0)
		iBytes = store.AppendInt8ToOrderedBytes(iBytes, int8(i))
		iFromBytes, remB, err := store.ReadOrderedBytesToInt8(iBytes)
		require.NoError(t, err)
		require.Empty(t, remB)
		require.Equal(t, int8(i), iFromBytes)
	}
}

func TestInt8FromOrderedBytesWithRemPart(t *testing.T) {
	b := make([]byte, 0)
	v1, v2, v3 := int8(-100), int8(0), int8(100)

	b = store.AppendInt8ToOrderedBytes(b, v1)
	b = store.AppendInt8ToOrderedBytes(b, v2)
	b = store.AppendInt8ToOrderedBytes(b, v3)

	gotV1, remB, err := store.ReadOrderedBytesToInt8(b)
	require.NoError(t, err)
	require.Equal(t, v1, gotV1)

	gotV2, remB, err := store.ReadOrderedBytesToInt8(remB)
	require.Equal(t, v2, gotV2)
	require.NoError(t, err)

	gotV3, remB, err := store.ReadOrderedBytesToInt8(remB)
	require.Equal(t, v3, gotV3)
	require.NoError(t, err)

	require.Empty(t, remB)
}

func TestUin64ToAndFromOrderedBytes(t *testing.T) {
	uint64Values := []uint64{
		0,
		9223372036854775807,
		math.MaxUint64,
	}
	for _, v := range uint64Values {
		iBytes := make([]byte, 0)
		iBytes = store.AppendUint64ToOrderedBytes(iBytes, v)
		iFromBytes, remB, err := store.ReadOrderedBytesToUint64(iBytes)
		require.NoError(t, err)
		require.Empty(t, remB)
		require.Equal(t, v, iFromBytes)
	}
}

func TestUint64FromOrderedBytesWithRemPart(t *testing.T) {
	b := make([]byte, 0)
	v1, v2, v3 := uint64(0), uint64(9223372036854775808), uint64(math.MaxUint64)

	b = store.AppendUint64ToOrderedBytes(b, v1)
	b = store.AppendUint64ToOrderedBytes(b, v2)
	b = store.AppendUint64ToOrderedBytes(b, v3)

	gotV1, remB, err := store.ReadOrderedBytesToUint64(b)
	require.NoError(t, err)
	require.Equal(t, v1, gotV1)

	gotV2, remB, err := store.ReadOrderedBytesToUint64(remB)
	require.Equal(t, v2, gotV2)
	require.NoError(t, err)

	gotV3, remB, err := store.ReadOrderedBytesToUint64(remB)
	require.Equal(t, v3, gotV3)
	require.NoError(t, err)

	require.Empty(t, remB)
}
