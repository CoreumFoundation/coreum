package keeper_test

import (
	"fmt"
	"math/big"
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

		denomCORE = "ucore"   // 1CORE = 10^6ucore
		denomUSDC = "uusdc"   // 1USDC = 10^6uusdc
		denomBTC  = "satoshi" // 1BTC  = 10^8satoshi
	)

	denomTicks := map[string]uint{
		denomCORE: 4, // tradable step for CORE is 10^4ucore = 10,000ucore = 0.01CORE // Value from bitrue.
		denomUSDC: 4, // tradable step for USDC is 10^4uusdc = 10,000uusdc = 0.01USDC
		denomBTC:  3, // tradable step for BTC is 10^3satoshi = 1,000satoshi = 10^(-5)BTC = 0.00001BTC // Value from binance.
	}

	type testCase struct {
		name               string
		newOrders          []keeper.Order
		expectedOrderBooks map[string]*keeper.OrderBook
		expectedBalances   map[string]sdk.Coins
		expectErr          bool
	}
	testCases := []testCase{
		{
			name: "sell_quantity_doesn't_satisfy_tick",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(5_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
			},
			expectErr: true,
		},
		{
			name: "buy_quantity_doesn't_satisfy_tick",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(50_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.21"),
				},
			},
			expectErr: true,
		},
		{
			name: "no_match",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(500_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:      sender3,
					ID:           "order2",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(100_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("5.2"), // ~.1923
				},
				{
					Account:      sender2,
					ID:           "order3",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(2_000_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.21"),
				},
				{
					Account:      sender3,
					ID:           "order4",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(300_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("5.1"), // ~.0.196
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomCORE + "/" + denomUSDC: {
					SellDenom: denomCORE,
					BuyDenom:  denomUSDC,
					Records: []keeper.OrderBookRecord{
						{
							Account:               sender1,
							OrderID:               "order1",
							RemainingSellQuantity: sdkmath.NewInt(500_000),
							Price:                 sdkmath.LegacyMustNewDecFromStr("0.2"),
						},
						{
							Account:               sender2,
							OrderID:               "order3",
							Price:                 sdkmath.LegacyMustNewDecFromStr("0.21"),
							RemainingSellQuantity: sdkmath.NewInt(2_000_000),
						},
					},
				},
				denomUSDC + "/" + denomCORE: {
					SellDenom: denomUSDC,
					BuyDenom:  denomCORE,
					Records: []keeper.OrderBookRecord{
						{
							Account:               sender3,
							OrderID:               "order4",
							Price:                 sdkmath.LegacyMustNewDecFromStr("5.1"), // ~.0.196
							RemainingSellQuantity: sdkmath.NewInt(300_000),
						},
						{
							Account:               sender3,
							OrderID:               "order2",
							Price:                 sdkmath.LegacyMustNewDecFromStr("5.2"), // ~.1923
							RemainingSellQuantity: sdkmath.NewInt(100_000),
						},
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
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(1_000_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				// filled fully by order1
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(50_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("4"), // 0.25
				},
				// order1 will be filled, and order3 remainder will be left
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(200_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomCORE + "/" + denomUSDC: {
					SellDenom: denomCORE,
					BuyDenom:  denomUSDC,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
				denomUSDC + "/" + denomCORE: {
					SellDenom: denomUSDC,
					BuyDenom:  denomCORE,
					Records: []keeper.OrderBookRecord{
						{
							Account:               sender3,
							OrderID:               "order3",
							Price:                 sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
							RemainingSellQuantity: sdkmath.NewInt(50_000),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 200_000)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 250_000)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 750_000)),
			},
		},
		{
			name: "order_rounding_issue_initial_int_expected_amount_reduced_to_float",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(50_000_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.375"), // expect 1875
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(10_000_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("2.631"), // ~0.38 | expect 2631
				},
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(10_000_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("2.637"), // ~0.3792 | expected 2637
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomCORE + "/" + denomUSDC: {
					SellDenom: denomCORE,
					BuyDenom:  denomUSDC,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
				denomUSDC + "/" + denomCORE: {
					SellDenom: denomUSDC,
					BuyDenom:  denomCORE,
					Records: []keeper.OrderBookRecord{
						{
							Account: sender3,
							OrderID: "order3",
							Price:   sdkmath.LegacyMustNewDecFromStr("2.637"), // ~0.3792
							// 2334 + 125 * 2.637 = 2663.625 (was expected 2637)
							RemainingSellQuantity: sdkmath.NewInt(125),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 1875)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 2666)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 2334)),
			},
		},
		//{
		//	name: "match_last_taker_with_all_makers",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(100),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(100),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.15"),
		//		},
		//		{
		//			Account:      sender3,
		//			ID:           "order3",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(100),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.1"),
		//		},
		//		{
		//			Account:      sender4,
		//			ID:           "order4",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(1000),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records: []keeper.OrderBookRecord{
		//				{
		//					Account:               sender4,
		//					OrderID:               "order4",
		//					Price:                 sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
		//					RemainingSellQuantity: sdkmath.NewInt(955),
		//				},
		//			},
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 20)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 15)),
		//		sender3: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 10)),
		//		sender4: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 300)),
		//	},
		//},
		//{
		//	name: "fill_with_equal_amount",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(100),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.2"),
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(20),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 20)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 100)),
		//	},
		//},
		//{
		//	name: "order_rounding_issue_smaller_order_filled_with_lower_than_expected_amount",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:   sender1,
		//			ID:        "order1",
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			// you can update that value to 10 as a result order will become smaller and take lower price
		//			SellQuantity: sdkmath.NewInt(1000000),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.375"), // expect 375000
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(10),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("2.63157894737"), //  ~0.38 | expect 26.3
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records: []keeper.OrderBookRecord{
		//				{
		//					Account: sender1,
		//					OrderID: "order1",
		//					Price:   sdkmath.LegacyMustNewDecFromStr("0.375"), // expect 375000
		//					// 999974 * 0.375 + 10(from balance) = 375000.25
		//					RemainingSellQuantity: sdkmath.NewInt(999974),
		//				},
		//			},
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 10)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 26)),
		//	},
		//},
		//{
		//	name: "order_rounding_issue_denom_with_high_price_rounded_in_favor_or_higher_volume",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(3),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("10000"),
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(10_101),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.00009999"), // ~10001.0001
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records: []keeper.OrderBookRecord{
		//				{
		//					Account:               sender1,
		//					OrderID:               "order1",
		//					Price:                 sdkmath.LegacyMustNewDecFromStr("10000"),
		//					RemainingSellQuantity: sdkmath.NewInt(2),
		//				},
		//			},
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 10101)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 1)),
		//	},
		//},
		//{
		//	name: "invalid_amount_maker_and_taker",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(2),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.4"), // expected 0.8 <- unachievable
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(5),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.13"), // ~7.6923 | expected 0.65 <- unachievable
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 2)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 5)),
		//	},
		//},
		//{
		//	name: "cancel_remaining_maker_order",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(3),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.5"), // expected 1.5
		//		},
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(1),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("2"), //  0,5 | expected 2
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 1), sdk.NewInt64Coin(denomUSDC, 1)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 2)),
		//	},
		//},
		//{
		//	name: "cancel_remaining_taker_order",
		//	newOrders: []keeper.Order{
		//		{
		//			Account:      sender2,
		//			ID:           "order2",
		//			SellDenom:    denomUSDC,
		//			BuyDenom:     denomCORE,
		//			SellQuantity: sdkmath.NewInt(1),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("2"), //  0,5 | expected 2
		//		},
		//		{
		//			Account:      sender1,
		//			ID:           "order1",
		//			SellDenom:    denomCORE,
		//			BuyDenom:     denomUSDC,
		//			SellQuantity: sdkmath.NewInt(3),
		//			Price:        sdkmath.LegacyMustNewDecFromStr("0.5"), // min 1.5
		//		},
		//	},
		//	expectedOrderBooks: map[string]*keeper.OrderBook{
		//		denomCORE + "/" + denomUSDC: {
		//			SellDenom: denomCORE,
		//			BuyDenom:  denomUSDC,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//		denomUSDC + "/" + denomCORE: {
		//			SellDenom: denomUSDC,
		//			BuyDenom:  denomCORE,
		//			Records:   make([]keeper.OrderBookRecord, 0),
		//		},
		//	},
		//	expectedBalances: map[string]sdk.Coins{
		//		sender1: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 1), sdk.NewInt64Coin(denomUSDC, 1)),
		//		sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 2)),
		//	},
		//},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := keeper.NewApp(denomTicks)

			passedOrdersSum := sdk.NewCoins()
			for _, order := range tc.newOrders {
				err := app.PlaceOrder(order)
				if tc.expectErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				// after any order execution the total balance of the market must remain the same
				passedOrdersSum = passedOrdersSum.Add(sdk.NewCoin(order.SellDenom, order.SellQuantity))

				marketSum := sdk.NewCoins()
				for obKey := range app.OrderBooks {
					ob := app.OrderBooks[obKey]
					ob.Iterate(func(obOrder keeper.OrderBookRecord) bool {
						marketSum = marketSum.Add(sdk.NewCoin(ob.SellDenom, obOrder.RemainingSellQuantity))
						return false
					})
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

func TestCalculateSwapAmountExactV2(t *testing.T) {
	amntA, amntB := keeper.CalculateSwapAmountExactV2(
		big.NewRat(10_000_000_000, 375),
		big.NewRat(375, 10_000),
		big.NewInt(10_000),
		big.NewInt(10_000),
	)

	// 27200000
	// 26666666

	fmt.Printf("amntA: %s, amntB: %s\n", amntA.String(), amntB.String())
}
