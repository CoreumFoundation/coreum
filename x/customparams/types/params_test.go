package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestStakingParams_ValidateBasic(t *testing.T) {
	p := DefaultStakingParams()
	require.NoError(t, p.ValidateBasic())

	p.MinSelfDelegation = sdkmath.NewInt(-1)
	require.Error(t, p.ValidateBasic())
}
