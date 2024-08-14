package keeper_test

import (
	"fmt"
	"reflect"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

type AccSet struct {
	acc1 sdk.AccAddress
	acc2 sdk.AccAddress
	acc3 sdk.AccAddress
}

func TestKeeper_MatchOrders(t *testing.T) {
	tests := []struct {
		name                  string
		balances              func(accSet AccSet) map[string]sdk.Coins
		orders                func(accSet AccSet) []types.Order
		wantAvailableBalances func(accSet AccSet) map[string]sdk.Coins
		wantOrders            func(accSet AccSet) []types.Order
	}{
		// ******************** No matching ********************

		{
			name: "no_match_self_and_opposite_buy_and_sell",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2659), sdk.NewInt64Coin(denom2, 375)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("266e-2"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id4",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("2659e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("376e-3"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("375e-3"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(375),
					},
					{
						Creator:           accSet.acc1.String(),
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("266e-2"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id4",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("2659e-3"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(2659),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},

		// ******************** Self matching ********************

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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc2.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						// only 1000 will be filled
						Quantity: sdkmath.NewInt(1005),
						Side:     types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc2.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc1.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc1.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc2.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc1.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:    accSet.acc1.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_sell,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:    accSet.acc3.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
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
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
					// "id3" closes "id1" and "id2", with better price for the "id3", expected to receive 80, but receive 100
					{
						Creator:    accSet.acc3.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100)),
				}
			},
		},
		{
			name: "match_self_multiple_maker_buy_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("377e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc1.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					// remain no match bad price
					{
						Creator:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					// the part of the order should remain
					{
						Creator:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
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
						Creator:           accSet.acc1.String(),
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
						Creator:    accSet.acc1.String(),
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
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1882)),
				}
			},
		},
		{
			name: "match_self_multiple_maker_sell_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000+2000+1000+2000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1890)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc1.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					// remain no match bad price
					{
						Creator:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					// the part of the order should remain
					{
						Creator:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id5",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("378e-3"),
						Quantity:   sdkmath.NewInt(5000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("4e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},

					{
						Creator:           accSet.acc1.String(),
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("376e-3"),
						Quantity:          sdkmath.NewInt(2000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1878)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000), sdk.NewInt64Coin(denom2, 12)),
				}
			},
		},

		// ******************** Opposite matching ********************

		{
			name: "match_opposite_maker_sell_taker_sell_close_maker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,

						RemainingQuantity: sdkmath.NewInt(9625),
						RemainingBalance:  sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_opposite_maker_sell_taker_sell_close_maker_with_partial_filling",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1001)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(1001),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,

						RemainingQuantity: sdkmath.NewInt(9625),
						RemainingBalance:  sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1), sdk.NewInt64Coin(denom2, 375)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_opposite_maker_sell_taker_sell_close_taker_with",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(999),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,

						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2664)),
				}
			},
		},
		{
			name: "match_opposite_maker_sell_taker_sell_close_taker_with_partial_filling",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1001)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(1001),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,

						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2664), sdk.NewInt64Coin(denom2, 2)),
				}
			},
		},
		{
			name: "match_opposite_maker_buy_taker_buy_close_maker",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 381)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 26500)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("381e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,

						RemainingQuantity: sdkmath.NewInt(9619),
						RemainingBalance:  sdkmath.NewInt(25500),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 381)),
				}
			},
		},
		{
			name: "match_opposite_maker_buy_taker_buy_close_taker_with_partial_filling",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3810)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2650)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("381e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("265e-2"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("381e-3"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,

						RemainingQuantity: sdkmath.NewInt(8000),
						RemainingBalance:  sdkmath.NewInt(3048),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 650), sdk.NewInt64Coin(denom2, 762)),
				}
			},
		},
		{
			name: "match_opposite_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,

						RemainingQuantity: sdkmath.NewInt(9500),
						RemainingBalance:  sdkmath.NewInt(9500),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
				}
			},
		},
		{
			name: "match_opposite_maker_sell_taker_sell_close_both",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("2"),
						Quantity:   sdkmath.NewInt(500),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
				}
			},
		},
		{
			name: "match_opposite_close_two_makers_buy_and_and_taker_buy_with_remainder",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 105)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_buy,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:    accSet.acc3.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("21e-1"),
						Quantity:   sdkmath.NewInt(50),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5), sdk.NewInt64Coin(denom2, 50)),
				}
			},
		},
		{
			name: "match_opposite_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4995)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("377e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc1.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					// remain no match bad price
					{
						Creator:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-3"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					// the part of the order should remain
					{
						Creator:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id5",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("27e-1"),
						Quantity:   sdkmath.NewInt(1850),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
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
						Creator:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_buy,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1125),
						RemainingBalance:  sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4875)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 120), sdk.NewInt64Coin(denom2, 1835)),
				}
			},
		},
		{
			name: "match_opposite_multiple_maker_sell_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000+2000+1000+2000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1880)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc1.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("375e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					// remain no match bad price
					{
						Creator:    accSet.acc1.String(),
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("4e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					// the part of the order should remain
					{
						Creator:    accSet.acc1.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("376e-3"),
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id5",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("26e-1"),
						Quantity:   sdkmath.NewInt(1880),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("4e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},

					{
						Creator:           accSet.acc1.String(),
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("376e-3"),
						Quantity:          sdkmath.NewInt(2000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1878)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000), sdk.NewInt64Coin(denom2, 2)),
				}
			},
		},

		// ******************** Combined matching ********************

		{
			name: "match_self_and_opposite_buy_close_opposite_taker_with_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100+10000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// order has fifo priority
						Price:    types.MustNewPriceFromString("181e-2"),
						Quantity: sdkmath.NewInt(100),
						Side:     types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// will remain with the partial filling
						Price:    types.MustNewPriceFromString("181e-2"),
						Quantity: sdkmath.NewInt(10000),
						Side:     types.Side_sell,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("49e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("5e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(500),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("181e-2"),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(4600),
						RemainingBalance:  sdkmath.NewInt(4600),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9955)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 45), sdk.NewInt64Coin(denom2, 5500)),
				}
			},
		},
		{
			name: "match_self_and_opposite_buy_close_self_taker_with_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500+5000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 220)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("21e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// order "id1", "id2" and "id3" matches, but we fill partially only "id2"
						//	with the best price and fifo priority
						Price:    types.MustNewPriceFromString("5e-1"),
						Quantity: sdkmath.NewInt(1000),
						Side:     types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("22e-1"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("21e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id2",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("5e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(800),
						RemainingBalance:  sdkmath.NewInt(400),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("5e-1"),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(5000),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 200)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100), sdk.NewInt64Coin(denom2, 20)),
				}
			},
		},
		{
			name: "match_self_and_opposite_sell_close_opposite_taker_with_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000+19000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 825)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("19e-1"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("55e-2"),
						Quantity:   sdkmath.NewInt(1500),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc2.String(),
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("5e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(500),
						RemainingBalance:  sdkmath.NewInt(500),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("19e-1"),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(19000),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 250)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1500), sdk.NewInt64Coin(denom2, 75)),
				}
			},
		},
		{
			name: "match_self_and_opposite_sell_close_self_taker_with_fifo_priority",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 2100)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 2100+10000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("21e-1"),
						// order "id1", "id2" and "id3" matches, but we fill partially only "id1"
						//	with the best price and fifo priority
						Quantity: sdkmath.NewInt(1000),
						Side:     types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("21e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("19e-1"),
						Quantity:   sdkmath.NewInt(10),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc1.String(),
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("21e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(990),
						RemainingBalance:  sdkmath.NewInt(2079),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("21e-1"),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(2100),
					},
					{
						Creator:           accSet.acc2.String(),
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             types.MustNewPriceFromString("5e-1"),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 21)),
				}
			},
		},
		{
			name: "match_self_and_opposite_buy_close_all_makers",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100+10000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100000)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("181e-2"),
						Quantity:   sdkmath.NewInt(100),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("181e-2"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("49e-2"),
						Quantity:   sdkmath.NewInt(100000),
						Side:       types.Side_sell,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc3.String(),
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("49e-2"),
						Quantity:          sdkmath.NewInt(100000),
						Side:              types.Side_sell,
						RemainingQuantity: sdkmath.NewInt(80719),
						RemainingBalance:  sdkmath.NewInt(80719),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 18281)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10600)),
				}
			},
		},
		{
			name: "match_self_and_opposite_sell_close_all_makers",
			balances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000+19000)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 82500)),
				}
			},
			orders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:    accSet.acc1.String(),
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("5e-1"),
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.Side_sell,
					},
					{
						Creator:    accSet.acc2.String(),
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Price:      types.MustNewPriceFromString("19e-1"),
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.Side_buy,
					},
					{
						Creator:    accSet.acc3.String(),
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      types.MustNewPriceFromString("55e-2"),
						Quantity:   sdkmath.NewInt(150000),
						Side:       types.Side_buy,
					},
				}
			},
			wantOrders: func(accSet AccSet) []types.Order {
				return []types.Order{
					{
						Creator:           accSet.acc3.String(),
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             types.MustNewPriceFromString("55e-2"),
						Quantity:          sdkmath.NewInt(150000),
						Side:              types.Side_buy,
						RemainingQuantity: sdkmath.NewInt(129000),
						RemainingBalance:  sdkmath.NewInt(71500),
					},
				}
			},
			wantAvailableBalances: func(accSet AccSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					accSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					accSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10500)),
					accSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 21000)),
				}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewTestLogger(t)
			testApp := simapp.New(simapp.WithCustomLogger(logger))
			sdkCtx := testApp.BaseApp.NewContext(false)

			accSet := getAccSet(sdkCtx, testApp)
			for addr, coins := range tt.balances(accSet) {
				testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
			}

			orderBooksIDs := make(map[uint32]struct{})
			initialOrders := tt.orders(accSet)
			for _, order := range initialOrders {
				require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
				orderBooksID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, order.BaseDenom, order.QuoteDenom)
				require.NoError(t, err)
				orderBooksIDs[orderBooksID] = struct{}{}
			}

			orders := make([]types.Order, 0)
			for orderBookID := range orderBooksIDs {
				orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.Side_buy)...)
				orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.Side_sell)...)
			}
			require.ElementsMatch(t, tt.wantOrders(accSet), orders)

			availableBalances := make(map[string]sdk.Coins)
			lockedBalances := make(map[string]sdk.Coins)
			for addr := range tt.balances(accSet) {
				addrBalances := testApp.BankKeeper.GetAllBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr))
				addrFTLockedBalances := sdk.NewCoins()
				for _, balance := range addrBalances {
					lockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
						sdkCtx, sdk.MustAccAddressFromBech32(addr), balance.Denom,
					)
					addrFTLockedBalances = addrFTLockedBalances.Add(lockedBalance)
					addrBalances = addrBalances.Sub(lockedBalance)
				}
				availableBalances[addr] = addrBalances
				lockedBalances[addr] = addrFTLockedBalances
			}
			availableBalances = removeEmptyBalances(availableBalances)
			lockedBalances = removeEmptyBalances(lockedBalances)

			wantAvailableBalances := tt.wantAvailableBalances(accSet)
			require.True(
				t,
				reflect.DeepEqual(wantAvailableBalances, availableBalances),
				fmt.Sprintf("want: %v, got: %v", wantAvailableBalances, availableBalances),
			)

			// check that balance locked in the orders correspond the balance locked in the asset ft
			orderLockedBalances := make(map[string]sdk.Coins)
			for _, order := range orders {
				coins, ok := orderLockedBalances[order.Creator]
				if !ok {
					coins = sdk.NewCoins()
				}
				coins = coins.Add(sdk.NewCoin(order.GetBalanceDenom(), order.RemainingBalance))
				orderLockedBalances[order.Creator] = coins
			}
			orderLockedBalances = removeEmptyBalances(orderLockedBalances)
			require.True(
				t,
				reflect.DeepEqual(lockedBalances, orderLockedBalances),
				fmt.Sprintf("want: %v, got: %v", lockedBalances, orderLockedBalances),
			)
		})
	}
}

func getAccSet(sdkCtx sdk.Context, testApp *simapp.App) AccSet {
	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)
	acc3, _ := testApp.GenAccount(sdkCtx)
	accSet := AccSet{
		acc1: acc1,
		acc2: acc2,
		acc3: acc3,
	}
	return accSet
}

func removeEmptyBalances(balances map[string]sdk.Coins) map[string]sdk.Coins {
	for addr, balance := range balances {
		if balance.IsZero() {
			delete(balances, addr)
		}
	}

	return balances
}
