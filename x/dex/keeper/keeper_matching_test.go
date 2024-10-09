package keeper_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

type TestSet struct {
	acc1 sdk.AccAddress
	acc2 sdk.AccAddress
	acc3 sdk.AccAddress

	issuer               sdk.AccAddress
	ftDenomWhitelisting1 string
	ftDenomWhitelisting2 string
}

func TestKeeper_MatchOrders(t *testing.T) {
	tests := []struct {
		name                             string
		balances                         func(testSet TestSet) map[string]sdk.Coins
		whitelistedBalances              func(testSet TestSet) map[string]sdk.Coins
		orders                           func(testSet TestSet) []types.Order
		wantOrders                       func(testSet TestSet) []types.Order
		wantAvailableBalances            func(testSet TestSet) map[string]sdk.Coins
		wantWhitelistingReservedBalances func(testSet TestSet) map[string]sdk.Coins
		wantErrorContains                string
	}{
		// ******************** No matching ********************

		{
			name: "no_match_limit_self_and_opposite_buy_and_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2659), sdk.NewInt64Coin(denom2, 375)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("266e-2")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2659e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(375),
					},
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("266e-2")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("2659e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(2659),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_market_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// no orders in the order book so nothing to lock
				return map[string]sdk.Coins{}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_market_buy",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// no orders in the order book so nothing to lock
				return map[string]sdk.Coins{}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "try_to_match_limit_self_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "1000denom1 is not available, available 999denom1",
		},

		// ******************** Self limit matching ********************

		{
			name: "match_limit_self_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3761)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10001 - 1000
						RemainingQuantity: sdkmath.NewInt(9001),
						// ceil(376e-3 * (10001 - 1000)) = 3385
						RemainingBalance: sdkmath.NewInt(3385),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_maker_same_account",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 3761)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10001 - 1000
						RemainingQuantity: sdkmath.NewInt(9001),
						// ceil(376e-3 * (10001-1000)) = 3385
						RemainingBalance: sdkmath.NewInt(3385),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "try_to_match_limit_self_maker_sell_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3758)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			// we fill the id1 first, so the remaining balance will be 3759 - 1000 * 375e-3 = 3384,
			// but we need to lock (10000 - 1000) * 376e-3 = 3384
			wantErrorContains: "3384denom2 is not available, available 3383denom2",
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_maker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1005)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// (10000 - 1000) * 376e-3 = 3384
						RemainingBalance: sdkmath.NewInt(3384),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5), sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 1)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 377)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 2)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 630)),
				}
			},
		},
		{
			name: "match_limit_self_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "try_to_match_limit_self_maker_buy_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "9000denom1 is not available, available 8999denom1",
		},
		{
			name: "match_limit_self_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 376e-3 * 10000 - 376e-3 * 1000 = 3384
						RemainingBalance: sdkmath.NewInt(3384),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "match_limit_self_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3750)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 375e-3 * 10000 - 375e-3 * 1000 = 3375
						RemainingBalance: sdkmath.NewInt(3375),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
				}
			},
		},
		{
			name: "match_limit_self_close_two_makers_sell_and_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 60)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("6e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100), sdk.NewInt64Coin(denom2, 10)),
				}
			},
		},
		{
			name: "match_limit_self_close_two_makers_buy_and_and_taker_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 50)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 200)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" closes "id1" and "id2", with better price for the "id3", expected to receive 80, but receive 100
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(200),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100)),
				}
			},
		},
		{
			name: "match_limit_self_multiple_maker_buy_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("37e-2")),
						Quantity:    sdkmath.NewInt(5000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(376),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1882)),
				}
			},
		},
		{
			name: "match_limit_self_multiple_maker_sell_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000+2000+1000+2000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1890)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("378e-3")),
						Quantity:    sdkmath.NewInt(5000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},

					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(2000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1878)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000), sdk.NewInt64Coin(denom2, 12)),
				}
			},
		},

		// ******************** Self market matching ********************

		{
			name: "match_market_self_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3750)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 3375)),
				}
			},
		},
		{
			name: "match_market_self_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4*1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375+555+777)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("555e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id5",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(3000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375+555+777)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 3000)),
				}
			},
		},
		{
			name: "try_to_match_market_self_maker_sell_taker_buy_close_with_no_change_zero_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1001)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the order will be placed but, since it cannot be matched, it will be executed with no state change
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1)), // 1000 locked by the order
				}
			},
		},
		{
			name: "match_market_self_maker_sell_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000)),
					// the account has coins to cover just one order and remainder
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375+7)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id3",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(2000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 7)),
				}
			},
		},
		{
			name: "match_market_self_maker_buy_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9000), sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "match_market_self_maker_buy_taker_sell_close_both_with_taker_partial_filling_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 8999), sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "match_market_self_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375+999)), // 999 should be filled
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(1000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 999)),
				}
			},
		},

		// ******************** Opposite limit matching ********************

		{
			name: "match_limit_opposite_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9625),
						RemainingBalance:  sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "try_to_match_limit_opposite_maker_sell_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 9999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "9625denom2 is not available, available 9624denom2",
		},
		{
			name: "match_limit_opposite_maker_sell_taker_sell_close_maker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1001)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,

						RemainingQuantity: sdkmath.NewInt(9625),
						RemainingBalance:  sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1), sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_opposite_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(999),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2664)),
				}
			},
		},
		{
			name: "match_limit_opposite_maker_sell_taker_sell_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1001)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2664), sdk.NewInt64Coin(denom2, 2)),
				}
			},
		},
		{
			name: "match_limit_opposite_maker_buy_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 381)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 26506)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10002),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:          sdkmath.NewInt(10002),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9621),
						RemainingBalance:  sdkmath.NewInt(25496),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10), sdk.NewInt64Coin(denom2, 381)),
				}
			},
		},
		{
			name: "try_to_match_limit_opposite_maker_buy_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 381)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 26490)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "25491denom1 is not available, available 25490denom1",
		},
		{
			name: "match_limit_opposite_maker_buy_taker_buy_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 4234)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2650)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(11111),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:          sdkmath.NewInt(11111),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9111),
						RemainingBalance:  sdkmath.NewInt(3472),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 650), sdk.NewInt64Coin(denom2, 762)),
				}
			},
		},
		{
			name: "match_limit_opposite_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9500),
						RemainingBalance:  sdkmath.NewInt(9500),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
				}
			},
		},
		{
			name: "match_limit_opposite_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(500),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500)),
				}
			},
		},
		{
			name: "match_limit_opposite_close_two_makers_buy_and_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 25)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 105)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 50)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5), sdk.NewInt64Coin(denom2, 50)),
				}
			},
		},
		{
			name: "match_limit_opposite_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4995)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1125),
						RemainingBalance:  sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4875)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 120), sdk.NewInt64Coin(denom2, 1835)),
				}
			},
		},
		{
			name: "match_limit_opposite_multiple_maker_sell_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2000+2000+1000+2000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1880)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("26e-1")),
						Quantity:    sdkmath.NewInt(1880),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},

					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(2000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1878)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5000), sdk.NewInt64Coin(denom2, 2)),
				}
			},
		},

		// ******************** Opposite market matching ********************

		{
			name: "match_market_opposite_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 9625)),
				}
			},
		},
		{
			name: "match_market_opposite_maker_sell_taker_sell_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 9999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 9624)),
				}
			},
		},
		{
			name: "match_market_opposite_maker_buy_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 381)),
					// ceil(10101*(1/381e-3))
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 26512)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Quantity:   sdkmath.NewInt(10101),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 25512), sdk.NewInt64Coin(denom2, 381)),
				}
			},
		},
		{
			name: "match_market_opposite_maker_buy_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 380)),
					// not enough balance to cover both orders
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100+900-1)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:    sdkmath.NewInt(900),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Quantity:   sdkmath.NewInt(1001),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:          sdkmath.NewInt(900),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(900),
						RemainingBalance:  sdkmath.NewInt(342),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 899), sdk.NewInt64Coin(denom2, 38)),
				}
			},
		},
		{
			name: "match_market_opposite_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						Quantity:   sdkmath.NewInt(999),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 999)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 2664)),
				}
			},
		},

		// ******************** Combined matching ********************

		{
			name: "match_limit_self_and_opposite_buy_close_opposite_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100+10000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// order has fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// will remain with the partial filling
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(500),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(4600),
						RemainingBalance:  sdkmath.NewInt(4600),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9955)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 45), sdk.NewInt64Coin(denom2, 5500)),
				}
			},
		},
		{
			name: "match_limit_self_and_opposite_buy_close_self_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500+5000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 220)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// order "id1", "id2" and "id3" matches, but we fill partially only "id2"
						//	with the best price and fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("22e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(800),
						RemainingBalance:  sdkmath.NewInt(400),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(5000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 200)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100), sdk.NewInt64Coin(denom2, 20)),
				}
			},
		},
		{
			name: "match_limit_self_and_opposite_sell_close_opposite_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000+19000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 825)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(1500),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(500),
						RemainingBalance:  sdkmath.NewInt(500),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(19000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 250)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1500), sdk.NewInt64Coin(denom2, 75)),
				}
			},
		},
		{
			name: "match_limit_self_and_opposite_sell_close_self_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 2100)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 2100+10000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						// order "id1", "id2" and "id3" matches, but we fill partially only "id1"
						//	with the best price and fifo priority
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(990),
						RemainingBalance:  sdkmath.NewInt(2079),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(2100),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 21)),
				}
			},
		},
		{
			name: "match_limit_self_and_opposite_buy_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100+10000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:    sdkmath.NewInt(100000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc3.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:          sdkmath.NewInt(100000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(80719),
						RemainingBalance:  sdkmath.NewInt(80719),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 18281)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10600)),
				}
			},
		},
		{
			name: "match_limit_self_and_opposite_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000+19000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 82500)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(150000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc3.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:          sdkmath.NewInt(150000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(129000),
						RemainingBalance:  sdkmath.NewInt(70950),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10500)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 21000), sdk.NewInt64Coin(denom2, 550)),
				}
			},
		},
		{
			name: "match_market_self_and_opposite_buy_close_opposite_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 100+10000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// order has fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id3",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// better price 181e-2 sell ~= 0.55 opposite buy, greater is better price
						// will remain with the partial filling
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc3.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(10000),
						Side:       types.SIDE_SELL,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(500),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(4600),
						RemainingBalance:  sdkmath.NewInt(4600),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 9955)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 45), sdk.NewInt64Coin(denom2, 5500)),
				}
			},
		},
		{
			name: "match_market_self_and_opposite_buy_close_self_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 500+5000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 200)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom2,
						QuoteDenom: denom1,
						// order "id1", "id2" and "id3" matches, but we fill partially only "id2"
						//	with the best price and fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc3.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(100),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(800),
						RemainingBalance:  sdkmath.NewInt(400),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom2,
						QuoteDenom:        denom1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(5000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 200)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
				}
			},
		},
		{
			name: "match_market_self_and_opposite_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000+19000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 75000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc3.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id4",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Quantity:   sdkmath.NewInt(150000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10500)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 21000), sdk.NewInt64Coin(denom2, 64000)),
				}
			},
		},

		// ******************** IOC matching ********************

		{
			name: "no_match_limit_sell_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1005)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 5), sdk.NewInt64Coin(denom2, 375)),
					// 3385denom2 refunded
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 3385)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled, but the order will be executed fully
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 630)),
				}
			},
		},

		// ******************** FOK matching ********************

		{
			name: "no_match_limit_sell_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1005+7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(1005),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1005),
						RemainingBalance:  sdkmath.NewInt(1005),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 3760)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000+3)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled, but the order will be executed fully, that's why we cancel the FOK
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(10000),
						RemainingBalance:  sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 3)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
		},
		{
			name: "match_limit_self_maker_sell_taker_buy_close_taker_with_full_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 10000+3)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 1005)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  denom1,
						QuoteDenom: denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled fully with the better price
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9000),
						RemainingBalance:  sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 3), sdk.NewInt64Coin(denom2, 375)),
					// expected result + not used amount
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000), sdk.NewInt64Coin(denom2, 630)),
				}
			},
		},

		// ******************** Whitelisting ********************

		{
			name: "no_match_whitelisting_limit_self_buy_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 42)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001), // initial balance
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 42), // initial balance
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(111),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:          sdkmath.NewInt(1001),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1001),
						RemainingBalance:  sdkmath.NewInt(1001),
					},
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(111),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(111),
						RemainingBalance:  sdkmath.NewInt(42),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					), // floor(376e-3 * 1001) + 1
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(
						testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
		},
		{
			name: "try_to_place_no_match_whitelisting_limit_self_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),  // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377-1), // expected to receive - 1
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "is not enough to receive 377",
		},
		{
			name: "match_whitelisting_limit_self_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 417)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),  // expected to receive
					),
					testSet.acc2.String(): sdk.NewCoins(
						// 1000 to receive when match the "id1" order, and 101 when place order to the order book
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+101), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 417),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("397e-3")),
						Quantity:    sdkmath.NewInt(1101),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("397e-3")),
						Quantity:          sdkmath.NewInt(1101),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(101),
						RemainingBalance:  sdkmath.NewInt(41),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_self_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 438)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1101), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 438),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 397),  // expected to receive
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("397e-3")),
						Quantity:    sdkmath.NewInt(1101),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("397e-3")),
						Quantity:          sdkmath.NewInt(1101),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(101),
						RemainingBalance:  sdkmath.NewInt(41),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 397),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_self_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "try_to_match_whitelisting_limit_self_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000-1),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: "is not enough to receive 1000",
		},
		{
			name: "match_whitelisting_limit_self_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376+3375),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_opposite_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+25507),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id2",
						BaseDenom:         testSet.ftDenomWhitelisting2,
						QuoteDenom:        testSet.ftDenomWhitelisting1,
						Price:             lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9625),
						RemainingBalance:  sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 25507)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_opposite_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2664),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(999),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(10000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(7336),
						RemainingBalance:  sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2664)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2751)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_opposite_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(500),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_self_and_opposite_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+19000)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 82500)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+19000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+149000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 82500),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(150000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc3.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:          sdkmath.NewInt(150000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(129000),
						RemainingBalance:  sdkmath.NewInt(70950),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500+10000)),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 21000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 550),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 129000)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_opposite_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 754+752+4+752)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4995)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2000+2000+1000+2000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4995),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1835),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1125),
						RemainingBalance:  sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 120),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1835),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2125)),
				}
			},
		},
		{
			name: "match_whitelisting_market_self_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4*1000)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4*1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777+777),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 3000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("555e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_MARKET,
						ID:         "id5",
						BaseDenom:  testSet.ftDenomWhitelisting1,
						QuoteDenom: testSet.ftDenomWhitelisting2,
						Quantity:   sdkmath.NewInt(3000),
						Side:       types.SIDE_BUY,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id4",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 3000)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 777),
					),
				}
			},
		},
		{
			name: "no_match_whitelisting_limit_sell_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // just initial amount
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_self_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.ftDenomWhitelisting1,
						QuoteDenom: testSet.ftDenomWhitelisting2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 5),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					// 3385testSet.ftDenomWhitelisting2 refunded
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3385),
					),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_whitelisting_limit_sell_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_self_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005+7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005+7),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
						// we don't need ft2 since we don't apply the matching changes
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id1",
						BaseDenom:         testSet.ftDenomWhitelisting1,
						QuoteDenom:        testSet.ftDenomWhitelisting2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:          sdkmath.NewInt(1005),
						Side:              types.SIDE_SELL,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1005),
						RemainingBalance:  sdkmath.NewInt(1005),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760)),
				}
			},
			wantWhitelistingReservedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377)),
				}
			},
		},
		{
			name: "match_limit_opposite_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 754+752+4+752)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4995)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remain no match bad price
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   denom2,
						QuoteDenom:  denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc1.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id3",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingQuantity: sdkmath.NewInt(1125),
						RemainingBalance:  sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 4875)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 120), sdk.NewInt64Coin(denom2, 1835)),
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

			testSet := genTestSet(t, sdkCtx, testApp)
			t.Logf(
				"Test set: acc1: %s, acc2: %s, acc3: %s",
				testSet.acc1, testSet.acc2, testSet.acc3,
			)

			if tt.whitelistedBalances != nil {
				for addr, coins := range tt.whitelistedBalances(testSet) {
					testApp.AssetFTKeeper.SetWhitelistedBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
				}
			}

			for addr, coins := range tt.balances(testSet) {
				testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
			}

			orderBooksIDs := make(map[uint32]struct{})
			initialOrders := tt.orders(testSet)

			ordersDenoms := make(map[string]struct{}, 0)
			for _, order := range initialOrders {
				ordersDenoms[order.BaseDenom] = struct{}{}
				ordersDenoms[order.QuoteDenom] = struct{}{}
				spendableBalancesBefore := getSpendableBalances(sdkCtx, testApp, sdk.MustAccAddressFromBech32(order.Creator))
				err := testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
				if err != nil && tt.wantErrorContains != "" {
					require.True(t, sdkerrors.IsOf(err, assetfttypes.ErrDEXLockFailed, assetfttypes.ErrWhitelistedLimitExceeded))
					require.ErrorContains(t, err, tt.wantErrorContains)
					return
				}
				require.NoError(t, err)
				assertOrderPlacementResult(t, sdkCtx, testApp, spendableBalancesBefore, order)
				orderBooksID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, order.BaseDenom, order.QuoteDenom)
				require.NoError(t, err)
				orderBooksIDs[orderBooksID] = struct{}{}
			}
			if tt.wantErrorContains != "" {
				require.Failf(t, "expected error not found", tt.wantErrorContains)
			}

			orders := make([]types.Order, 0)
			for orderBookID := range orderBooksIDs {
				orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_BUY)...)
				orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_SELL)...)
			}
			require.ElementsMatch(t, tt.wantOrders(testSet), orders)

			availableBalances := make(map[string]sdk.Coins)
			lockedBalances := make(map[string]sdk.Coins)
			whitelistingReservedBalances := make(map[string]sdk.Coins)
			for addr := range tt.balances(testSet) {
				addrBalances := testApp.BankKeeper.GetAllBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr))
				addrFTLockedBalances := sdk.NewCoins()
				for _, balance := range addrBalances {
					lockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
						sdkCtx, sdk.MustAccAddressFromBech32(addr), balance.Denom,
					)
					addrFTLockedBalances = addrFTLockedBalances.Add(lockedBalance)
					addrBalances = addrBalances.Sub(lockedBalance)
				}

				addrFTWhitelistingReservedBalances := sdk.NewCoins()
				for denom := range ordersDenoms {
					addrFTWhitelistingReservedBalance := testApp.AssetFTKeeper.GetDEXWhitelistingReservedBalance(
						sdkCtx, sdk.MustAccAddressFromBech32(addr), denom,
					)
					addrFTWhitelistingReservedBalances = addrFTWhitelistingReservedBalances.Add(addrFTWhitelistingReservedBalance)
				}

				availableBalances[addr] = addrBalances
				lockedBalances[addr] = addrFTLockedBalances
				whitelistingReservedBalances[addr] = addrFTWhitelistingReservedBalances
			}
			availableBalances = removeEmptyBalances(availableBalances)
			lockedBalances = removeEmptyBalances(lockedBalances)
			whitelistingReservedBalances = removeEmptyBalances(whitelistingReservedBalances)

			wantAvailableBalances := tt.wantAvailableBalances(testSet)
			require.True(
				t,
				reflect.DeepEqual(wantAvailableBalances, availableBalances),
				fmt.Sprintf("want: %v, got: %v", wantAvailableBalances, availableBalances),
			)

			// by default must be empty
			wantWhitelistingReservedBalances := make(map[string]sdk.Coins)
			if tt.wantWhitelistingReservedBalances != nil {
				wantWhitelistingReservedBalances = tt.wantWhitelistingReservedBalances(testSet)
			}

			require.True(
				t,
				reflect.DeepEqual(wantWhitelistingReservedBalances, whitelistingReservedBalances),
				fmt.Sprintf("want: %v, got: %v", wantWhitelistingReservedBalances, whitelistingReservedBalances),
			)

			// check that balance locked in the orders correspond the balance locked in the asset ft
			orderLockedBalances := make(map[string]sdk.Coins)
			for _, order := range orders {
				coins, ok := orderLockedBalances[order.Creator]
				if !ok {
					coins = sdk.NewCoins()
				}
				coins = coins.Add(sdk.NewCoin(order.GetSpendDenom(), order.RemainingBalance))
				orderLockedBalances[order.Creator] = coins
			}
			orderLockedBalances = removeEmptyBalances(orderLockedBalances)
			require.True(
				t,
				reflect.DeepEqual(lockedBalances, orderLockedBalances),
				fmt.Sprintf("want: %v, got: %v", lockedBalances, orderLockedBalances),
			)

			cancelAllOrdersAndAssertState(t, sdkCtx, testApp)
		})
	}
}

