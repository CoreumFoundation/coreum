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

	orderReserve sdk.Coin
}

func (t TestSet) orderReserveTimes(times int64) sdk.Coin {
	return sdk.NewCoin(t.orderReserve.Denom, t.orderReserve.Amount.MulRaw(times))
}

func TestKeeper_MatchOrders(t *testing.T) {
	tests := []struct {
		name                          string
		balances                      func(testSet TestSet) map[string]sdk.Coins
		whitelistedBalances           func(testSet TestSet) map[string]sdk.Coins
		orders                        func(testSet TestSet) []types.Order
		wantOrders                    func(testSet TestSet) []types.Order
		wantAvailableBalances         func(testSet TestSet) map[string]sdk.Coins
		wantExpectedToReceiveBalances func(testSet TestSet) map[string]sdk.Coins
		wantErrorContains             string
	}{
		// ******************** No matching ********************

		{
			name: "no_match_limit_directOB_and_invertedOB_buy_and_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 2659),
						sdk.NewInt64Coin(denom2, 375),
					),
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
			name: "try_to_match_limit_directOB_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 999),
					),
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

		// ******************** Direct OB limit matching ********************

		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 3761),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 1),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker_same_account",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 3761),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 376),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_directOB_maker_sell_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 3758),
						testSet.orderReserve,
					),
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
			// we fill the id1 first, so the used balance from id2 is 1000 * 375e-3 = 1000 * 375e-3 = 375
			// to fill remaining part we need (10000 - 1000) * 376e-3 = 3384, so total expected to send 3384 + 375 = 3759
			wantErrorContains: "3759denom2 is not available, available 3758denom2",
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1005),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 3760),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5),
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 1),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 377),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 2),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1005),
					),
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
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
						// 630 = (1005*1) - (1000*0.375). Where 1005 is amount locked and 375 amount spent.
						sdk.NewInt64Coin(denom2, 630),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 376)),
				}
			},
		},
		{
			name: "try_to_match_limit_directOB_maker_buy_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 9999),
					),
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
			wantErrorContains: "10000denom1 is not available, available 9999denom1",
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 3760)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
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
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 376),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
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
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 50),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 50),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_close_two_makers_sell_and_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 50),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 50),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 60),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 25),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 25),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100),
						sdk.NewInt64Coin(denom2, 10),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_close_two_makers_buy_and_and_taker_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 50),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 50),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 200),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 100),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_multiple_maker_buy_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5000),
					),
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
					// remains unmatched price is too low
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
					// the part of the order should remain. Order sequence respected.
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
						// partially executed
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(376),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 5000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1882), // 1882 = (754+752+4+752)-4-376
					),
				}
			},
		},
		{
			name: "match_limit_directOB_multiple_maker_sell_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom1, 2000+2000+1000+2000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1890),
					),
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
					// remains unmatched price is too low
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 1878),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5000),
						sdk.NewInt64Coin(denom2, 12),
					),
				}
			},
		},

		// ******************** Direct OB market matching ********************

		{
			name: "match_market_directOB_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(denom2, 3750),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						// Locked full balance and returned remainer: 3750 - 375 = 3375
						sdk.NewInt64Coin(denom2, 3375),
					),
				}
			},
		},
		{
			name: "match_market_directOB_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom1, 4*1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(denom2, 375+555+777),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(3),
						sdk.NewInt64Coin(denom2, 375+555+777),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 3000),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_sell_taker_buy_close_with_no_change_zero_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1001),
					),
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
			name: "match_market_directOB_maker_sell_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 2000),
					),
					// the account has coins to cover just one order and remainder,
					// also not reserve is needed for the market order
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 375+7),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 7),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_buy_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 9000),
						sdk.NewInt64Coin(denom2, 376),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_buy_taker_sell_close_both_with_taker_partial_filling_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 9999),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 8999),
						sdk.NewInt64Coin(denom2, 376),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 375+999), // 999 should be filled
					),
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

		// ******************** Inverted OB limit matching ********************

		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_invertedOB_maker_sell_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 9999),
					),
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
			wantErrorContains: "10000denom2 is not available, available 9999denom2",
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_maker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1001)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1),
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 999),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 2664),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1001),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 2664), // 2664 = 999 / 0.375, so 999 matched for price 0.375.
						sdk.NewInt64Coin(denom2, 2),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 381)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 26506),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 10),
						sdk.NewInt64Coin(denom2, 381),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_invertedOB_maker_buy_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 381),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 26490),
					),
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
			wantErrorContains: "26491denom1 is not available, available 26490denom1",
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_buy_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 4234),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 2650),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 2000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 650),
						sdk.NewInt64Coin(denom2, 762),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 500),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 500),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_close_two_makers_buy_and_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 25),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 25),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 105),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 50)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 50)),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5),
						sdk.NewInt64Coin(denom2, 50),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 4995),
					),
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
					// remains unmatched price is too low
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 120),
						sdk.NewInt64Coin(denom2, 1835),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_sell_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom1, 2000+2000+1000+2000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 1880),
					),
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
					// remains unmatched price is too low
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 1878),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5000),
						sdk.NewInt64Coin(denom2, 2),
					),
				}
			},
		},

		// ******************** Inverted OB market matching ********************

		{
			name: "match_market_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(denom2, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 9625),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_sell_taker_sell_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 9999),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 9624),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_buy_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 381),
					),
					// ceil(10101*(1/381e-3))
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 26512),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 25512),
						sdk.NewInt64Coin(denom2, 381),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_buy_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 380),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 899),
						sdk.NewInt64Coin(denom2, 38),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 999),
					),
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
			name: "match_limit_directOB_and_invertedOB_buy_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
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
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
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
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
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
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 9955),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 45),
						sdk.NewInt64Coin(denom2, 5500),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_buy_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 500+5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 220),
					),
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
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 200),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 100),
						sdk.NewInt64Coin(denom2, 20),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 825),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 250),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1500),
						sdk.NewInt64Coin(denom2, 75),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 2100),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 2100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 10),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 21),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_buy_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 100000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 18281),
					),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(denom2, 10600)),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 82500)),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 10500),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 21000),
						sdk.NewInt64Coin(denom2, 550),
					),
				}
			},
		},
		{
			name: "match_market_directOB_and_invertedOB_buy_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 10000),
					),
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
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
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
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
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
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 9955),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 45),
						sdk.NewInt64Coin(denom2, 5500),
					),
				}
			},
		},
		{
			name: "match_market_directOB_and_invertedOB_buy_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 500+5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 200),
					),
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
			name: "match_market_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 75000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom2, 10500),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 21000),
						sdk.NewInt64Coin(denom2, 64000),
					),
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
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1005),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 3760),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 5),
						sdk.NewInt64Coin(denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 1000),
						// 3385denom2 refunded
						sdk.NewInt64Coin(denom2, 3385),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 1005),
					),
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
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 1005+7),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 3760),
					),
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
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000+3)),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 1005),
					),
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
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_full_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 10000+3),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom2, 1005),
					),
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
			name: "no_match_whitelisting_limit_directOB_buy_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 42),
					),
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
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
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
			name: "try_to_place_no_match_whitelisting_limit_directOB_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
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
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 417),
					),
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
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 438),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
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
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 397),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
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
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "try_to_match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
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
			name: "match_whitelisting_limit_directOB_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 25507)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2664),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2751)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 82500),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 21000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 550),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 129000)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4995),
					),
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
					// remains unmatched price is too low
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
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 120),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1835),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2125)),
				}
			},
		},
		{
			name: "match_whitelisting_market_directOB_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4*1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(3),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 3000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
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
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
					),
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
						testSet.orderReserve,
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
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
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
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005+7),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(
						testSet.ftDenomWhitelisting2, 3760),
					),
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
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377)),
				}
			},
		},

		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker_with_zero_filled_quantity",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1),   // expected to receive
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
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
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// can't fill since 111 * 376e-5 ~= 0.41736
						Quantity:    sdkmath.NewInt(111),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("1e1")),
						Quantity:    sdkmath.NewInt(1000),
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
						Price:             lo.ToPtr(types.MustNewPriceFromString("1e1")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_buy_taker_sell_close_taker_with_zero_filled_quantity",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 4),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 4),    // initial
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1),   // expected to receive
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
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.ftDenomWhitelisting1,
						QuoteDenom: testSet.ftDenomWhitelisting2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// can't fill since 111 * 376e-5 ~= 0.41736
						Quantity:    sdkmath.NewInt(111),
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
						Price:             lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						Quantity:          sdkmath.NewInt(1000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(1000),
						RemainingBalance:  sdkmath.NewInt(4),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(denom2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 4995),
					),
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
					// remains unmatched price is too low
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
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(denom1, 120),
						sdk.NewInt64Coin(denom2, 1835),
					),
				}
			},
		},

		{
			name: "no_match_limit_directOB_and_invertedOB_buy_and_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 1000),
						sdk.NewInt64Coin(denom2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(denom1, 2659),
						sdk.NewInt64Coin(denom2, 375),
					),
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
			name: "match_limit_directOB_maker_sell_taker_buy_ten_orders",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(10),
						sdk.NewInt64Coin(denom1, 100000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(1),
						sdk.NewInt64Coin(denom2, 98991560000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				orders := make([]types.Order, 0)
				for i := range 10 {
					orders = append(orders, types.Order{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          fmt.Sprintf("id%d", i),
						BaseDenom:   denom1,
						QuoteDenom:  denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString(fmt.Sprintf("1%d1e-1", i))),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					})
				}
				orders = append(orders, types.Order{
					Creator:     testSet.acc2.String(),
					Type:        types.ORDER_TYPE_LIMIT,
					ID:          "id101",
					BaseDenom:   denom1,
					QuoteDenom:  denom2,
					Price:       lo.ToPtr(types.MustNewPriceFromString("9999")),
					Quantity:    sdkmath.NewInt(10_000_000),
					Side:        types.SIDE_BUY,
					TimeInForce: types.TIME_IN_FORCE_GTC,
				})
				return orders
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:           testSet.acc2.String(),
						Type:              types.ORDER_TYPE_LIMIT,
						ID:                "id101",
						BaseDenom:         denom1,
						QuoteDenom:        denom2,
						Price:             lo.ToPtr(types.MustNewPriceFromString("9999")),
						Quantity:          sdkmath.NewInt(10_000_000),
						Side:              types.SIDE_BUY,
						TimeInForce:       types.TIME_IN_FORCE_GTC,
						RemainingQuantity: sdkmath.NewInt(9900000),
						RemainingBalance:  sdkmath.NewInt(98990100000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(10),
						sdk.NewInt64Coin(denom2, 1460000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(denom1, 100000),
					),
				}
			},
		},
	}
	for _, tt := range tests {
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
			for i, order := range initialOrders {
				ordersDenoms[order.BaseDenom] = struct{}{}
				ordersDenoms[order.QuoteDenom] = struct{}{}
				availableBalancesBefore, err := getAvailableBalances(sdkCtx, testApp, sdk.MustAccAddressFromBech32(order.Creator))
				require.NoError(t, err)

				// use new event manager for each order
				sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
				gasBefore := sdkCtx.GasMeter().GasConsumed()
				err = testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
				if err != nil && tt.wantErrorContains != "" {
					require.True(t, sdkerrors.IsOf(
						err,
						assetfttypes.ErrDEXInsufficientSpendableBalance, assetfttypes.ErrWhitelistedLimitExceeded,
					))
					require.ErrorContains(t, err, tt.wantErrorContains)
					return
				}
				gasAfter := sdkCtx.GasMeter().GasConsumed()
				t.Logf("Used gas for order %d placement: %d", i, gasAfter-gasBefore)
				require.NoError(t, err)
				assertOrderPlacementResult(t, sdkCtx, testApp, availableBalancesBefore, order)
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
			wantOrders := tt.wantOrders(testSet)
			// set order reserve and order sequence for all orders
			wantOrders = fillReserveAndOrderSequence(t, sdkCtx, testApp, wantOrders)
			require.ElementsMatch(t, wantOrders, orders)

			availableBalances := make(map[string]sdk.Coins)
			lockedBalances := make(map[string]sdk.Coins)
			expectedToReceiveBalances := make(map[string]sdk.Coins)
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

				addrFTExpectedToReceiveBalances := sdk.NewCoins()
				for denom := range ordersDenoms {
					addrFTExpectedToReceiveBalance := testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
						sdkCtx, sdk.MustAccAddressFromBech32(addr), denom,
					)
					addrFTExpectedToReceiveBalances = addrFTExpectedToReceiveBalances.Add(addrFTExpectedToReceiveBalance)
				}

				availableBalances[addr] = addrBalances
				lockedBalances[addr] = addrFTLockedBalances
				expectedToReceiveBalances[addr] = addrFTExpectedToReceiveBalances
			}
			availableBalances = removeEmptyBalances(availableBalances)
			lockedBalances = removeEmptyBalances(lockedBalances)
			expectedToReceiveBalances = removeEmptyBalances(expectedToReceiveBalances)

			wantAvailableBalances := tt.wantAvailableBalances(testSet)
			require.True(
				t,
				reflect.DeepEqual(wantAvailableBalances, availableBalances),
				"want: %v, got: %v", wantAvailableBalances, availableBalances,
			)

			// by default must be empty
			wantExpectedToReceiveBalances := make(map[string]sdk.Coins)
			if tt.wantExpectedToReceiveBalances != nil {
				wantExpectedToReceiveBalances = tt.wantExpectedToReceiveBalances(testSet)
			}

			require.True(
				t,
				reflect.DeepEqual(wantExpectedToReceiveBalances, expectedToReceiveBalances),
				"want: %v, got: %v", wantExpectedToReceiveBalances, expectedToReceiveBalances,
			)

			// check that balance locked in the orders correspond the balance locked in the asset ft
			orderLockedBalances := make(map[string]sdk.Coins)
			for _, order := range orders {
				coins, ok := orderLockedBalances[order.Creator]
				if !ok {
					coins = sdk.NewCoins()
				}
				coins = coins.Add(sdk.NewCoin(order.GetSpendDenom(), order.RemainingBalance))
				params, err := testApp.DEXKeeper.GetParams(sdkCtx)
				require.NoError(t, err)
				// add reserve for each order
				coins = coins.Add(params.OrderReserve)
				orderLockedBalances[order.Creator] = coins
			}
			orderLockedBalances = removeEmptyBalances(orderLockedBalances)
			require.True(
				t,
				reflect.DeepEqual(lockedBalances, orderLockedBalances),
				"want: %v, got: %v", lockedBalances, orderLockedBalances,
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

	param, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)

	testSet := TestSet{
		acc1: acc1,
		acc2: acc2,
		acc3: acc3,

		issuer:               issuer,
		ftDenomWhitelisting1: ftDenomWhitelisting1,
		ftDenomWhitelisting2: ftDenomWhitelisting2,

		orderReserve: param.OrderReserve,
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

func fillReserveAndOrderSequence(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	orders []types.Order,
) []types.Order {
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	orderReserve := params.OrderReserve
	for i, order := range orders {
		storedOrder, err := testApp.DEXKeeper.GetOrderByAddressAndID(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID,
		)
		require.NoError(t, err)
		require.Positive(t, storedOrder.Sequence)
		orders[i].Sequence = storedOrder.Sequence
		orders[i].Reserve = orderReserve
	}

	return orders
}

func assertOrderPlacementResult(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	availableBalancesBefore map[string]sdkmath.Int,
	order types.Order,
) {
	events := readOrderEvents(t, sdkCtx)
	assertPlacementEvents(t, order, events)

	sentAmt, receivedAmt := assetOrderSentReceivedAmounts(
		t, sdkCtx, testApp, availableBalancesBefore, order, events,
	)
	storedOrder, err := testApp.DEXKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID,
	)
	if err != nil {
		require.ErrorIs(t, err, types.ErrRecordNotFound)
		t.Logf("Order not found in the order book.")
		assertFilledQuantity(t, order, sentAmt, receivedAmt)
		return
	}

	t.Logf("Order found in the order book.")
	require.NotNil(t, events.OrderCreated)
	require.Equal(t, types.EventOrderCreated{
		Creator:           storedOrder.Creator,
		ID:                storedOrder.ID,
		Sequence:          events.OrderPlaced.Sequence,
		RemainingQuantity: storedOrder.RemainingQuantity,
		RemainingBalance:  storedOrder.RemainingBalance,
	}, *events.OrderCreated)

	if order.Type != types.ORDER_TYPE_LIMIT {
		t.Fatalf("Saved not market order, type: %s", order.Type.String())
	}
	if order.TimeInForce != types.TIME_IN_FORCE_GTC {
		t.Fatalf("Saved not GTC order, time in force: %s", order.TimeInForce.String())
	}
}

