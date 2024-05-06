package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"math/big"
	"sort"
	"testing"
)

var (
	// constant from sdkmath.Dec
	decPrecisionReuse = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdkmath.LegacyPrecision), nil)
	oneRat            = (&big.Rat{}).SetFrac(big.NewInt(1), big.NewInt(1))
)

type Side string

var (
	SideSell Side = "Sell"
	SideBuy  Side = "Buy"
)

type Order struct {
	ID           string
	Account      string
	SellDenom    string
	BuyDenom     string
	SellQuantity sdkmath.Int
	BuyPrice     sdkmath.LegacyDec
}

func (o Order) String() string {
	return fmt.Sprintf(
		"ID:%s | account:%s | sellDenom:%s | buyDenom:%s | sellQuantity:%s | ~buyQuantity:%s | buyPrice:%s | ~sellPrice:%s",
		o.ID, o.Account, o.SellDenom, o.BuyDenom, o.SellQuantity.String(), o.SellQuantity.ToLegacyDec().Mul(o.BuyPrice).String(), o.BuyPrice.String(), sdkmath.LegacyOneDec().Quo(o.BuyPrice).String(),
	)
}

type OrderBookRecord struct {
	ID                    string
	Account               string // same as sender
	Price                 sdkmath.LegacyDec
	RemainingSellQuantity sdkmath.Int
}

func (r OrderBookRecord) String() string {
	return fmt.Sprintf(
		"ID:%s | account:%s | remainingSellQuantity:%s | ~remainingBuyQuantity:%s | buyPrice:%s | ~sellPrice:%s",
		r.ID, r.Account, r.RemainingSellQuantity.String(), r.RemainingSellQuantity.ToLegacyDec().Mul(r.Price).String(), r.Price.String(), sdkmath.LegacyOneDec().Quo(r.Price).String(),
	)
}

type App struct {
	// pair key + order book order
	OrderBooks map[string]map[Side][]OrderBookRecord
	Balances   map[string]sdk.Coins
}

func NewApp() *App {
	return &App{
		OrderBooks: make(map[string]map[Side][]OrderBookRecord),
		Balances:   make(map[string]sdk.Coins),
	}
}