func genTestSet(t *testing.T, sdkCtx sdk.Context, testApp *simapp.App) TestSet {
	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)
	acc3, _ := testApp.GenAccount(sdkCtx)

	issuer, _ := testApp.GenAccount(sdkCtx)

	ftDenomWhitelisting1, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
		Issuer:        issuer,
		Subunit:       "ftwhitelisting1",
		Symbol:        "FTWHITELISTING1",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	})
	require.NoError(t, err)

	ftDenomWhitelisting2, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
		Issuer:        issuer,
		Subunit:       "ftwhitelisting2",
		Symbol:        "FTWHITELISTING2",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	})
	require.NoError(t, err)

	testSet := TestSet{
		acc1: acc1,
		acc2: acc2,
		acc3: acc3,

		issuer:               issuer,
		ftDenomWhitelisting1: ftDenomWhitelisting1,
		ftDenomWhitelisting2: ftDenomWhitelisting2,
	}

	return testSet
}

func removeEmptyBalances(balances map[string]sdk.Coins) map[string]sdk.Coins {
	for addr, balance := range balances {
		if balance.IsZero() {
			delete(balances, addr)
		}
	}

	return balances
}

func assertOrderPlacementResult(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	spendableBalancesBefore map[string]sdkmath.Int,
	order types.Order,
) {
	creator := sdk.MustAccAddressFromBech32(order.Creator)
	spendableBalancesAfter := getSpendableBalances(sdkCtx, testApp, creator)

	spDenom := order.GetSpendDenom()
	spBalanceAmtBefore, ok := spendableBalancesBefore[spDenom]
	if !ok {
		spBalanceAmtBefore = sdkmath.ZeroInt()
	}
	spBalanceAmtAfter, ok := spendableBalancesAfter[spDenom]
	if !ok {
		spBalanceAmtAfter = sdkmath.ZeroInt()
	}

	recvDenom := order.GetReceiveDenom()
	recvBalanceAmtBefore, ok := spendableBalancesBefore[recvDenom]
	if !ok {
		recvBalanceAmtBefore = sdkmath.ZeroInt()
	}
	recvBalanceAmtAfter, ok := spendableBalancesAfter[recvDenom]
	if !ok {
		recvBalanceAmtAfter = sdkmath.ZeroInt()
	}

	storedOrder, err := testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, creator, order.ID)
	if err != nil {
		require.ErrorIs(t, err, types.ErrRecordNotFound)
		t.Logf("Order not found in the order book.")

		spentAmt := spBalanceAmtBefore.Sub(spBalanceAmtAfter)
		require.False(t, spentAmt.IsNegative())
		receivedAmt := recvBalanceAmtAfter.Sub(recvBalanceAmtBefore)
		require.False(t, receivedAmt.IsNegative())

		// assertSpentAndReceivedAmounts(t, spentAmt, receivedAmt)
		if order.Type == types.ORDER_TYPE_LIMIT {
			// limit order is matched partially or fully and closed, so the remaining balance is refunded
			assertExecutionPrice(t, order, spentAmt, receivedAmt)
		}
		assertFilledQuantity(t, order, spentAmt, receivedAmt)

		return
	}

	t.Logf("Order found in the order book.")

	if order.Type != types.ORDER_TYPE_LIMIT {
		t.Fatalf("Saved not market order, type: %s", order.Type.String())
	}
	if order.TimeInForce != types.TIME_IN_FORCE_GTC {
		t.Fatalf("Saved not GTC order, time in force: %s", order.TimeInForce.String())
	}

	// limit order is matched partially but the remaining balance is not refunded
	spentAmt := spBalanceAmtBefore.Sub(spBalanceAmtAfter).Sub(storedOrder.RemainingBalance)
	receivedAmt := recvBalanceAmtAfter.Sub(recvBalanceAmtBefore)

	// order is in the store and executed partially
	if !order.Quantity.Equal(storedOrder.RemainingQuantity) {
		require.False(t, receivedAmt.IsNegative())
		assertExecutionPrice(t, order, spentAmt, receivedAmt)
	} else {
		// if quantity is not filled, nothing is spent
		require.True(t, spentAmt.IsZero())
	}
}