func assertPlacementEvents(t *testing.T, order types.Order, events OrderPlacementEvents) {
	require.Positive(t, events.OrderPlaced.Sequence)
	require.Equal(t, types.EventOrderPlaced{
		Creator:  order.Creator,
		ID:       order.ID,
		Sequence: events.OrderPlaced.Sequence,
	}, events.OrderPlaced)

	// set initial quantity
	expectedRemainingQuantity := order.Quantity

	takerSentAmt := sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt())
	takerReceivedAmt := sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt())
	makerSentAmt := sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt())
	makerReceivedAmt := sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt())
	for _, reducedEvt := range events.OrdersReduced {
		// is taker
		if events.OrderPlaced.Sequence == reducedEvt.Sequence {
			require.True(t, takerSentAmt.IsZero())
			require.True(t, takerReceivedAmt.IsZero())
			takerSentAmt = reducedEvt.SentCoin
			takerReceivedAmt = reducedEvt.ReceivedCoin

			if order.Side == types.SIDE_BUY {
				expectedRemainingQuantity = expectedRemainingQuantity.Sub(reducedEvt.ReceivedCoin.Amount)
			} else {
				expectedRemainingQuantity = expectedRemainingQuantity.Sub(reducedEvt.SentCoin.Amount)
			}
			continue
		}
		makerSentAmt = makerSentAmt.Add(reducedEvt.SentCoin)
		makerReceivedAmt = makerReceivedAmt.Add(reducedEvt.ReceivedCoin)
	}
	require.Equal(t, takerSentAmt.String(), makerReceivedAmt.String())
	require.Equal(t, makerSentAmt.String(), takerReceivedAmt.String())
	if events.OrderCreated != nil {
		expectedRemainingBalance, err := types.ComputeLimitOrderLockedBalance(
			order.Side, order.BaseDenom, order.QuoteDenom, expectedRemainingQuantity, *order.Price,
		)
		require.NoError(t, err)
		require.Equal(t, types.EventOrderCreated{
			Creator:           order.Creator,
			ID:                order.ID,
			Sequence:          events.OrderPlaced.Sequence,
			RemainingQuantity: expectedRemainingQuantity,
			RemainingBalance:  expectedRemainingBalance.Amount,
		}, *events.OrderCreated)
	}
}

