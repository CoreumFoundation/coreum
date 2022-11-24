package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestBuildFungibleTokenDenom(t *testing.T) {
	symbol := "ABC"
	addr, err := sdk.AccAddressFromBech32("devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5")
	require.NoError(t, err)

	denom := types.BuildFungibleTokenDenom(symbol, addr)
	require.Equal(t, "abc-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5", denom)
}

func TestValidateSymbol(t *testing.T) {
	requireT := require.New(t)
	unacceptableSymbols := []string{
		"ABC-1",
		"ABC/1",
		"btc-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"BTC-devcore1phjrez5j2wp5qzp0zvlqavasvw60mkp2zmfe6h",
		"core",
		"ucore",
		"Core",
		"uCore",
		"CORE",
		"UCORE",
		"3abc",
		"3ABC",
		"AB1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	acceptableSymbols := []string{
		"ABC1",
		"coreum",
		"ucoreum",
		"Coreum",
		"uCoreum",
		"COREeum",
		"A1234567890123456789012345678901234567890123456789012345678901234567890",
	}

	assertValidSymbol := func(symbol string, isValid bool) {
		err := types.ValidateSymbol(symbol)
		if isValid {
			requireT.NoError(err)
		} else {
			requireT.True(types.ErrInvalidSymbol.Is(err))
		}
	}

	for _, symbol := range unacceptableSymbols {
		assertValidSymbol(symbol, false)
	}

	for _, symbol := range acceptableSymbols {
		assertValidSymbol(symbol, true)
	}
}
