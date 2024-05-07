package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/x/dex/keeper"
)

func TestMatching(t *testing.T) {
	const (
		sender1 = "sender1"
		sender2 = "sender2"
		sender3 = "sender3"
		sender4 = "sender4"

		denom1 = "denom1"
		denom2 = "denom2"
	)
	type testCase struct {
		name               string
		newOrders          []keeper.Order
		expectedOrderBooks map[string][]keeper.Order
		expectedBalances   map[string]sdk.Coins
	}
	testCases := []testCase{
		{
			name: "no_match",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(50),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:      sender3,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(10),
					Price:        sdkmath.LegacyMustNewDecFromStr("5.2"),
				},
				{
					Account:      sender2,
					ID:           "order3",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(20),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.21"),
				},
				{
					Account:      sender3,
					ID:           "order4",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(30),
					Price:        sdkmath.LegacyMustNewDecFromStr("5.1"),
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {
					{
						Account:               sender1,
						ID:                    "order1",
						SellDenom:             denom1,
						BuyDenom:              denom2,
						SellQuantity:          sdkmath.NewInt(50),
						Price:                 sdkmath.LegacyMustNewDecFromStr("0.2"),
						RemainingSellQuantity: sdkmath.NewInt(50),
					},
					{
						Account:               sender2,
						ID:                    "order3",
						SellDenom:             denom1,
						BuyDenom:              denom2,
						SellQuantity:          sdkmath.NewInt(20),
						Price:                 sdkmath.LegacyMustNewDecFromStr("0.21"),
						RemainingSellQuantity: sdkmath.NewInt(20),
					},
				},
				denom2 + "/" + denom1: {
					{
						Account:               sender3,
						ID:                    "order4",
						SellDenom:             denom2,
						BuyDenom:              denom1,
						SellQuantity:          sdkmath.NewInt(30),
						Price:                 sdkmath.LegacyMustNewDecFromStr("5.1"),
						RemainingSellQuantity: sdkmath.NewInt(30),
					},
					{
						Account:               sender3,
						ID:                    "order2",
						SellDenom:             denom2,
						BuyDenom:              denom1,
						SellQuantity:          sdkmath.NewInt(10),
						Price:                 sdkmath.LegacyMustNewDecFromStr("5.2"),
						RemainingSellQuantity: sdkmath.NewInt(10),
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{},
		},
		{
			name: "fill_maker_and_partially_fill_next_taker",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				// filled fully by order1
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(5),
					Price:        sdkmath.LegacyMustNewDecFromStr("4"),
				},
				// order1 will be filled, and order3 remainder will be left
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(20),
					Price:        sdkmath.LegacyMustNewDecFromStr("5"),
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {},
				denom2 + "/" + denom1: {
					{
						Account:               sender3,
						ID:                    "order3",
						SellDenom:             denom2,
						BuyDenom:              denom1,
						SellQuantity:          sdkmath.NewInt(20),
						Price:                 sdkmath.LegacyMustNewDecFromStr("5"),
						RemainingSellQuantity: sdkmath.NewInt(5),
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 20)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 25)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denom1, 75)),
			},
		},
		{
			name: "match_last_taker_with_all_makers",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.15"),
				},
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.1"),
				},
				{
					Account:      sender4,
					ID:           "order4",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1000),
					Price:        sdkmath.LegacyMustNewDecFromStr("5"),
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {},
				denom2 + "/" + denom1: {
					{
						Account:               sender4,
						ID:                    "order4",
						SellDenom:             denom2,
						BuyDenom:              denom1,
						SellQuantity:          sdkmath.NewInt(1000),
						Price:                 sdkmath.LegacyMustNewDecFromStr("5"),
						RemainingSellQuantity: sdkmath.NewInt(955),
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 20)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom2, 15)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denom2, 10)),
				sender4: sdk.NewCoins(sdk.NewInt64Coin(denom1, 300)),
			},
		},
		{
			name: "fill_with_equal_amount",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(20),
					Price:        sdkmath.LegacyMustNewDecFromStr("5"),
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {},
				denom2 + "/" + denom1: {},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 20)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
			},
		},
		{
			name: "order_rounding_issuer_smaller_order_filled_with_lower_than_expected_amount",

			// with the `filling a small order (lower volume) with a lower price` strategy sender2 receives 26 with price of 2.63157894737
			// so the effective price is 10 / 26 = 0.38461538461
			// sender1 price diff `0.38461538461 - 0.375 = 0.00961538461`
			// sender2 price diff `0.38461538461 - 0.38 = 0.00461538461` (1/2.63157894737 = 0.38)
			// with the update strategy `filling a bigger (greater volume) order with a lower price` :
			// so the effective price is 10 / 27 = 0.37037037037
			// sender1 price diff `0.37037037037 - 0.375 = -0.00462962963`
			// sender2 price diff `0.37037037037 - 0.38 = -0.00962962963`
			// it will affect the price of smaller order less than price of bigger
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denom1,
					BuyDenom:  denom2,
					// you can update that value to 10 as a result order will become smaller and take lower price
					SellQuantity: sdkmath.NewInt(1000000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.375"), // min 375000
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(10),
					Price:        sdkmath.LegacyMustNewDecFromStr("2.63157894737"), // min 26.3
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {
					{
						Account:               sender1,
						ID:                    "order1",
						SellDenom:             denom1,
						BuyDenom:              denom2,
						SellQuantity:          sdkmath.NewInt(1000000),
						Price:                 sdkmath.LegacyMustNewDecFromStr("0.375"), // min 375000
						RemainingSellQuantity: sdkmath.NewInt(999973),                   // 999973 * 0.375 + 10(from balance) = 374999.875
					},
				},
				denom2 + "/" + denom1: {},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 10)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 27)), // the taker receives more
			},
		},
		{
			name: "invalid_amount_maker_and_taker",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(2),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.4"), // min 0.8 <- unachievable
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(5),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.13"), // min 0.65 <- unachievable
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom1, 2)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom2, 5)),
			},
		},
		{
			name: "cancel_remaining_maker_order",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(3),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.5"), // min 1.5
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1),
					Price:        sdkmath.LegacyMustNewDecFromStr("2"),
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {},
				denom2 + "/" + denom1: {},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom1, 1), sdk.NewInt64Coin(denom2, 1)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 2)),
			},
		},
		{
			name: "cancel_remaining_taker_order",
			newOrders: []keeper.Order{
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1),
					Price:        sdkmath.LegacyMustNewDecFromStr("2"),
				},
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(3),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.5"), // min 1.5
				},
			},
			expectedOrderBooks: map[string][]keeper.Order{
				denom1 + "/" + denom2: {},
				denom2 + "/" + denom1: {},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom1, 1), sdk.NewInt64Coin(denom2, 1)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 2)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := keeper.NewApp()

			passedOrdersSum := sdk.NewCoins()
			for _, order := range tc.newOrders {
				app.PlaceOrder(order)
				// after any order execution the total balance of the market must remain the same
				passedOrdersSum = passedOrdersSum.Add(sdk.NewCoin(order.SellDenom, order.SellQuantity))

				marketSum := sdk.NewCoins()
				for obKey := range app.OrderBooks {
					for _, obOrder := range app.OrderBooks[obKey] {
						marketSum = marketSum.Add(sdk.NewCoin(obOrder.SellDenom, obOrder.RemainingSellQuantity))
					}
				}
				for account := range app.Balances {
					marketSum = marketSum.Add(app.Balances[account]...)
				}
				require.Equal(t, passedOrdersSum.String(), marketSum.String())
			}
			require.EqualValues(t, tc.expectedOrderBooks, app.OrderBooks)
			require.EqualValues(t, tc.expectedBalances, app.Balances)
		})
	}
}