func assetOrderSentReceivedAmounts(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	availableBalancesBefore map[string]sdkmath.Int,
	order types.Order,
	events OrderPlacementEvents,
) (sdkmath.Int, sdkmath.Int) {
	var (
		orderSentAmt     = sdkmath.ZeroInt()
		orderReceivedAmt = sdkmath.ZeroInt()
	)
	if len(events.OrdersReduced) > 0 {
		require.GreaterOrEqual(t, len(events.OrdersReduced), 2)
		currentOrderReducedEvent, ok := events.getOrderReduced(order.Creator, order.ID)
		require.True(t, ok)
		matchedSent := sdk.NewCoin(currentOrderReducedEvent.ReceivedCoin.Denom, sdkmath.ZeroInt())
		matchedReceived := sdk.NewCoin(currentOrderReducedEvent.SentCoin.Denom, sdkmath.ZeroInt())
		for _, evt := range events.OrdersReduced {
			if evt.Creator == order.Creator && evt.ID == order.ID {
				continue
			}
			matchedSent = matchedSent.Add(evt.SentCoin)
			matchedReceived = matchedReceived.Add(evt.ReceivedCoin)
		}
		require.Equal(t, currentOrderReducedEvent.SentCoin.String(), matchedReceived.String())
		require.Equal(t, currentOrderReducedEvent.ReceivedCoin.String(), matchedSent.String())

		orderSentAmt = currentOrderReducedEvent.SentCoin.Amount
		orderReceivedAmt = currentOrderReducedEvent.ReceivedCoin.Amount

		if order.Type == types.ORDER_TYPE_LIMIT {
			assertExecutionPrice(
				t,
				order,
				orderSentAmt,
				orderReceivedAmt,
			)
		}
	}

	// check that balance used amount either sent or locked in the order
	orderUsedAmt := orderSentAmt
	if events.OrderCreated != nil {
		// locked balance
		orderUsedAmt = orderUsedAmt.Add(events.OrderCreated.RemainingBalance)
	}

	creator := sdk.MustAccAddressFromBech32(order.Creator)
	// check the balances updated
	availableAmtBefore, ok := availableBalancesBefore[order.GetSpendDenom()]
	if !ok {
		availableAmtBefore = sdkmath.ZeroInt()
	}
	availableBalancesAfter, err := getAvailableBalances(sdkCtx, testApp, creator)
	require.NoError(t, err)
	availableBalanceAmtAfter, ok := availableBalancesAfter[order.GetSpendDenom()]
	if !ok {
		availableBalanceAmtAfter = sdkmath.ZeroInt()
	}
	// balanceUsedAmt includes locked and frozen balances
	balanceUsedAmt := availableAmtBefore.Sub(availableBalanceAmtAfter)
	require.False(t, balanceUsedAmt.IsNegative())

	// adjust with the direct OB matched orders
	for _, evt := range events.OrdersReduced {
		if evt.Creator != order.Creator {
			continue
		}
		if evt.ID == order.ID {
			continue
		}
		if evt.SentCoin.Denom != order.GetSpendDenom() {
			balanceUsedAmt = balanceUsedAmt.Add(evt.ReceivedCoin.Amount)
		} else {
			// we can't same coin as `order.GetSpendDenom()` with the direct OB match
			t.Fatalf("Unexpected to sent coin: %s", evt.SentCoin)
		}
	}

	require.Equal(t, balanceUsedAmt.String(), orderUsedAmt.String())

	return orderSentAmt, orderReceivedAmt
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
			"orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String(),
		)
	} else {
		if spendAmt.IsZero() {
			return
		}
		executionPriceRat = cbig.NewRatFromBigInts(receiveAmt.BigInt(), spendAmt.BigInt())
		require.True(
			t,
			cbig.RatGTE(executionPriceRat, orderPriceRat),
			"orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String(),
		)
	}
	t.Logf(
		"Execution prices, side:%s, order: %s, execution: %s, ",
		order.Side.String(), orderPriceRat.String(), executionPriceRat.String(),
	)
}

