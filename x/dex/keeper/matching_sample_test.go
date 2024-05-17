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

		denomCORE = "ucore"   // 1CORE = 10^6ucore
		denomUSDC = "uusdc"   // 1USDC = 10^6uusdc
		denomBTC  = "satoshi" // 1BTC  = 10^8satoshi
	)

	minAmntIncrements := map[string]int64{
		denomCORE: 10_000, // tradable step for CORE is 10,000ucore = 0.01CORE // Value from bitrue.
		denomUSDC: 10_000, // tradable step for USDC is 10,000uusdc = 0.01USDC
		denomBTC:  1_000,  // tradable step for BTC is 1,000satoshi = 10^(-5)BTC = 0.00001BTC // Value from binance.
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
			name: "buy_quantity_less_than_tick",
			newOrders: []keeper.Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denomCORE,
					BuyDenom:     denomUSDC,
					SellQuantity: sdkmath.NewInt(10_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.01"),
				},
			},
			expectErr:        true,
			expectedBalances: map[string]sdk.Coins{},
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
			expectedBalances: map[string]sdk.Coins{},
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
			expectedOrderBooks: nil,
			expectedBalances:   map[string]sdk.Coins{},
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
			expectedOrderBooks: nil,
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
					SellQuantity: sdkmath.NewInt(500_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("3.75"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(100_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.26"),
				},
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denomUSDC,
					BuyDenom:     denomCORE,
					SellQuantity: sdkmath.NewInt(100_000),
					Price:        sdkmath.LegacyMustNewDecFromStr("0.26"),
				},
			},
			expectedOrderBooks: nil,
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomUSDC, 150000)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 20000)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denomCORE, 20000)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := keeper.NewApp(minAmntIncrements)

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

				//marketSum := sdk.NewCoins()
				//for obKey := range app.OrderBooks {
				//	ob := app.OrderBooks[obKey]
				//	ob.Iterate(func(obOrder keeper.OrderBookRecord) bool {
				//		marketSum = marketSum.Add(sdk.NewCoin(ob.SellDenom, obOrder.RemainingSellQuantity))
				//		return false
				//	})
				//}
				//for account := range app.Balances {
				//	marketSum = marketSum.Add(app.Balances[account]...)
				//}
				//require.Equal(t, passedOrdersSum.String(), marketSum.String())
			}
			//require.EqualValues(t, tc.expectedOrderBooks, app.OrderBooks)
			require.EqualValues(t, tc.expectedBalances, app.Balances)
		})
	}
}
