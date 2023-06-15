package types

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestDelayedItemKey(t *testing.T) {
	t.Parallel()

	type tCase struct {
		Name               string
		ID                 string
		Timestamp          int64
		Ok                 bool
		ExpectedTimePrefix []byte
	}

	tCases := []tCase{
		// invalid cases
		{
			Name:      "EmptyID",
			ID:        "",
			Timestamp: 1,
			Ok:        false,
		},
		{
			Name:      "MaxNegativeTimestamp",
			ID:        "id",
			Timestamp: -1,
			Ok:        false,
		},
		{
			Name:      "MinNegativeTimestamp",
			ID:        "id",
			Timestamp: math.MinInt64,
			Ok:        false,
		},

		// valid cases
		{
			Name:               "ZeroTimestamp",
			ID:                 "i",
			Timestamp:          0,
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			Name:               "MaxTimestamp",
			ID:                 "id",
			Timestamp:          math.MaxInt64,
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		{
			Name:               "TimestampByte1",
			ID:                 "idjgudf98gds8rsrfkds;fkfdisdiofidso",
			Timestamp:          1,
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x01},
		},
		{
			Name:               "TimestampByte2",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << 8,
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0},
		},
		{
			Name:               "TimestampByte3",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (2 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0},
		},
		{
			Name:               "TimestampByte4",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (3 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0},
		},
		{
			Name:               "TimestampByte5",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (4 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0},
		},
		{
			Name:               "TimestampByte6",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (5 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			Name:               "TimestampByte7",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (6 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
		{
			Name:               "TimestampByte8",
			ID:                 "idjgudf98gHJJLKJKLjdoopiods8rsrfkds;fkfdisdiofidso",
			Timestamp:          1 << (7 * 8),
			Ok:                 true,
			ExpectedTimePrefix: []byte{0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}

	for _, tc := range tCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			requireT := require.New(t)
			assertT := assert.New(t)

			expectedExecTime := time.Unix(tc.Timestamp, 0).UTC()
			key, err := CreateDelayedItemKey(tc.ID, expectedExecTime)
			if !tc.Ok {
				requireT.Error(err)
				return
			}

			requireT.NoError(err)
			assertT.Equal(DelayedItemKeyPrefix, key[0:1])
			assertT.Equal(tc.ExpectedTimePrefix, key[1:9])
			assertT.Equal(tc.ID, string(key[9:]))

			execTime, id, err := ExtractTimeAndIDFromDelayedItemKey(key[1:])
			requireT.NoError(err)

			assertT.Equal(expectedExecTime, execTime)
			assertT.Equal(tc.ID, id)
		})
	}
}

func TestInvalidDelayedItemKeys(t *testing.T) {
	t.Parallel()

	tCases := [][]byte{
		nil,
		{},
		{0x00},
		{0x00, 0x01},
		{0x00, 0x01, 0x02},
		{0x00, 0x01, 0x02, 0x03},
		{0x00, 0x01, 0x02, 0x03, 0x04},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
	}

	assertT := assert.New(t)
	for _, tc := range tCases {
		_, _, err := ExtractTimeAndIDFromDelayedItemKey(tc)
		assertT.Error(err, tc)
	}
}
