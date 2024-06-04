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

		denomAAA = "AAA"
		denomBBB = "BBB"
	)

	defaultTickMultiplier := big.NewRat(1, 100) // 0.01

	type testCase struct {
		name                    string
		tickMultiplier          *big.Rat
		denomSignificantAmounts map[string]int64
		newOrders               []keeper.Order
		expectedOrderBooks      map[string]*keeper.OrderBook
		expectedBalances        map[string]sdk.Coins
	}
	testCases := []testCase{
		{
			name:           "maker_order_to_taker_and_taker_to_maker_with_remainders",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(50_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.371"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(10_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("2.6"),
				},
				{
					Account:   sender3,
					ID:        "order3",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(70_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("2.3"),
				},
				{
					Account:   sender4,
					ID:        "order4",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(220_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.36"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender4,
							OrderID:           "order4",
							Price:             sdkmath.LegacyMustNewDecFromStr("0.36"),
							RemainingQuantity: sdkmath.NewInt(78_665_161),
						},
					},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 18_550_000)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 26_954_000), sdk.NewInt64Coin(denomBBB, 66)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 164_380_839), sdk.NewInt64Coin(denomBBB, 4)),
				sender4: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 61_449_930)),
			},
		},
		{
			name:           "fill_with_equal_amount",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(100),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(20),
					Price:     sdkmath.LegacyMustNewDecFromStr("5"), // 0.2
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 20)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 100)),
			},
		},
		{
			name:           "no_match",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(50),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				{
					Account:   sender3,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(10),
					Price:     sdkmath.LegacyMustNewDecFromStr("5.2"), // ~.1923
				},
				{
					Account:   sender2,
					ID:        "order3",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(20),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.21"),
				},
				{
					Account:   sender3,
					ID:        "order4",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(30),
					Price:     sdkmath.LegacyMustNewDecFromStr("5.1"), // ~.0.196
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender1,
							OrderID:           "order1",
							RemainingQuantity: sdkmath.NewInt(50),
							Price:             sdkmath.LegacyMustNewDecFromStr("0.2"),
						},
						{
							Account:           sender2,
							OrderID:           "order3",
							Price:             sdkmath.LegacyMustNewDecFromStr("0.21"),
							RemainingQuantity: sdkmath.NewInt(20),
						},
					},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender3,
							OrderID:           "order4",
							Price:             sdkmath.LegacyMustNewDecFromStr("5.1"), // ~.0.196
							RemainingQuantity: sdkmath.NewInt(30),
						},
						{
							Account:           sender3,
							OrderID:           "order2",
							Price:             sdkmath.LegacyMustNewDecFromStr("5.2"), // ~.1923
							RemainingQuantity: sdkmath.NewInt(10),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{},
		},
		{
			name:           "initial_maker_order_with_buy_price_lower_than_one_and_canceled",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(1),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.999"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records:   []keeper.OrderBookRecord{},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 1)),
			},
		},
		{
			name:           "taker_order_with_buy_price_lower_than_one_but_executed",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(100),
					Price:     sdkmath.LegacyMustNewDecFromStr("1"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(1),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.9"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender1,
							OrderID:           "order1",
							RemainingQuantity: sdkmath.NewInt(99),
							Price:             sdkmath.LegacyMustNewDecFromStr("1"),
						},
					},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 1)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 1)),
			},
		},
		{
			name:           "remaining_maker_order_with_buy_price_lower_than_one_and_cancelled",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(10),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.25"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(2),
					Price:     sdkmath.LegacyMustNewDecFromStr("3"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 2), sdk.NewInt64Coin(denomBBB, 2)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 8)),
			},
		},
		{
			name:           "fill_multiple_maker_orders_by_one_taker",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 100,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(10),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.25"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(10),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.26"),
				},
				{
					Account:   sender3,
					ID:        "order3",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(20),
					Price:     sdkmath.LegacyMustNewDecFromStr("3"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records:   make([]keeper.OrderBookRecord, 0),
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender3,
							OrderID:           "order3",
							RemainingQuantity: sdkmath.NewInt(18),
							Price:             sdkmath.LegacyMustNewDecFromStr("3"),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 2), sdk.NewInt64Coin(denomBBB, 2)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 10)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 8)),
			},
		},
		{
			name:           "big_diff_of_ticks",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 1000000000,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(50_000_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.000000001"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(10_000_000),
					Price:     sdkmath.LegacyMustNewDecFromStr("210000000"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records:   []keeper.OrderBookRecord{},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender2,
							OrderID:           "order2",
							Price:             sdkmath.LegacyMustNewDecFromStr("210000000"),
							RemainingQuantity: sdkmath.NewInt(9999950),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 50)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomAAA, 50_000_000_000)),
			},
		},
		{
			name:           "taker_order_100%_remainder",
			tickMultiplier: defaultTickMultiplier,
			denomSignificantAmounts: map[string]int64{
				denomAAA: 10,
				denomBBB: 10,
			},
			newOrders: []keeper.Order{
				{
					Account:   sender1,
					ID:        "order1",
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Quantity:  sdkmath.NewInt(50),
					Price:     sdkmath.LegacyMustNewDecFromStr("0.37"),
				},
				{
					Account:   sender2,
					ID:        "order2",
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Quantity:  sdkmath.NewInt(10),
					Price:     sdkmath.LegacyMustNewDecFromStr("2.6"),
				},
			},
			expectedOrderBooks: map[string]*keeper.OrderBook{
				denomAAA + "/" + denomBBB: {
					SellDenom: denomAAA,
					BuyDenom:  denomBBB,
					Records: []keeper.OrderBookRecord{
						{
							Account:           sender1,
							OrderID:           "order1",
							Price:             sdkmath.LegacyMustNewDecFromStr("0.37"),
							RemainingQuantity: sdkmath.NewInt(50),
						},
					},
				},
				denomBBB + "/" + denomAAA: {
					SellDenom: denomBBB,
					BuyDenom:  denomAAA,
					Records:   []keeper.OrderBookRecord{},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denomBBB, 10)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := keeper.NewApp(tc.tickMultiplier, tc.denomSignificantAmounts)

			// copy balances
			balancesBefore := make(map[string]sdk.Coins, len(app.Balances))
			for k, v := range app.Balances {
				balancesBefore[k] = v
			}

			passedOrdersSum := sdk.NewCoins()
			for i, order := range tc.newOrders {
				// execute order
				require.NoError(t, app.PlaceOrder(order))
				// validate state
				passedOrdersSum = validateCoinsSum(t, app, passedOrdersSum, order)
				// validate for all passed orders
				fmt.Printf("----------  Orders execution statistic ----------\n")
				for j, passedOrder := range tc.newOrders {
					validatePriceViolation(t, app, balancesBefore, passedOrder)
					if i == j {
						break
					}
				}
			}

			require.EqualValues(t, tc.expectedOrderBooks, app.OrderBooks)
			require.EqualValues(t, tc.expectedBalances, app.Balances)
		})
	}
}

