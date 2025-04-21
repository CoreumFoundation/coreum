package types_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func TestOrder_Validate(t *testing.T) {
	validOrder := func() types.Order {
		return types.Order{
			Creator:     sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          "aA09+:._-",
			BaseDenom:   "denom1",
			QuoteDenom:  "denom2",
			Price:       lo.ToPtr(types.MustNewPriceFromString("1e-1")),
			Quantity:    sdkmath.NewInt(100),
			Side:        types.SIDE_BUY,
			TimeInForce: types.TIME_IN_FORCE_GTC,
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
			name: "invalid_order_sequence",
			order: func() types.Order {
				order := validOrder()
				order.Sequence = 1
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
			name: "invalid_same_base_and_quote_denoms",
			order: func() types.Order {
				order := validOrder()
				order.BaseDenom = order.QuoteDenom
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
				order.Side = types.SIDE_UNSPECIFIED
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
				order.RemainingBaseQuantity = sdkmath.NewInt(1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_not_nil_remaining_balance",
			order: func() types.Order {
				order := validOrder()
				order.RemainingSpendableBalance = sdkmath.NewInt(1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_locked_balance_out_or_range",
			order: func() types.Order {
				order := validOrder()
				order.Quantity = sdkmath.NewInt(1_000_000)
				order.Price = lo.ToPtr(types.MustNewPriceFromString(fmt.Sprintf("9999999999999999999e%d", types.MaxExp)))
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_expected_to_receive_balance_out_or_range",
			order: func() types.Order {
				order := validOrder()
				order.Side = types.SIDE_SELL
				order.Quantity = sdkmath.NewInt(1_000_000)
				order.Price = lo.ToPtr(types.MustNewPriceFromString(fmt.Sprintf("9999999999999999999e%d", types.MaxExp)))
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_limit_with_nil_price",
			order: func() types.Order {
				order := validOrder()
				order.Price = nil
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_market_with_not_nil_price",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_MARKET
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_unspecified_order_type",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_UNSPECIFIED
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_good_til",
			order: func() types.Order {
				order := validOrder()
				order.GoodTil = &types.GoodTil{}
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "valid_good_til_block_height",
			order: func() types.Order {
				order := validOrder()
				order.GoodTil = &types.GoodTil{
					GoodTilBlockHeight: 1,
				}
				return order
			}(),
		},
		{
			name: "valid_good_til_block_time",
			order: func() types.Order {
				order := validOrder()
				order.GoodTil = &types.GoodTil{
					GoodTilBlockTime: lo.ToPtr(time.Now()),
				}
				return order
			}(),
		},
		{
			name: "valid_good_til_block_time_and_height",
			order: func() types.Order {
				order := validOrder()
				order.GoodTil = &types.GoodTil{
					GoodTilBlockHeight: 1,
					GoodTilBlockTime:   lo.ToPtr(time.Now()),
				}
				return order
			}(),
		},
		{
			name: "invalid_good_til_with_market",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_MARKET
				order.Price = nil
				order.GoodTil = &types.GoodTil{
					GoodTilBlockTime: lo.ToPtr(time.Now()),
				}
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "valid_ioc_time_in_force_for_market",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_MARKET
				order.Price = nil
				order.TimeInForce = types.TIME_IN_FORCE_IOC
				return order
			}(),
		},
		{
			name: "invalid_unspecified_time_in_force_for_market",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_MARKET
				order.Price = nil
				order.TimeInForce = types.TIME_IN_FORCE_UNSPECIFIED
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_unspecified_time_in_force_for_limit",
			order: func() types.Order {
				order := validOrder()
				order.TimeInForce = types.TIME_IN_FORCE_UNSPECIFIED
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_gtc_time_in_force_for_market_order",
			order: func() types.Order {
				order := validOrder()
				order.Type = types.ORDER_TYPE_MARKET
				order.Price = nil
				order.TimeInForce = types.TIME_IN_FORCE_GTC
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "valid_limit_order_with_unspecified_time_in_force_ioc",
			order: func() types.Order {
				order := validOrder()
				order.TimeInForce = types.TIME_IN_FORCE_IOC
				return order
			}(),
		},
		{
			name: "valid_limit_order_with_unspecified_time_in_force_fok",
			order: func() types.Order {
				order := validOrder()
				order.TimeInForce = types.TIME_IN_FORCE_FOK
				return order
			}(),
		},
		{
			name: "invalid_not_nil_reserve",
			order: func() types.Order {
				order := validOrder()
				order.Reserve = sdk.NewInt64Coin("denom1", 1)
				return order
			}(),
			wantErr: types.ErrInvalidInput,
		},
	}
	for _, tt := range tests {
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
