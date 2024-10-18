package types_test

import (
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	validMsg := func() types.MsgUpdateParams {
		return types.MsgUpdateParams{
			Authority: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			Params:    types.DefaultParams(),
		}
	}

	tests := []struct {
		name    string
		msg     types.MsgUpdateParams
		wantErr error
	}{
		{
			name: "valid",
			msg:  validMsg(),
		},
		{
			name: "invalid_account",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Authority = "inv-acc"
				return msg
			}(),
			wantErr: cosmoserrors.ErrInvalidAddress,
		},
		{
			name: "invalid_default_unified_ref_amount",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Params.DefaultUnifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("-0.1")
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_zero_price_tick_exponent",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Params.PriceTickExponent = 0
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_positive_price_tick_exponent",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Params.PriceTickExponent = 1
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_max_orders_per_denom",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Params.MaxOrdersPerDenom = 0
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_order_reserve",
			msg: func() types.MsgUpdateParams {
				msg := validMsg()
				msg.Params.OrderReserve = sdk.Coin{Denom: "101"}
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

func TestMsgCancelOrdersByDenom_ValidateBasic(t *testing.T) {
	validMsg := func() types.MsgCancelOrdersByDenom {
		return types.MsgCancelOrdersByDenom{
			Sender:  sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			Account: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
			Denom:   "denom1",
		}
	}

	tests := []struct {
		name    string
		msg     types.MsgCancelOrdersByDenom
		wantErr error
	}{
		{
			name: "valid",
			msg:  validMsg(),
		},
		{
			name: "invalid_sender",
			msg: func() types.MsgCancelOrdersByDenom {
				msg := validMsg()
				msg.Sender = "inv_sender"
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_account",
			msg: func() types.MsgCancelOrdersByDenom {
				msg := validMsg()
				msg.Account = "inv_sender"
				return msg
			}(),
			wantErr: types.ErrInvalidInput,
		},
		{
			name: "invalid_denom",
			msg: func() types.MsgCancelOrdersByDenom {
				msg := validMsg()
				msg.Denom = "1@1"
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

//nolint:lll // assertion strings
func TestAmino(t *testing.T) {
	const address = "devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"

	tests := []struct {
		name          string
		msg           sdk.Msg
		wantAminoJSON string
	}{
		{
			name: sdk.MsgTypeURL(&types.MsgUpdateParams{}),
			msg: &types.MsgUpdateParams{
				Authority: address,
				Params:    types.DefaultParams(),
			},
			wantAminoJSON: `{"type":"dex/MsgUpdateParams","value":{"authority":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","params":{"default_unified_ref_amount":"1000000.000000000000000000","max_orders_per_denom":"100","order_reserve":{"amount":"10000000","denom":"stake"},"price_tick_exponent":-8}}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgPlaceOrder{}),
			msg: &types.MsgPlaceOrder{
				Sender:      address,
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          "id1",
				BaseDenom:   "denom1",
				QuoteDenom:  "denom2",
				Price:       lo.ToPtr(types.MustNewPriceFromString("1.0e-1")),
				Quantity:    sdkmath.NewInt(100),
				Side:        types.SIDE_SELL,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			},
			wantAminoJSON: `{"type":"dex/MsgPlaceOrder","value":{"base_denom":"denom1","id":"id1","price":"1.0e-1","quantity":"100","quote_denom":"denom2","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","side":2,"time_in_force":1,"type":1}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgCancelOrder{}),
			msg: &types.MsgCancelOrder{
				Sender: address,
				ID:     "id1",
			},
			wantAminoJSON: `{"type":"dex/MsgCancelOrder","value":{"id":"id1","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
		{
			name: sdk.MsgTypeURL(&types.MsgCancelOrdersByDenom{}),
			msg: &types.MsgCancelOrdersByDenom{
				Sender:  address,
				Account: address,
				Denom:   "denom1",
			},
			wantAminoJSON: `{"type":"dex/MsgCancelOrdersByDenom","value":{"account":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5","denom":"denom1","sender":"devcore172rc5sz2uclpsy3vvx3y79ah5dk450z5ruq2r5"}}`,
		},
	}

	legacyAmino := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(legacyAmino)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			generatedJSON := legacyAmino.Amino.MustMarshalJSON(tt.msg)
			require.Equal(t, tt.wantAminoJSON, string(sdk.MustSortJSON(generatedJSON)))
		})
	}
}
