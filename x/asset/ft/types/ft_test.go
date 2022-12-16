package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func TestBuildDenom(t *testing.T) {
	subunit := "abc"
	addr, err := sdk.AccAddressFromBech32("devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5")
	require.NoError(t, err)

	denom := types.BuildDenom(subunit, addr)
	require.Equal(t, "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5", denom)
}

func TestValidateSubunit(t *testing.T) {
	requireT := require.New(t)
	unacceptableSubunits := []string{
		"",
		"T",
		"ABC1",
		"ABC-1",
		"ABC/1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"core",
		"ucore",
		"Coreum",
		"uCoreum",
		"COREeum",
		"A1234567890123456789012345678901234567890123456789012345678901234567890",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	acceptableSubunits := []string{
		"t",
		"abc1",
		"coreum",
		"ucoreum",
		"coreum",
		"ucoreum",
		"coreeum",
		"a1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSubunit := func(symbol string, isValid bool) {
		err := types.ValidateSubunit(symbol)
		if isValid {
			requireT.NoError(err)
		} else {
			requireT.True(types.ErrInvalidInput.Is(err))
		}
	}

	for _, symbol := range unacceptableSubunits {
		assertValidSubunit(symbol, false)
	}

	for _, symbol := range acceptableSubunits {
		assertValidSubunit(symbol, true)
	}
}

func TestValidateSymbol(t *testing.T) {
	assertT := assert.New(t)
	unacceptableSymbols := []string{
		"",
		".",
		"-",
		"t$",
		"t ",
		"t=",
		"t@",
		"t!",
		"ABC/1",
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
	}

	acceptableSymbols := []string{
		"t",
		"t.",
		"t-",
		"ABC-1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"abc1",
		"T",
		"ABC1",
		"coreum",
		"ucoreum",
		"Coreum",
		"uCoreum",
		"COREeum",
		"coreum",
		"ucoreum",
		"coreeum",
		"a1234567890123456789012345678901234567890123456789012345678901234567890",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSymbol := func(symbol string, isValid bool) {
		err := types.ValidateSymbol(symbol)
		if types.ErrInvalidInput.Is(err) == isValid {
			assertT.Failf("", "case: %s", symbol)
		}
	}

	for _, symbol := range unacceptableSymbols {
		assertValidSymbol(symbol, false)
	}

	for _, symbol := range acceptableSymbols {
		assertValidSymbol(symbol, true)
	}
}

func TestValidateBurnRate(t *testing.T) {
	testCases := []struct {
		rate    string
		invalid bool
	}{
		{
			rate: "0",
		},
		{
			rate: "0.00",
		},
		{
			rate: "1.00",
		},
		{
			rate: "0.10",
		},
		{
			rate: "0.10000",
		},
		{
			rate: "0.0001",
		},
		{
			rate:    "0.00001",
			invalid: true,
		},
		{
			rate:    "-0.01",
			invalid: true,
		},
		{
			rate:    "-1.0",
			invalid: true,
		},
		{
			rate:    "1.0002",
			invalid: true,
		},
		{
			rate:    "1.00023",
			invalid: true,
		},
		{
			rate:    "0.12345",
			invalid: true,
		},
		{
			rate:    "0.000000000000000001",
			invalid: true,
		},
		{
			rate:    "0.0000000000000000001",
			invalid: true,
		},
	}

	parseAndValidate := func(in string) error {
		rate, err := sdk.NewDecFromStr(in)
		if err != nil {
			return err
		}

		err = types.ValidateBurnRate(rate)
		return err
	}

	for _, tc := range testCases {
		tc := tc
		name := fmt.Sprintf("%+v", tc)
		t.Run(name, func(t *testing.T) {
			assertT := assert.New(t)
			err := parseAndValidate(tc.rate)
			if tc.invalid {
				assertT.Error(err)
			} else {
				assertT.NoError(err)
			}
		})
	}
}
