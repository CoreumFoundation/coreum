package types_test

import (
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestOrder_Validate(t *testing.T) {
	validOrder := func() types.Order {
		price, err := types.NewPriceFromString("1e1")
		require.NoError(t, err)
		return types.Order{
			Account:    sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			ID:         "aA09+:._-",
			BaseDenom:  "denom1",
			QuoteDenom: "denom2",
			Price:      price,
			Quantity:   sdk.NewInt(123),
			Side:       types.Side_sell,
		}
	}

	testCases := []struct {
		name      string
		order     types.Order
		wantError error
	}{
		{
			name:  "valid",
			order: validOrder(),
		},
		{
			name: "invalid_account",
			order: func() types.Order {
				order := validOrder()
				order.Account = "inv_acc"
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_empty",
			order: func() types.Order {
				order := validOrder()
				order.ID = ""
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_too_long",
			order: func() types.Order {
				order := validOrder()
				order.ID = strings.Repeat("a", 41)
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_prohibited_symbol",
			order: func() types.Order {
				order := validOrder()
				order.ID += "?"
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_base_denom_empty",
			order: func() types.Order {
				order := validOrder()
				order.BaseDenom = ""
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_quote_denom_empty",
			order: func() types.Order {
				order := validOrder()
				order.QuoteDenom = ""
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_quantity_negative",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.NewInt(-1)
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_quantity_zero",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.ZeroInt()
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_side_unspecified",
			order: func() types.Order {
				order := validOrder()
				order.Side = types.Side_unspecified
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
		{
			name: "invalid_side_unknown",
			order: func() types.Order {
				order := validOrder()
				order.Side = 123
				return order
			}(),
			wantError: types.ErrInvalidInput,
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tc.order.Validate()
			if tc.wantError == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tc.wantError))
			}
		})
	}
}
