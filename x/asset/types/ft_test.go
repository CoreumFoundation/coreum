package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestBuildFungibleTokenDenom(t *testing.T) {
	symbol := "CORE"
	addr, err := sdk.AccAddressFromBech32("devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5")
	require.NoError(t, err)

	denom := types.BuildFungibleTokenDenom(symbol, addr)
	require.Equal(t, "CORE-devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5-nvR5", denom)
}
