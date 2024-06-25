package types_test

import (
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestNewPriceFromString(t *testing.T) {
	tests := []struct {
		strPrice     string
		wantStrPrice string
		wantErr      bool
	}{
		{
			strPrice:     "1",
			wantStrPrice: "1",
			wantErr:      false,
		},
		{
			strPrice:     "1.1",
			wantStrPrice: "1.1",
			wantErr:      false,
		},
		{
			strPrice:     "01.12",
			wantStrPrice: "1.1",
			wantErr:      false,
		},
		{
			strPrice:     "23194923818283828382838182848283828381.0",
			wantStrPrice: "23194923818283828382838182848283828381",
			wantErr:      false,
		},
		{
			strPrice: "231949238182838283828381828482838283810.0",
			wantErr:  true,
		},
		{
			strPrice:     "0.23194923818283828382838182848283828381",
			wantStrPrice: "0.23194923818283828382838182848283828381",
			wantErr:      false,
		},
		{
			strPrice:     "1.23194923818283828382838182848283820000",
			wantStrPrice: "0.2319492381828382838283818284828382",
			wantErr:      false,
		},
		{
			strPrice:     "23194923818283828382838182848283820000.23194923818283828382838182848283821234",
			wantStrPrice: "23194923818283828382838182848283820000.23194923818283828382838182848283821234",
			wantErr:      false,
		},
		{
			strPrice: "0.231949238182838283828381828482838212342",
			wantErr:  true,
		},
		{
			strPrice: "231949238182838283828381828482838212342.231949238182838283828381828482838212342",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.strPrice, func(t *testing.T) {
			got, err := types.NewPriceFromString(tt.wantStrPrice)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.Equal(t, tt.wantStrPrice, got.String())
		})
	}
}

func TestPrice_EndianBytesMarshalling(t *testing.T) {
	tests := []struct {
		name        string
		priceString string
		wantBytes   []byte
	}{
		{
			name:        "one_part_small_int",
			priceString: "100.00",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 100)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				return bytes
			}(),
		},
		{
			name:        "one_part_max_int",
			priceString: "9999999999999999999", // 10^19 - 1
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 9999999999999999999)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				return bytes
			}(),
		},
		{
			name:        "two_parts_min_int",
			priceString: "10000000000000000000", // 10^19
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 1)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				return bytes
			}(),
		},
		{
			name:        "two_parts_max_int",
			priceString: "99999999999999999999999999999999999999", // 10^(19 * 2) - 1
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 9999999999999999999)
				bytes = binary.BigEndian.AppendUint64(bytes, 9999999999999999999)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				return bytes
			}(),
		},
		{
			name:        "two_with_zeros",
			priceString: "101020203030404050506060707080809090",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 10102020303040405)
				bytes = binary.BigEndian.AppendUint64(bytes, 506060707080809090)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				return bytes
			}(),
		},
		{
			name:        "one_part_smallest_dec",
			priceString: "0.00000000000000000000000000000000000001",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 1)
				return bytes
			}(),
		},
		{
			name:        "two_parts_max_length_dec",
			priceString: "0.10000000000000000000000000000000000001",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 1000000000000000000)
				bytes = binary.BigEndian.AppendUint64(bytes, 1)
				return bytes
			}(),
		},
		{
			name:        "max_size_four_parts_int_with_dec",
			priceString: "23456789123456789123456789123456789123.12345678912345678912345678912345678912",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 2345678912345678912)
				bytes = binary.BigEndian.AppendUint64(bytes, 3456789123456789123)
				bytes = binary.BigEndian.AppendUint64(bytes, 1234567891234567891)
				bytes = binary.BigEndian.AppendUint64(bytes, 2345678912345678912)
				return bytes
			}(),
		},
		{
			name:        "four_parts_with_zeros_in_the_middle",
			priceString: "1000000020000000000000000300000004.0000000000500000000000000600000007",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = binary.BigEndian.AppendUint64(bytes, 100000002000000)
				bytes = binary.BigEndian.AppendUint64(bytes, 300000004)
				bytes = binary.BigEndian.AppendUint64(bytes, 500000000)
				bytes = binary.BigEndian.AppendUint64(bytes, 6000000070000)
				return bytes
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			priceToMarshall, err := types.NewPriceFromString(tt.priceString)
			require.NoError(t, err)
			got, err := priceToMarshall.MarshallToEndianBytes()
			require.NoError(t, err)
			if !reflect.DeepEqual(got, tt.wantBytes) {
				t.Errorf("MarshallToEndianBytes() got = %v, want %v", got, tt.wantBytes)
			}
			var priceToUnmarshall types.Price
			nextBytes, err := priceToUnmarshall.UnmarshallFromEndianBytes(got)
			require.NoError(t, err)
			require.Empty(t, nextBytes)
			require.Equal(t, priceToMarshall.String(), priceToUnmarshall.String())
		})
	}
}
