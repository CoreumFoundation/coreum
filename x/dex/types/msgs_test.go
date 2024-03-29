package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMinPrice(t *testing.T) {
	require.True(t, sdk.OneDec().Quo(minPrice).Equal(maxPrice.Add(minPrice)))
}

func TestMaxPrice(t *testing.T) {
	require.True(t, sdk.OneDec().Quo(maxPrice).Equal(minPrice))
}