// validates that total sum of the coins in orders and balances are equal to sum of all coins in the orders.
func validateCoinsSum(t *testing.T, app *keeper.App, passedOrdersSum sdk.Coins, order keeper.Order) sdk.Coins {
	passedOrdersSum = passedOrdersSum.Add(sdk.NewCoin(order.SellDenom, order.Quantity))
	marketSum := sdk.NewCoins()
	for obKey := range app.OrderBooks {
		ob := app.OrderBooks[obKey]
		require.NoError(t, ob.Iterate(func(obOrder keeper.OrderBookRecord) (bool, error) {
			marketSum = marketSum.Add(sdk.NewCoin(ob.SellDenom, obOrder.RemainingQuantity))
			return false, nil
		}))
	}
	for account := range app.Balances {
		marketSum = marketSum.Add(app.Balances[account]...)
	}
	require.Equal(t, passedOrdersSum.String(), marketSum.String(), fmt.Sprintf("order:%s", order.String()))

	return passedOrdersSum
}

// validatePriceViolation validates that the limit order was executed with the provided price or better.
func validatePriceViolation(t *testing.T, app *keeper.App, balancesBefore map[string]sdk.Coins, order keeper.Order) {
	accountBalancesBefore, ok := balancesBefore[order.Account]
	if !ok {
		accountBalancesBefore = sdk.NewCoins()
	}

	accountBalancesAfter, ok := app.Balances[order.Account]
	if !ok {
		accountBalancesAfter = sdk.NewCoins()
	}
	accountBalancesDiff := accountBalancesAfter.Sub(accountBalancesBefore...)

	obKey := order.OrderBookKey()
	ob, ok := app.OrderBooks[obKey]
	require.True(t, ok)

	// sub remainder
	soldQuantity := keeper.BigIntSub(order.Quantity.BigInt(), accountBalancesDiff.AmountOf(order.SellDenom).BigInt())
	record, found := ob.GetRecordByAccountAndOrderID(order.Account, order.ID)
	if found {
		// sub not executed quantity
		soldQuantity = keeper.BigIntSub(soldQuantity, record.RemainingQuantity.BigInt())
	}
	receivedAmount := accountBalancesDiff.AmountOf(order.BuyDenom).BigInt()
	remainder := accountBalancesDiff.AmountOf(order.SellDenom).BigInt()

	if !keeper.BigIntEqZero(receivedAmount) {
		fmt.Printf("----------  Order: sender: %s, orderID: %s  ----------\n", order.Account, order.ID)
		fmt.Printf("Sold: %s%s\n", soldQuantity, order.SellDenom)
		fmt.Printf("Received: %s%s\n", receivedAmount, order.BuyDenom)
		fmt.Printf("Remainder: %s%s\n", remainder, order.SellDenom)
		// If order is not in the order book
		if !found {
			// (quantity - remainder) / quantity * 100
			fmt.Printf("Filled percent: %s\n",
				keeper.BigRatMul(
					keeper.BigRatQuo(
						keeper.NewBigRatFromBigInt(keeper.BigIntSub(order.Quantity.BigInt(), remainder)),
						keeper.NewBigRatFromSDKInt(order.Quantity),
					),
					keeper.NewBigRatFromInt64(100),
				).
					FloatString(sdkmath.LegacyPrecision),
			)
		}
		fmt.Printf("Expected price: %s\n", order.Price.String())
		executionPrice := keeper.BigRatQuo(
			keeper.NewBigRatFromBigInt(receivedAmount), keeper.NewBigRatFromBigInt(soldQuantity),
		)
		fmt.Printf("Execution price: %s\n", executionPrice.FloatString(sdkmath.LegacyPrecision))
		require.True(t, keeper.BigRatGTE(executionPrice, keeper.NewBigRatFromSDKDec(order.Price)))
	}
}