func (app *App) PlaceOrder(order Order) {
	fmt.Printf("Adding order: %s\n", order.String())
	pairKey, side, denom0, denom1 := getPairKeyAndSide(order.BuyDenom, order.SellDenom)
	pairOrderBook, ok := app.OrderBooks[pairKey]
	if !ok {
		pairOrderBook = make(map[Side][]OrderBookRecord)
	}
	sidePairOrderBook, ok := pairOrderBook[side]
	if !ok {
		sidePairOrderBook = make([]OrderBookRecord, 0)
	}

	var oppositeSide Side
	switch side {
	case SideBuy:
		oppositeSide = SideSell
	case SideSell:
		oppositeSide = SideBuy
	}
	oppositeSidePairOrderBook, ok := pairOrderBook[oppositeSide]
	newRecord := OrderBookRecord{
		Price:                 order.BuyPrice,
		Account:               order.Account,
		RemainingSellQuantity: order.SellQuantity,
		ID:                    order.ID,
	}
	if !ok {
		sidePairOrderBook = append(sidePairOrderBook, newRecord)
	} else {

		for i, oppositeOrderRecord := range oppositeSidePairOrderBook {
			ratBuyPrice := (&big.Rat{}).SetFrac(newRecord.Price.BigInt(), decPrecisionReuse)
			ratOppositeSellPrice := (&big.Rat{}).SetFrac(decPrecisionReuse, oppositeOrderRecord.Price.BigInt())
			if ratBuyPrice.Cmp(ratOppositeSellPrice) == 1 {
				sidePairOrderBook = append(sidePairOrderBook, newRecord)
				break
			}

			buyAmount := (&big.Rat{}).SetInt(newRecord.RemainingSellQuantity.BigInt())
			oppositeSellAmount := (&big.Rat{}).Quo((&big.Rat{}).SetInt(oppositeOrderRecord.RemainingSellQuantity.BigInt()), ratOppositeSellPrice)

			// buy price >= sell price (1/oppositeOrderPrice)
			fmt.Printf("Match (%s/%s): buyPrice:%s | opposite sellPrice: %s | buyAmount:%s | oppositeSellAmount:%s\n",
				newRecord.ID, oppositeOrderRecord.ID,
				ratBuyPrice.FloatString(18), ratOppositeSellPrice.FloatString(18),
				buyAmount.FloatString(18), oppositeSellAmount.FloatString(18),
			)

			switch buyAmount.Cmp(oppositeSellAmount) {
			// new record will be fully filed with the sell price of opposite record
			case -1:
				reduceOppositeRatAmount := (&big.Rat{}).Mul(buyAmount, ratOppositeSellPrice)
				// reduceOppositeAmount < 1
				if reduceOppositeRatAmount.Cmp(oneRat) == -1 {
					returnAmount := sdk.NewCoin(denom1, newRecord.RemainingSellQuantity)
					fmt.Printf("Cancelling record %s, %s sent to %s  \n", newRecord.String(), returnAmount, newRecord.Account)
					app.sendCoin(newRecord.Account, returnAmount)
				} else {
					reduceOppositeAmount := ratAmountToInt(reduceOppositeRatAmount)
					oppositeOrderRecord.RemainingSellQuantity = oppositeOrderRecord.RemainingSellQuantity.Sub(reduceOppositeAmount)
					app.sendCoin(newRecord.Account, sdk.NewCoin(denom0, reduceOppositeAmount))
					app.sendCoin(oppositeOrderRecord.Account, sdk.NewCoin(denom1, newRecord.RemainingSellQuantity))
					oppositeSidePairOrderBook[i] = oppositeOrderRecord
				}
			case 0:
				// exchange coins one to one
				app.sendCoin(newRecord.Account, sdk.NewCoin(denom0, oppositeOrderRecord.RemainingSellQuantity))
				app.sendCoin(oppositeOrderRecord.Account, sdk.NewCoin(denom1, newRecord.RemainingSellQuantity))
				// remove record
				oppositeSidePairOrderBook = append(oppositeSidePairOrderBook[:i], oppositeSidePairOrderBook[i+1:]...)
			// opposite record will be fully filed with its price
			case 1:
				// oppositeSellAmount < 1
				if oppositeSellAmount.Cmp(oneRat) == -1 {
					returnAmount := sdk.NewCoin(denom0, oppositeOrderRecord.RemainingSellQuantity)
					fmt.Printf("Cancelling record %s, %s sent to %s  \n", oppositeOrderRecord.String(), returnAmount, oppositeOrderRecord.Account)
					app.sendCoin(oppositeOrderRecord.Account, returnAmount)
				} else {
					reduceAmount := ratAmountToInt(oppositeSellAmount)
					newRecord.RemainingSellQuantity = newRecord.RemainingSellQuantity.Sub(reduceAmount)
					app.sendCoin(newRecord.Account, sdk.NewCoin(denom0, oppositeOrderRecord.RemainingSellQuantity))
					app.sendCoin(oppositeOrderRecord.Account, sdk.NewCoin(denom1, reduceAmount))
				}
				// remove record
				oppositeSidePairOrderBook = append(oppositeSidePairOrderBook[:i], oppositeSidePairOrderBook[i+1:]...)
				// if nothing to match add order
				if len(oppositeSidePairOrderBook) == 0 {
					sidePairOrderBook = append(sidePairOrderBook, newRecord)
				}
			}
			pairOrderBook[oppositeSide] = oppositeSidePairOrderBook
		}
	}
	sort.Slice(sidePairOrderBook, func(i, j int) bool {
		return sidePairOrderBook[i].Price.LTE(sidePairOrderBook[j].Price)
	})

	pairOrderBook[side] = sidePairOrderBook
	app.OrderBooks[pairKey] = pairOrderBook
	app.printOrderBook(pairKey)
	app.printBalances()
	fmt.Println()
}

func (app *App) printOrderBook(pairKey string) {
	fmt.Printf("---------- Order book:%s ----------\n", pairKey)
	ob, ok := app.OrderBooks[pairKey]
	if !ok {
		fmt.Println("Empty...")
		return
	}
	sellOb, ok := ob[SideSell]
	if ok {
		fmt.Println("---------- Sell ----------")
		if len(sellOb) == 0 {
			fmt.Println("Empty...")
		}
		for _, o := range sellOb {
			fmt.Printf("Record: %s\n", o.String())
		}
	}
	buyOb, ok := ob[SideBuy]
	if ok {
		fmt.Println("---------- Buy ----------")
		if len(buyOb) == 0 {
			fmt.Println("Empty...")
		}
		for _, o := range buyOb {
			fmt.Printf("Record: %s\n", o.String())
		}
	}
}

func getPairKeyAndSide(buyDenom, sellDenom string) (string, Side, string, string) {
	var (
		denom0, denom1 string
		side           Side
	)
	if buyDenom < sellDenom {
		denom0 = buyDenom
		denom1 = sellDenom
		side = SideSell
	} else {
		denom0 = sellDenom
		denom1 = buyDenom
		side = SideBuy
	}

	return denom0 + denom1, side, denom0, denom1
}

