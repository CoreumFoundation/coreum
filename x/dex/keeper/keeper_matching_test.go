package keeper_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestKeeper_MatchOrders(t *testing.T) {
	type AccSet struct {
		acc1 sdk.AccAddress
		acc2 sdk.AccAddress
		acc3 sdk.AccAddress
	}

	tests := []struct {
		name         string
		balances     func(accSet AccSet) map[string]sdk.Coins
		orders       func(accSet AccSet) []types.Order
		wantBalances func(accSet AccSet) map[string]sdk.Coins
		wantOrders   func(accSet AccSet) []types.Order
	}{
		{
			name: "match_self_maker_sell_taker_buy_close_maker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 376e-3 * 10000 - 375e-3 * 1000 = 3385
						RemainingBalance: sdkmath.NewInt(3385),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_self_maker_sell_taker_buy_close_maker_with_partial_filling",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1005)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						// only 1000 will be filled
						Quantity: sdkmath.NewInt(1005),
						Side:     types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 376e-3 * 10000 - 375e-3 * 1000 = 3385
						RemainingBalance: sdkmath.NewInt(3385),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5), sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_self_maker_sell_taker_buy_close_taker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1)),
				}
			},
		},
		{
			name: "match_self_maker_sell_taker_buy_close_taker_with_partial_filling",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("1"),
						// only 1000 will be filled
						Quantity: sdkmath.NewInt(1005),
						Side:     types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 630)),
				}
			},
		},
		{
			name: "match_self_maker_buy_taker_sell_close_maker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "match_self_maker_buy_taker_sell_close_taker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 376e-3 * 10000 - 376e-3 * 1000 = 3384
						RemainingBalance: sdkmath.NewInt(3384),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "match_self_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3750)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 375e-3 * 10000 - 375e-3 * 1000 = 3375
						RemainingBalance: sdkmath.NewInt(3375),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
				}
			},
		},
		{
			name: "match_self_maker_sell_taker_buy_close_both",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
				}
			},
		},
		{
			name: "match_self_close_two_makers_sell_and_and_taker_buy_with_remainder",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 60)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_sell,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_sell,
					},
					// "id3" will match with the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the account's balance
					{
						Account:    accSet.acc3.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("6e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100), sdk.NewInt64Coin(denom2, 10)),
				}
			},
		},
		{
			name: "match_self_close_two_makers_buy_and_and_taker_sell",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 200)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
					// "id3" closes "id1" and "id2", with better price for the "id3", expected to receive 80, but receive 100
					{
						Account:    accSet.acc3.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-1"),
						Quantity:   sdkmath.NewInt(200),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100)),
				}
			},
		},
		{
			name: "match_self_multiple_maker_buy_taker_sell_close_all_but_one_taker_with_same_price_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("377e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc1.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					// remain no match bad price
					{
						Account:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					// the part of the order should remain
					{
						Account:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Account:    accSet.acc2.String(),
						ID:         "id5",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("37e-2"),
						Quantity:   sdkmath.NewInt(5000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Account:           accSet.acc1.String(),
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("4e-3"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
					{
						Account:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(376),
					},
				}
			},
			wantBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1882)),
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			testApp := simapp.New(simapp.WithCustomLogger(logger))
			sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})

			acc1, _ := testApp.GenAccount(sdkCtx)
			acc2, _ := testApp.GenAccount(sdkCtx)
			acc3, _ := testApp.GenAccount(sdkCtx)
			accSet := AccSet{
				acc1: acc1,
				acc2: acc2,
				acc3: acc3,
			}

			for addr, coins := range tt.balances(accSet) {
				testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
			}

			orderBooksIDs := make(map[uint32]struct{})
			for _, order := range tt.orders(accSet) {
				require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
				orderBooksID, found, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, order.BaseDenom, order.QuoteDenom)
				require.True(t, found)
				require.NoError(t, err)
				orderBooksIDs[orderBooksID] = struct{}{}
			}

			orders := make([]types.Order, 0)
			for orderBookID := range orderBooksIDs {
				orders = append(orders, getOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.Side_buy)...)
				orders = append(orders, getOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.Side_sell)...)
			}
			require.ElementsMatch(t, tt.wantOrders(accSet), orders)

			balances := make(map[string]sdk.Coins)
			for addr := range tt.balances(accSet) {
				balances[addr] = testApp.BankKeeper.GetAllBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr))
			}

			wantBalances := tt.wantBalances(accSet)
			require.True(t, reflect.DeepEqual(wantBalances, balances), fmt.Sprintf("want: %v, got: %v", wantBalances, balances))

			// check the DEX locked balances
			lockedBalances := sdk.NewCoins()
			for _, order := range orders {
				lockedBalances = lockedBalances.Add(sdk.NewCoin(order.GetLockedBalanceDenom(), order.RemainingBalance))
			}
			wantLockedBalances := testApp.BankKeeper.GetAllBalances(
				sdkCtx, testApp.AccountKeeper.GetModuleAddress(types.ModuleName),
			)
			require.Equal(t, wantLockedBalances, lockedBalances)
		})
	}
}
