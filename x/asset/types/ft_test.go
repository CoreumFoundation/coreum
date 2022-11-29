package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestBuildFungibleTokenDenom(t *testing.T) {
	subunit := "abc"
	addr, err := sdk.AccAddressFromBech32("devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5")
	require.NoError(t, err)

	denom := types.BuildFungibleTokenDenom(subunit, addr)
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
			requireT.True(types.ErrInvalidSubunit.Is(err))
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
		if types.ErrInvalidSymbol.Is(err) == isValid {
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
