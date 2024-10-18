package types_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/pkg/store"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func TestNewPriceFromString(t *testing.T) {
	tests := []struct {
		strPrice string
		wantErr  bool
	}{
		{
			strPrice: "1.0e+0",
			wantErr:  false,
		},
		{
			strPrice: "11.0e+0",
			wantErr:  true,
		},
		{
			strPrice: "1.0e-0",
			wantErr:  true,
		},
		{
			strPrice: "1.0e+00",
			wantErr:  true,
		},
		{
			strPrice: "1.0e+01",
			wantErr:  true,
		},
		{
			strPrice: "1.0e-00",
			wantErr:  true,
		},
		{
			strPrice: "1.0e-01",
			wantErr:  true,
		},
		{
			strPrice: "1.0e+0+0",
			wantErr:  true,
		},
		{
			strPrice: "0.1e+1",
			wantErr:  true,
		},
		{
			strPrice: "1.10e+1",
			wantErr:  true,
		},
		{
			strPrice: "1.00e+1",
			wantErr:  true,
		},
		{
			strPrice: "1.0e1",
			wantErr:  true,
		},
		{
			strPrice: "0.2",
			wantErr:  true,
		},
		{
			strPrice: "1.00001e+0",
			wantErr:  false,
		},
		{
			strPrice: "1.23e+0",
			wantErr:  false,
		},
		{
			strPrice: "1.0e+1",
			wantErr:  false,
		},
		{
			strPrice: "1.0e-1",
			wantErr:  false,
		},
		{
			strPrice: "0.0e+1",
			wantErr:  true,
		},
		{
			strPrice: "0.0e+0",
			wantErr:  true,
		},
		{
			strPrice: "0.0e0",
			wantErr:  true,
		},
		{
			strPrice: "1.0e0",
			wantErr:  true,
		},
		{
			strPrice: "1.1e0",
			wantErr:  true,
		},
		{
			strPrice: "0.23e+1",
			wantErr:  true,
		},
		{
			strPrice: "0.03e-1",
			wantErr:  true,
		},
		{
			strPrice: "4.23e+3",
			wantErr:  false,
		},
		{
			strPrice: ".1e+1",
			wantErr:  true,
		},
		{
			strPrice: "0e+1",
			wantErr:  true,
		},
		{
			strPrice: "01e+1",
			wantErr:  true,
		},
		{
			strPrice: "9.999999999999999999e+100",
			wantErr:  false,
		},
		{
			strPrice: "9.99999999999999999999e+100",
			wantErr:  true,
		},
		{
			strPrice: "9.999999999999999999e+101",
			wantErr:  true,
		},
		{
			strPrice: "1.0e-100",
			wantErr:  false,
		},
		{
			strPrice: "0.99999999999999999e-100",
			wantErr:  true,
		},
		{
			strPrice: "1.0e-101",
			wantErr:  true,
		},
		{
			strPrice: "e+1",
			wantErr:  true,
		},
		{
			strPrice: "-1.0",
			wantErr:  true,
		},
		{
			strPrice: "-1.0e+0",
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

			wantRat, ok := big.NewRat(1, 1).SetString(tt.strPrice)
			require.True(t, ok)
			require.Equal(t, wantRat.String(), got.Rat().String())
		})
	}
}

func TestPrice_OrderedBytesMarshalling(t *testing.T) {
	tests := []struct {
		priceStr  string
		wantBytes []byte
	}{
		{
			priceStr: "1.11e+2",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -16)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1110000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "1.0e+2",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -16)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1000000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "1.3124151231234e-5",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -23)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1312415123123400000)
				return bytes
			}(),
		},
		{
			priceStr: "1.0e-100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, -118)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 1000000000000000000)
				return bytes
			}(),
		},
		{
			priceStr: "9.999999999999999999e+100",
			wantBytes: func() []byte {
				bytes := make([]byte, 0)
				bytes = store.AppendInt8ToOrderedBytes(bytes, 82)
				bytes = store.AppendUint64ToOrderedBytes(bytes, 9999999999999999999)
				return bytes
			}(),
		},
		{
			priceStr: "1.0e+100",
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
			priceStr: "1.3e-17",
			want: cbig.NewRatFromBigInts(
				big.NewInt(13), cbig.IntTenToThePower(big.NewInt(int64(18))),
			),
		},
		{
			priceStr: "1.11e+100",
			want: cbig.NewRatFromBigInts(
				cbig.IntMul(big.NewInt(111), cbig.IntTenToThePower(big.NewInt(int64(98)))), big.NewInt(1),
			),
		},
		{
			priceStr: "1.0e-100",
			want: cbig.NewRatFromBigInts(
				big.NewInt(1), cbig.IntTenToThePower(big.NewInt(int64(100))),
			),
		},
		{
			priceStr: "9.999999999999999999e+100",
			want: cbig.NewRatFromBigInts(
				cbig.IntMul(
					cbig.NewBigIntFromUint64(9999999999999999999),
					cbig.IntTenToThePower(big.NewInt(int64(82))),
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
		"1.0e+0",
		"2.3e+1",
		"1.11e+100",
		"3.1241241231241233e-23",
		"1.0e-100",
		"9.999999999999999999e+100",
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