func assertFilledQuantity(t *testing.T, order types.Order, sent, receiveAmt sdkmath.Int) {
	var filledQuantity sdkmath.Int
	if order.Side == types.SIDE_BUY {
		filledQuantity = receiveAmt
	} else {
		filledQuantity = sent
	}
	t.Logf(
		"Filled qunitities, side:%s, orderTimeInForce: %s, orderQuantity: %s, filledQuantity: %s",
		order.Side.String(), order.TimeInForce.String(), order.Quantity.String(), filledQuantity.String(),
	)
	// check that we never exceed the order's quantity
	require.True(t, order.Quantity.GTE(filledQuantity))
}

func getAvailableBalances(sdkCtx sdk.Context, testApp *simapp.App, acc sdk.AccAddress) (map[string]sdkmath.Int, error) {
	balances := testApp.BankKeeper.GetAllBalances(sdkCtx, acc)
	spendableBalances := make(map[string]sdkmath.Int)
	for _, balance := range balances {
		frozenBalance, err := testApp.AssetFTKeeper.GetFrozenBalance(sdkCtx, acc, balance.Denom)
		if err != nil {
			return nil, err
		}
		frozenAmt := frozenBalance.Amount
		dexLockedAmt := testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, balance.Denom).Amount
		// can be negative
		spendableBalances[balance.Denom] = balance.Amount.Sub(frozenAmt).Sub(dexLockedAmt)
	}

	return spendableBalances, nil
}

func cancelAllOrdersAndAssertState(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
) {
	t.Helper()

	orders, _, err := testApp.DEXKeeper.GetAccountsOrders(sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)

	t.Logf("Cancelling all orders, count %d", len(orders))

	accounts := make(map[string]struct{})
	denoms := make(map[string]struct{})
	for _, order := range orders {
		require.NoError(t, testApp.DEXKeeper.CancelOrder(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID),
		)
		accounts[order.Creator] = struct{}{}
		denoms[order.BaseDenom] = struct{}{}
		denoms[order.QuoteDenom] = struct{}{}
	}
	for acc := range accounts {
		for denom := range denoms {
			dexLockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexLockedBalance.IsZero(),
				"denom: %s, acc: %s, dexLockedBalance: %s", denom, acc, dexLockedBalance.String(),
			)

			dexExpectedToReceiveBalance := testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexExpectedToReceiveBalance.IsZero(),
				"denom: %s, acc: %s, dexExpectedToReceiveBalance: %s",
				denom, acc, dexExpectedToReceiveBalance.String(),
			)

			accountDenomOrdersCount, err := testApp.DEXKeeper.GetAccountDenomOrdersCount(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.NoError(t, err)
			require.Zero(t, accountDenomOrdersCount)
		}
	}
}
