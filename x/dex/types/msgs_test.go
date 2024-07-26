package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMsgPlaceOrder_ValidateBasic(t *testing.T) {
	// single case just to test that we call the Order.Validate
	m := MsgPlaceOrder{}
	require.Error(t, m.ValidateBasic())
}
