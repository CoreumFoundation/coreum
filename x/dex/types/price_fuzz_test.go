package types_test

import (
	"fmt"
	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func FuzzPriceFromRandomString(f *testing.F) {
	// valid
	f.Add("1.0e+0")
	f.Add("1.231e-6")
	f.Add("5.423e+3")
	f.Add("3.23141245e+10")
	f.Add("1.0e+1")
	f.Add("9.999999999999999999e+100")
	f.Add("1.0e-100")
	// invalid
	f.Add("0")
	f.Add("0.0")
	f.Add("18446744073709551616")
	f.Add("9.999999999999999999e+101")
	f.Add("1.2e-101")
	f.Add("0e+1")
	f.Add("e+0")

	f.Fuzz(func(t *testing.T, priceStr string) {
		assertPriceConversation(t, priceStr, true)
	})
}

func FuzzPriceFromValidParts(f *testing.F) {
	f.Add(uint64(123), types.MaxExp)
	f.Add(uint64(4123123123), types.MinExp)
	f.Add(uint64(9999999999999999999), types.MaxExp)
	f.Add(uint64(1), types.MinExp)

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		if num == 0 { // num 0 is invalid update 1 to make valid
			num = 1
		}
		numPart := strconv.FormatUint(num, 10)
		numPart = strings.TrimRight(numPart, "0")

		if len(numPart) > 1 {
			numPart = string(numPart[0]) + types.DotSymbol + numPart[1:]
		} else {
			numPart = numPart + types.DotSymbol + "0"
		}

		if len(numPart) > types.MaxNumLen {
			numPart = numPart[:types.MaxNumLen]
		}

		numParts := strings.Split(numPart, types.DotSymbol)
		numDecimalPart := numParts[1]

		if len(numDecimalPart) > 1 && strings.HasSuffix(numDecimalPart, "0") {
			// if ends with zero update to 1 to make valid
			numPart = numPart[:len(numPart)-1]
			numPart += "1"
		}

		if exp > types.MaxExp {
			exp = types.MaxExp
		}
		if exp < types.MinExp {
			exp = types.MinExp
		}

		var expPart string
		if exp < 0 {
			expPart = types.ExponentSymbol + strconv.Itoa(int(exp))
		} else {
			expPart = types.ExponentSymbol + "+" + strconv.Itoa(int(exp))
		}
		priceStr := numPart + expPart
		assertPriceConversation(t, priceStr, false)
	})
}

func FuzzPriceFromFloatScientific(f *testing.F) {
	f.Add(uint64(1), int8(0))
	f.Add(uint64(123), int8(50))
	f.Add(uint64(4123123123), int8(30))
	f.Add(uint64(9999999999999999999), int8(-20))
	f.Add(uint64(1), int8(-99))

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		if num == 0 { // num 0 is invalid update 1 to make valid
			num = 1
		}

		v := big.NewFloat(0).
			Mul(
				big.NewFloat(0).SetInt(cbig.NewBigIntFromUint64(num)),
				big.NewFloat(0).SetInt(cbig.IntTenToThePower(big.NewInt(int64(exp)))),
			)

		rawPriceStr := v.Text('e', types.MaxNumLen-1)
		parts := strings.Split(rawPriceStr, types.ExponentSymbol)

		numPart := parts[0]
		numParts := strings.Split(numPart, types.DotSymbol)
		numIntPart := numParts[0]
		numDecimalPart := numParts[1]

		numDecimalPart = strings.TrimRight(numDecimalPart, "0")
		if len(numDecimalPart) == 0 {
			numDecimalPart = "0"
		}

		expPart := parts[1]
		if len(expPart) > 2 && strings.HasPrefix(expPart, "+0") {
			expPart = strings.ReplaceAll(expPart, "+0", "+")
		}

		intExp, err := strconv.ParseInt(expPart, 10, 64)
		require.NoError(t, err)
		// adjust the exponent to valid
		if intExp > int64(types.MaxExp) {
			expPart = fmt.Sprintf("+%d", types.MaxExp)
		}
		if intExp < int64(types.MinExp) {
			expPart = fmt.Sprintf("-%d", types.MinExp)
		}

		priceStr := numIntPart + types.DotSymbol + numDecimalPart + types.ExponentSymbol + expPart
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
