package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestStakingParams_ValidateBasic(t *testing.T) {
	p := DefaultStakingParams()
	require.NoError(t, p.ValidateBasic())

	p.MinSelfDelegation = sdk.NewInt(-1)
	require.Error(t, p.ValidateBasic())
}
