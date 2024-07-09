package types_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v4/pkg/store"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestNewPriceFromString(t *testing.T) {
	tests := []struct {
		strPrice string
		wantErr  bool
	}{
		{
			// normal price
			strPrice: "1231e-3",
			wantErr:  false,
		},
		{
			// normal price
			strPrice: "423e3",
			wantErr:  false,
		},
		{
			// normal price
			strPrice: "323141245",
			wantErr:  false,
		},
		{
			// zero price
			strPrice: "0",
			wantErr:  false,
		},
		{
			// invalid zero price with exponent
			strPrice: "0e1",
			wantErr:  true,
		},
		{
			// invalid price with leading
			strPrice: "01e1",
			wantErr:  true,
		},
		{
			// max uint64 num
			strPrice: "9999999999999999999",
			wantErr:  false,
		},
		{
			// invalid max uint64 + 1 num
			strPrice: "18446744073709551616",
			wantErr:  true,
		},
		{
			// max exp
			strPrice: "9999999999999999999e100",
			wantErr:  false,
		},
		{
			// invalid max exp + 1
			strPrice: "9999999999999999999e101",
			wantErr:  true,
		},
		{
			// min exp
			strPrice: "9999999999999999999e-100",
			wantErr:  false,
		},
		{
			// invalid min exp - 1
			strPrice: "9999999999999999999e-101",
			wantErr:  true,
		},
		{
			// invalid structure
			strPrice: "1e1e1",
			wantErr:  true,
		},
		{
			// invalid (empty) num part
			strPrice: "e1",
			wantErr:  true,
		},
		{
			// invalid num part
			strPrice: "xe1",
			wantErr:  true,
		},
		{
			// invalid empty exp part
			strPrice: "1e",
			wantErr:  true,
		},
		{
			// invalid exp part
			strPrice: "1ex",
			wantErr:  true,
		},
		{
			// invalid negative num part
			strPrice: "-1",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.strPrice, func(t *testing.T) {
			got, err := types.NewPriceFromString(tt.strPrice)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.strPrice, got.String())
		})
	}
}

func TestPrice_OrderedBytesMarshalling(t *testing.T) {
	tests := []struct {
		priceStr  string
		wantBytes []byte
	}{
		{
			priceStr: "111",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -16)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1110000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "1e2",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -16)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1000000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "13124151231234e-18",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -23)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1312415123123400000)
				return bytes
			}(),
		},
		{
			priceStr: "9999999999999999999e-100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -100)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 9999999999999999999)
				return bytes
			}(),
		},
		{
			priceStr: "9999999999999999999e100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, 100)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 9999999999999999999)
				return bytes
			}(),
		},
		{
			priceStr: "1e-100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -118)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1000000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "1e100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, 82)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1000000000000000000)
				return bytes
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.priceStr, func(t *testing.T) {
			priceToMarshall, err := types.NewPriceFromString(tt.priceStr)
			require.NoError(t, err)
			got, err := priceToMarshall.MarshallToOrderedBytes()
			require.NoError(t, err)
			require.Equal(t, tt.wantBytes, got)
			var priceToUnmarshall types.Price
			nextBytes, err := priceToUnmarshall.UnmarshallFromOrderedBytes(got)
			require.NoError(t, err)
			require.Empty(t, nextBytes)
			require.Equal(t, priceToMarshall.String(), priceToUnmarshall.String())
		})
	}
}

func TestPrice_Rat(t *testing.T) {
	tests := []struct {
		priceStr string
		want     *big.Rat
	}{
		{
			priceStr: "0",
			want: cbig.NewRatFromBigInts(
				big.NewInt(0), big.NewInt(1),
			),
		},
		{
			priceStr: "13e-18",
			want: cbig.NewRatFromBigInts(
				big.NewInt(13), cbig.IntTenToThePower(big.NewInt(int64(18))),
			),
		},
		{
			priceStr: "111e100",
			want: cbig.NewRatFromBigInts(
				cbig.IntMul(big.NewInt(111), cbig.IntTenToThePower(big.NewInt(int64(100)))), big.NewInt(1),
			),
		},
		{
			priceStr: "1e-100",
			want: cbig.NewRatFromBigInts(
				big.NewInt(1), cbig.IntTenToThePower(big.NewInt(int64(100))),
			),
		},
		{
			priceStr: "9999999999999999999e100",
			want: cbig.NewRatFromBigInts(
				cbig.IntMul(
					cbig.NewBigIntFromUint64(9999999999999999999),
					cbig.IntTenToThePower(big.NewInt(int64(100))),
				), big.NewInt(1),
			),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.priceStr, func(t *testing.T) {
			p, err := types.NewPriceFromString(tt.priceStr)
			require.NoError(t, err)
			require.Equal(t, tt.want, p.Rat())
		})
	}
}

func TestPrice_Marshalling(t *testing.T) {
	t.Parallel()

	priceStrings := []string{
		"0",
		"1",
		"23",
		"111e100",
		"31241241231241233e-23",
		"1e-100",
		"9999999999999999999e100",
	}

	for _, priceStr := range priceStrings {
		priceStr := priceStr
		t.Run(priceStr, func(t *testing.T) {
			t.Parallel()

			p, err := types.NewPriceFromString(priceStr)
			require.NoError(t, err)

			// decode and restore from buffer
			var buffer [100]byte
			n, err := p.MarshalTo(buffer[:])
			require.NoError(t, err)
			mp := types.Price{}
			require.NoError(t, mp.Unmarshal(buffer[:n]))
			require.Equal(t, mp.String(), p.String())

			// decode and restore from bytes
			b, err := p.Marshal()
			require.NoError(t, err)
			mp = types.Price{}
			require.NoError(t, mp.Unmarshal(b))
			require.Equal(t, mp.String(), p.String())

			// decode and restore json
			jb, err := p.MarshalJSON()
			require.NoError(t, err)
			mp = types.Price{}
			require.NoError(t, mp.UnmarshalJSON(jb))
			require.Equal(t, mp.String(), p.String())

			// decode and restore amino
			ab, err := p.MarshalAmino()
			require.NoError(t, err)
			mp = types.Price{}
			require.NoError(t, mp.UnmarshalAmino(ab))
			require.Equal(t, mp.String(), p.String())
		})
	}
}
