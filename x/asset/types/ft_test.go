package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestBuildFungibleTokenDenom(t *testing.T) {
	symbol := "CORE"
	addr, err := sdk.AccAddressFromBech32("cosmos1suzxj944nktr30u97g3xs3w2r8vknqnchsrg0x")
	require.NoError(t, err)

	denom := types.BuildFungibleTokenDenom(symbol, addr)
	require.Equal(t, "CORE-cosmos1suzxj944nktr30u97g3xs3w2r8vknqnchsrg0x-cbDq", denom)
}