func (app *App) sendCoin(recipient string, amt sdk.Coin) {
	accountBalances, ok := app.Balances[recipient]
	if !ok {
		accountBalances = make(sdk.Coins, 0)
	}
	app.Balances[recipient] = accountBalances.Add(amt)
}

func (app *App) printBalances() {
	fmt.Println("---------- Balances: ----------")
	addresses := lo.Keys(app.Balances)
	sort.Strings(addresses)
	for _, address := range addresses {
		fmt.Printf("Account %s: %s\n", address, app.Balances[address].String())
	}
}

func ratAmountToInt(amt *big.Rat) sdkmath.Int {
	return sdk.NewIntFromBigInt((&big.Int{}).Quo(amt.Num(), amt.Denom()))
}

func TestMatching(t *testing.T) {
	const (
		sender1 = "sender1"
		sender2 = "sender2"
		sender3 = "sender3"
		sender4 = "sender4"

		denom1 = "denom1"
		denom2 = "denom2"
		denom3 = "denom3"
	)
	type testCase struct {
		name               string
		newOrders          []Order
		expectedOrderBooks map[string]map[Side][]OrderBookRecord
		expectedBalances   map[string]sdk.Coins
	}
	testCases := []testCase{
		{
			name: "fill_new_and_fill_in_order_book",
			newOrders: []Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("0.2"),
				},
				// filled fully by order1
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(5),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("4"),
				},
				// order1 will be filled, and order3 remainder will be left
				{
					Account:      sender3,
					ID:           "order3",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(20),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("5"),
				},
			},
			expectedOrderBooks: map[string]map[Side][]OrderBookRecord{
				denom1 + denom2: {
					SideSell: {
						{
							ID:                    "order3",
							Account:               sender3,
							Price:                 sdkmath.LegacyMustNewDecFromStr("5"),
							RemainingSellQuantity: sdkmath.NewInt(5),
						},
					},
					SideBuy: {},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 20)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 25)),
				sender3: sdk.NewCoins(sdk.NewInt64Coin(denom1, 75)),
			},
		},
		{
			name: "not_precise_price_buy_order",
			newOrders: []Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(1),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("0.13"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("7.69"),
				},
			},
			expectedOrderBooks: map[string]map[Side][]OrderBookRecord{
				denom1 + denom2: {
					SideSell: {
						{
							ID:                    "order2",
							Account:               sender2,
							Price:                 sdkmath.LegacyMustNewDecFromStr("7.69"),
							RemainingSellQuantity: sdkmath.NewInt(1),
						},
					},
					SideBuy: {},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom1, 1)),
			},
		},
		{
			name: "not_precise_price_sell_order",
			newOrders: []Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(100),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("10"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("0.1"),
				},
			},
			expectedOrderBooks: map[string]map[Side][]OrderBookRecord{
				denom1 + denom2: {
					SideSell: {},
					SideBuy: {
						{
							ID:                    "order1",
							Account:               sender1,
							Price:                 sdkmath.LegacyMustNewDecFromStr("10"),
							RemainingSellQuantity: sdkmath.NewInt(100),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom2, 1)),
			},
		},
		{
			name: "partially_not_precise_price_buy_order",
			newOrders: []Order{
				{
					Account:      sender1,
					ID:           "order1",
					SellDenom:    denom1,
					BuyDenom:     denom2,
					SellQuantity: sdkmath.NewInt(11),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("0.13"),
				},
				{
					Account:      sender2,
					ID:           "order2",
					SellDenom:    denom2,
					BuyDenom:     denom1,
					SellQuantity: sdkmath.NewInt(1),
					BuyPrice:     sdkmath.LegacyMustNewDecFromStr("7.69"),
				},
			},
			expectedOrderBooks: map[string]map[Side][]OrderBookRecord{
				denom1 + denom2: {
					SideSell: {},
					SideBuy: {
						{
							ID:                    "order1",
							Account:               sender1,
							Price:                 sdkmath.LegacyMustNewDecFromStr("0.13"),
							RemainingSellQuantity: sdkmath.NewInt(4),
						},
					},
				},
			},
			expectedBalances: map[string]sdk.Coins{
				sender1: sdk.NewCoins(sdk.NewInt64Coin(denom2, 1)),
				sender2: sdk.NewCoins(sdk.NewInt64Coin(denom1, 7)),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := NewApp()

			initialOrdersSum := sdk.NewCoins()
			for _, order := range tc.newOrders {
				initialOrdersSum = initialOrdersSum.Add(sdk.NewCoin(order.SellDenom, order.SellQuantity))
				app.PlaceOrder(order)
			}
			require.EqualValues(t, tc.expectedOrderBooks, app.OrderBooks)
			require.EqualValues(t, tc.expectedBalances, app.Balances)
		})
	}
}
