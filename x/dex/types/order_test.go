package types_test

import (
	"fmt"
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
		price := types.MustNewPriceFromString("1e-1")
		return types.Order{
			Creator:    sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			ID:         "aA09+:._-",
			BaseDenom:  "denom1",
			QuoteDenom: "denom2",
			Price:      price,
			Quantity:   sdkmath.NewInt(100),
			Side:       types.Side_buy,
		}
	}

	tests := []struct {
		name    string
		order   types.Order
		wantErr error
	}{
		{
			name:  "valid",
			order: validOrder(),
		},
		{
			name: "invalid_account",
			order: func() types.Order {
				order := validOrder()
				order.Creator = "inv_acc"
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_empty",
			order: func() types.Order {
				order := validOrder()
				order.ID = ""
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_too_long",
			order: func() types.Order {
				order := validOrder()
				order.ID = strings.Repeat("a", 41)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_id_prohibited_symbol",
			order: func() types.Order {
				order := validOrder()
				order.ID += "?"
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_base_denom_empty",
			order: func() types.Order {
				order := validOrder()
				order.BaseDenom = ""
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_quote_denom_empty",
			order: func() types.Order {
				order := validOrder()
				order.QuoteDenom = ""
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_quantity_negative",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.NewInt(-1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_quantity_zero",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.ZeroInt()
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_side_unspecified",
			order: func() types.Order {
				order := validOrder()
				order.Side = types.Side_unspecified
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_side_unknown",
			order: func() types.Order {
				order := validOrder()
				order.Side = 123
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_not_nil_remaining_quantity",
			order: func() types.Order {
				order := validOrder()
				order.RemainingQuantity = sdkmath.NewInt(1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_not_nil_remaining_balance",
			order: func() types.Order {
				order := validOrder()
				order.RemainingBalance = sdkmath.NewInt(1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_locked_balance",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.NewInt(111)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_locked_balance_out_or_range",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.NewInt(1_000_000)
				order.Price = types.MustNewPriceFromString(fmt.Sprintf("9999999999999999999e%d", types.MaxExp))
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			requireT := require.New(t)
			err := tt.order.Validate()
			if tt.wantErr == nil {
				requireT.NoError(err)
			} else {
				requireT.True(sdkerrors.IsOf(err, tt.wantErr))
			}
		})
	}
}