func assertExecutionPrice(t *testing.T, order types.Order, spendAmt, receiveAmt sdkmath.Int) {
	orderPriceRat := order.Price.Rat()
	var executionPriceRat *big.Rat
	if order.Side == types.SIDE_BUY {
		if receiveAmt.IsZero() {
			return
		}
		executionPriceRat = cbig.NewRatFromBigInts(spendAmt.BigInt(), receiveAmt.BigInt())
		require.True(
			t,
			cbig.RatLTE(executionPriceRat, orderPriceRat),
			fmt.Sprintf("orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String()),
		)
	} else {
		if spendAmt.IsZero() {
			return
		}
		executionPriceRat = cbig.NewRatFromBigInts(receiveAmt.BigInt(), spendAmt.BigInt())
		require.True(
			t,
			cbig.RatGTE(executionPriceRat, orderPriceRat),
			fmt.Sprintf("orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String()),
		)
	}
	t.Logf(
		"Execution prices, side:%s, order: %s, execution: %s, ",
		order.Side.String(), orderPriceRat.String(), executionPriceRat.String(),
	)
}

func assertFilledQuantity(t *testing.T, order types.Order, spendAmt, receiveAmt sdkmath.Int) {
	var filledQuantity sdkmath.Int
	if order.Side == types.SIDE_BUY {
		filledQuantity = receiveAmt
	} else {
		filledQuantity = spendAmt
	}
	t.Logf(
		"Filled qunitities, side:%s, orderTimeInForce: %s, orderQuantity: %s, filledQuantity: %s",
		order.Side.String(), order.TimeInForce.String(), order.Quantity.String(), filledQuantity.String(),
	)
	if order.Type == types.ORDER_TYPE_MARKET {
		if order.TimeInForce == types.TIME_IN_FORCE_FOK {
			// it's possible that order is machined with the order of the same account, so the filled quantity
			// includes multiple increased that's why we use GTE, but not EQ
			require.True(t, order.Quantity.GTE(filledQuantity))
		}
	}
}

func getSpendableBalances(sdkCtx sdk.Context, testApp *simapp.App, acc sdk.AccAddress) map[string]sdkmath.Int {
	balances := testApp.BankKeeper.GetAllBalances(sdkCtx, acc)
	spendableBalances := make(map[string]sdkmath.Int)
	for _, balance := range balances {
		frozenAmt := testApp.AssetFTKeeper.GetFrozenBalance(sdkCtx, acc, balance.Denom).Amount
		dexLockedAmt := testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, balance.Denom).Amount
		// can be negative
		spendableBalances[balance.Denom] = balance.Amount.Sub(frozenAmt).Sub(dexLockedAmt)
	}

	return spendableBalances
}

func cancelAllOrdersAndAssertState(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
) {
	t.Helper()

	orders, _, err := testApp.DEXKeeper.GetOrdersWithSequence(sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)

	t.Logf("Cancelling all orders, count %d", len(orders))

	accounts := make(map[string]struct{})
	denoms := make(map[string]struct{})
	for _, order := range orders {
		require.NoError(t, testApp.DEXKeeper.CancelOrder(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Order.Creator), order.Order.ID),
		)
		accounts[order.Order.Creator] = struct{}{}
		denoms[order.Order.BaseDenom] = struct{}{}
		denoms[order.Order.QuoteDenom] = struct{}{}
	}
	for acc := range accounts {
		for denom := range denoms {
			dexLockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexLockedBalance.IsZero(),
				fmt.Sprintf("denom: %s, acc: %s, dexLockedBalance: %s", denom, acc, dexLockedBalance.String()),
			)

			dexWhitelistingReservedBalance := testApp.AssetFTKeeper.GetDEXWhitelistingReservedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexWhitelistingReservedBalance.IsZero(),
				fmt.Sprintf(
					"denom: %s, acc: %s, dexWhitelistingReservedBalance: %s",
					denom, acc, dexWhitelistingReservedBalance.String(),
				),
			)

			accountDenomOrdersCount, err := testApp.DEXKeeper.GetAccountDenomOrdersCount(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.NoError(t, err)
			require.Zero(t, accountDenomOrdersCount)
		}
	}
}

