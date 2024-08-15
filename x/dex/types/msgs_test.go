package types_test

import (
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestMsgPlaceOrder_ValidateBasic(t *testing.T) {
	// single case just to test that we call the Order.Validate
	m := types.MsgPlaceOrder{}
	require.Error(t, m.ValidateBasic())
}

func TestMsgCancelOrder_ValidateBasic(t *testing.T) {
	validMsg := func() types.MsgCancelOrder {
		return types.MsgCancelOrder{
			Sender: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			ID:     "aA09+:._-",
		}
	}

	tests := []struct {
		name    string
		msg     types.MsgCancelOrder
		wantErr error
	}{
		{
			name: "valid",
			msg:  validMsg(),
		},
		{
			name: "invalid_account",
			msg: func() types.MsgCancelOrder {
				msg := validMsg()
				msg.Sender = "inv_acc"
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_id",
			msg: func() types.MsgCancelOrder {
				msg := validMsg()
				msg.ID = strings.Repeat("a", 41)
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tt.msg.ValidateBasic()
			if tt.wantErr == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tt.wantErr))
			}
		})
	}
}
