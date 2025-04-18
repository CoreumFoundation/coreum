package types_test

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func FuzzPriceFromRandomString(f *testing.F) {
	// valid
	f.Add("1")
	f.Add("1231e-3")
	f.Add("423e3")
	f.Add("323141245")
	f.Add("1e1")
	f.Add("9999999999999999999")
	f.Add("9999999999999999999e100")
	f.Add("1e-100")
	// invalid
	f.Add("0")
	f.Add("18446744073709551616")
	f.Add("9999999999999999999e101")
	f.Add("9999999999999999999e-101")
	f.Add("0e1")
	f.Add("e+0")

	f.Fuzz(func(t *testing.T, priceStr string) {
		assertPriceConversation(t, priceStr, true)
	})
}

func FuzzPriceFromValidParts(f *testing.F) {
	f.Add(uint64(1), int8(0))
	f.Add(uint64(123), types.MaxExp)
	f.Add(uint64(4123123123), types.MinExp)
	f.Add(uint64(9999999999999999999), types.MaxExp)
	f.Add(uint64(1), types.MinExp)

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		var expPart string
		if exp != 0 {
			expPart = types.ExponentSymbol + strconv.Itoa(int(exp))
		}
		numPart := strconv.FormatUint(num, 10)
		if strings.HasSuffix(numPart, "0") || len(numPart) > types.MaxNumLen {
			t.Skip()
		}
		if exp > types.MaxExp || exp < types.MinExp {
			t.Skip()
		}
		priceStr := strconv.FormatUint(num, 10) + expPart

		assertPriceConversation(t, priceStr, false)
	})
}

func assertPriceConversation(t *testing.T, priceStr string, allowStringToPriceError bool) {
	t.Logf("Asserting price conversations priceStr:%s, rev=%t", priceStr, allowStringToPriceError)
	price := assetPriceFromStringAndBack(t, priceStr, allowStringToPriceError)
	assetPriceMarshalAndUnmarshalOrderedBytes(t, price)
	assetPriceMarshalAndUnmarshal(t, price)
	assetPriceMarshalAndUnmarshalAmino(t, price)
	assetPriceMarshalAndUnmarshalJSON(t, price)

	// check that rat from string is same as we generate
	wantRat, ok := big.NewRat(1, 1).SetString(priceStr)
	require.True(t, ok)
	require.Equal(t, wantRat.String(), price.Rat().String())
}

func assetPriceFromStringAndBack(t *testing.T, priceStr string, allowStringToPriceError bool) types.Price {
	price, err := types.NewPriceFromString(priceStr)
	if allowStringToPriceError && err != nil {
		t.Skip()
	}
	require.NoError(t, err)
	require.Equal(t, priceStr, price.String())

	restoredFromString, err := types.NewPriceFromString(price.String())
	require.NoError(t, err)
	require.Equal(t, priceStr, restoredFromString.String())

	return price
}

func assetPriceMarshalAndUnmarshalOrderedBytes(t *testing.T, price types.Price) {
	bytes, err := price.MarshallToOrderedBytes()
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	restoredFromBytes := &types.Price{}
	rem, err := restoredFromBytes.UnmarshallFromOrderedBytes(bytes)
	require.NoError(t, err)
	require.Empty(t, rem)
	require.Equal(t, price.String(), restoredFromBytes.String())
}

func assetPriceMarshalAndUnmarshal(t *testing.T, price types.Price) {
	bytes, err := price.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	restoredFromBytes := &types.Price{}
	require.NoError(t, restoredFromBytes.Unmarshal(bytes))
	require.Equal(t, price.String(), restoredFromBytes.String())
}

func assetPriceMarshalAndUnmarshalAmino(t *testing.T, price types.Price) {
	bytes, err := price.MarshalAmino()
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	restoredFromBytes := &types.Price{}
	require.NoError(t, restoredFromBytes.UnmarshalAmino(bytes))
	require.Equal(t, price.String(), restoredFromBytes.String())
}

func assetPriceMarshalAndUnmarshalJSON(t *testing.T, price types.Price) {
	bytes, err := price.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, bytes)

	restoredFromBytes := &types.Price{}
	require.NoError(t, restoredFromBytes.UnmarshalJSON(bytes))
	require.Equal(t, price.String(), restoredFromBytes.String())
}