func cancelAllOrdersAndAssertState(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
) {
	t.Helper()

	orders, _, err := testApp.DEXKeeper.GetOrdersWithSequence(sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)

	t.Logf("Cancelling all orders, count %d", len(orders))

	accounts := make(map[string]struct{})
	denoms := make(map[string]struct{})
	for _, order := range orders {
		require.NoError(t, testApp.DEXKeeper.CancelOrder(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Order.Creator), order.Order.ID),
		)
		accounts[order.Order.Creator] = struct{}{}
		denoms[order.Order.BaseDenom] = struct{}{}
		denoms[order.Order.QuoteDenom] = struct{}{}
	}
	for acc := range accounts {
		for denom := range denoms {
			dexLockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexLockedBalance.IsZero(),
				fmt.Sprintf("denom: %s, acc: %s, dexLockedBalance: %s", denom, acc, dexLockedBalance.String()),
			)

			dexWhitelistingReservedBalance := testApp.AssetFTKeeper.GetDEXWhitelistingReservedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexWhitelistingReservedBalance.IsZero(),
				fmt.Sprintf(
					"denom: %s, acc: %s, dexWhitelistingReservedBalance: %s",
					denom, acc, dexWhitelistingReservedBalance.String(),
				),
			)

			accountDenomOrdersCount, err := testApp.DEXKeeper.GetAccountDenomOrdersCount(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.NoError(t, err)
			require.Zero(t, accountDenomOrdersCount)
		}
	}
}
