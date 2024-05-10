package keeper

import (
	"fmt"
	"math/big"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
)

var (
	// DecPrecisionReuse is  DecPrecisionReuse constant from sdkmath.Dec.
	DecPrecisionReuse = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdkmath.LegacyPrecision), nil)
	// OneRat defins one in big.Rat.
	OneRat = (&big.Rat{}).SetFrac(big.NewInt(1), big.NewInt(1))
)

// Order is matching sample ordrer.
type Order struct {
	ID        string
	Account   string
	SellDenom string
	BuyDenom  string
	// sellQuantity = buyQuantity / price | buyQuantity = sellQuantity * price
	SellQuantity sdkmath.Int
	// price = buyQuantity / sellQuantity
	// TODO(dzmitryhil) update the Price to custom type.
	Price sdkmath.LegacyDec
}

// OrderBookKey returns SellDenom/BuyDenom order book key.
func (o Order) OrderBookKey() string {
	return fmt.Sprintf("%s/%s", o.SellDenom, o.BuyDenom)
}

// ReversedOrderBookKey returns BuyDenom/SellDenom order book key.
func (o Order) ReversedOrderBookKey() string {
	return fmt.Sprintf("%s/%s", o.BuyDenom, o.SellDenom)
}

// IsBuyQuantityLessThanOne returns true if buy quantity is less than one.
func (o Order) IsBuyQuantityLessThanOne() bool {
	return IsAmountLessThanOne(o.Price, o.SellQuantity)
}

// String returns string representation of the order.
func (o Order) String() string {
	return fmt.Sprintf(
		"ID:%s | account:%s | sellDenom:%s | buyDenom:%s | sellQuantity:%s | ~buyQuantity:%s | buyPrice:%s | ~sellPrice:%s", //nolint:lll // string line.
		o.ID, o.Account, o.SellDenom, o.BuyDenom, o.SellQuantity.String(),
		o.SellQuantity.ToLegacyDec().Mul(o.Price).String(), o.Price.String(), sdkmath.LegacyOneDec().Quo(o.Price).String(),
	)
}

// OrderBookRecord is order record.
type OrderBookRecord struct {
	Account string
	OrderID string
	// the remaining sell quantity to fill the order
	RemainingSellQuantity sdkmath.Int
	// TODO(dzmitryhil) update the Price to custom type.
	Price sdkmath.LegacyDec
}

// IsRemainingBuyQuantityLessThanOne returns true if remaining buy quantity is less than one.
func (r *OrderBookRecord) IsRemainingBuyQuantityLessThanOne() bool {
	return IsAmountLessThanOne(r.Price, r.RemainingSellQuantity)
}

// String returns string representation of the order.
func (r *OrderBookRecord) String() string {
	return fmt.Sprintf(
		"OrderID:%s | account:%s | remainingSellQuantity:%s | ~remainingbuyQuantity:%s | buyPrice:%s | ~sellPrice:%s", //nolint:lll // string line.
		r.OrderID, r.Account, r.RemainingSellQuantity.String(),
		r.RemainingSellQuantity.ToLegacyDec().Mul(r.Price).String(), r.Price.String(),
		sdkmath.LegacyOneDec().Quo(r.Price).String(),
	)
}

// OrderBook is order book.
type OrderBook struct {
	SellDenom string
	BuyDenom  string
	Records   []OrderBookRecord
}

// NewOrderBook returns new instance of the OrderBook.
func NewOrderBook(sellDenom, buyDenom string) *OrderBook {
	return &OrderBook{
		SellDenom: sellDenom,
		BuyDenom:  buyDenom,
		Records:   make([]OrderBookRecord, 0),
	}
}

// AddOrder add order book record from the order.
func (ob *OrderBook) AddOrder(order Order) {
	i := ob.findRecordIndex(order.Account, order.ID)
	if i >= 0 {
		panic(fmt.Sprintf("Record with the same account and orderID is already exists in the order book, order:%s", order))
	}
	record := OrderBookRecord{
		Account:               order.Account,
		OrderID:               order.ID,
		RemainingSellQuantity: order.SellQuantity,
		Price:                 order.Price,
	}
	fmt.Printf("Adding record to the order book, %s/%s, record:%s\n",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	ob.Records = append(ob.Records, record)
	ob.Sort()
}

// UpdateRecord updates order book record.
func (ob *OrderBook) UpdateRecord(record OrderBookRecord) {
	fmt.Printf("Updating record in the order book, %s/%s, record:%s\n",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	i := ob.findRecordIndex(record.Account, record.OrderID)
	if i < 0 {
		panic(fmt.Sprintf("Failed to find record to update in order book: %s", record))
	}
	ob.Records[i] = record
}

// RemoveRecord updates order book records.
func (ob *OrderBook) RemoveRecord(record OrderBookRecord) {
	fmt.Printf("Removing record from the order book, %s/%s, record:%s\n",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	i := ob.findRecordIndex(record.Account, record.OrderID)
	if i < 0 {
		panic(fmt.Sprintf("Failed to find record to remove in order book: %s", record))
	}
	ob.Records = append(ob.Records[:i], ob.Records[i+1:]...)
}

// IsEmpty returns true if empty.
func (ob *OrderBook) IsEmpty() bool {
	return len(ob.Records) == 0
}

// Iterate iterates over order book records.
func (ob *OrderBook) Iterate(iterator func(record OrderBookRecord) bool) {
	// use copy to allow removal at the time of the iterating
	recordCopy := lo.Map(ob.Records, func(record OrderBookRecord, _ int) OrderBookRecord {
		return record
	})
	for _, record := range recordCopy {
		if iterator(record) {
			break
		}
	}
}

// Print prints the order book.
func (ob *OrderBook) Print() {
	fmt.Printf("---------- Order book:%s/%s ----------\n", ob.SellDenom, ob.BuyDenom)
	if ob.IsEmpty() {
		fmt.Println("Empty...")
		return
	}
	i := 0
	ob.Iterate(func(record OrderBookRecord) bool {
		fmt.Printf("OrderBookRecord [%d]: %s\n", i, record.String())
		i++
		return false
	})
}

func (ob *OrderBook) findRecordIndex(account, orderID string) int {
	for i, record := range ob.Records {
		if record.Account == account && record.OrderID == orderID {
			return i
		}
	}
	return -1
}

// Sort sorts order book records by price desc.
func (ob *OrderBook) Sort() {
	sort.Slice(ob.Records, func(i, j int) bool {
		return ob.Records[i].Price.LTE(ob.Records[j].Price)
	})
}

// App is sample matching app.
type App struct {
	// sellDenom/buyDenom[]Order
	OrderBooks map[string]*OrderBook
	Balances   map[string]sdk.Coins

	DenomTicks map[string]uint
}

// NewApp returns new instance of an app.
func NewApp(denomTicks map[string]uint) *App {
	return &App{
		OrderBooks: make(map[string]*OrderBook),
		Balances:   make(map[string]sdk.Coins),
		DenomTicks: denomTicks,
	}
}

// PlaceOrder places and matches the order into the order book.
func (app *App) PlaceOrder(order Order) error {
	fmt.Printf("\nAdding new order: %s\n", order.String())

	sellTick, ok := app.DenomTicks[order.SellDenom]
	if !ok {
		return fmt.Errorf("unspecified tick for sell denom: %s", order.SellDenom)
	}
	buyTick, ok := app.DenomTicks[order.BuyDenom]
	if !ok {
		return fmt.Errorf("unspecified tick for buy denom: %s", order.BuyDenom)
	}

	sellTickPow10 := sdkmath.NewInt(int64(10 ^ sellTick))
	if !order.SellQuantity.Mod(sellTickPow10).Equal(sdkmath.ZeroInt()) {
		return fmt.Errorf("invalid sell quantity: %s, tick not satisfied: %d", order.SellQuantity, sellTick)
	}

	buyTickPow10 := sdkmath.NewInt(int64(10 ^ buyTick))
	buyQuantityDec := order.SellQuantity.ToLegacyDec().Mul(order.Price)
	if !buyQuantityDec.IsInteger() {
		return fmt.Errorf("invalid buy quantity: %s, not integer", buyQuantityDec.String())
	}
	buyQuantity := sdkmath.NewIntFromBigInt(buyQuantityDec.BigInt())
	if !buyQuantity.Mod(buyTickPow10).Equal(sdkmath.ZeroInt()) {
		return fmt.Errorf("invalid buy quantity: %s, tick not satisfied: %d", buyQuantity, buyTick)
	}

	// init remaining order quantity
	if order.IsBuyQuantityLessThanOne() {
		fmt.Printf("\nOrder cancelled, buy quantity < 1, %s\n", order.String())
		app.SendCoin(order.Account, sdk.NewCoin(order.SellDenom, order.SellQuantity))
		return nil // todo
	}

	obKey := order.OrderBookKey()
	revOBKey := order.ReversedOrderBookKey()
	ob, ok := app.OrderBooks[obKey]
	if !ok {
		ob = NewOrderBook(order.SellDenom, order.BuyDenom)
		app.OrderBooks[obKey] = ob
	}
	revOB, ok := app.OrderBooks[revOBKey]
	if !ok {
		revOB = NewOrderBook(order.BuyDenom, order.SellDenom)
		app.OrderBooks[revOBKey] = revOB
	}

	app.matchOrder(order, revOB, ob)
	app.PrintOrderBooks(obKey, revOBKey)
	app.PrintBalances()
}

func (app *App) matchOrder(takerOrder Order, ob, revOB *OrderBook) {
	if ob.IsEmpty() {
		revOB.AddOrder(takerOrder)
	}

	ob.Iterate(func(revOBRecord OrderBookRecord) bool {
		takerPriceReversed := (&big.Rat{}).SetFrac(DecPrecisionReuse, takerOrder.Price.BigInt())
		makerPrice := (&big.Rat{}).SetFrac(revOBRecord.Price.BigInt(), DecPrecisionReuse)

		// If takerPriceReversed is less than makerPrice, then orders don't match.
		if takerPriceReversed.Cmp(makerPrice) == -1 {
			revOB.AddOrder(takerOrder)
			return true
		}

		// To calculate takerBuyAmount, we use makerPrice because it is better or equal to takerPriceReversed.
		// takerBuyAmount = takerSellQuantity / makerPrice
		takerBuyAmount := (&big.Rat{}).Quo((&big.Rat{}).SetInt(takerOrder.SellQuantity.BigInt()), makerPrice)
		makerSellAmount := (&big.Rat{}).SetInt(revOBRecord.RemainingSellQuantity.BigInt())

		fmt.Printf(
			"Match (%s/%s): takerPriceReversed:%s >= makerPrice:%s | takerBuyAmount: %s | makerSellAmount:%s\n",
			takerOrder.ID, revOBRecord.OrderID, takerPriceReversed.FloatString(10), makerPrice.FloatString(10),
			takerBuyAmount.FloatString(10), makerSellAmount.FloatString(10),
		)

		switch takerBuyAmount.Cmp(makerSellAmount) {
		case -1: // takerBuyAmount < makerSellAmount: taker order is matched fully, and maker order is matched partially.
			takerReceiveAmount := RatAmountToIntRoundDown(takerBuyAmount)
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))
			// maker receives the taker sell quantity
			makerReceiveAmount := takerOrder.SellQuantity
			app.SendCoin(revOBRecord.Account, sdk.NewCoin(ob.BuyDenom, makerReceiveAmount))
			// update state
			revOBRecord.RemainingSellQuantity = revOBRecord.RemainingSellQuantity.Sub(takerReceiveAmount)
			if revOBRecord.IsRemainingBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration and remove
				app.SendCoin(revOBRecord.Account, sdk.NewCoin(ob.SellDenom, revOBRecord.RemainingSellQuantity))
				ob.RemoveRecord(revOBRecord)
			} else {
				ob.UpdateRecord(revOBRecord)
			}
			return true
		case 0: // takerBuyAmount == makerSellAmount: both taker and maker orders are matched fully.
			app.SendCoin(revOBRecord.Account, sdk.NewCoin(ob.BuyDenom, takerOrder.SellQuantity))
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, revOBRecord.RemainingSellQuantity))
			// remove reduced record
			ob.RemoveRecord(revOBRecord)
			return true
		case 1: // takerBuyAmount > makerSellAmount: taker order is matched partially, and maker order is matched fully.
			// taker receives the amount maker sells
			takerReceiveAmount := sdk.NewIntFromBigInt(revOBRecord.RemainingSellQuantity.BigInt())
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))

			makerReceiveAmount := RatAmountToIntRoundDown((&big.Rat{}).Mul(makerSellAmount, makerPrice))
			app.SendCoin(revOBRecord.Account, sdk.NewCoin(ob.BuyDenom, makerReceiveAmount))
			// update state
			takerOrder.SellQuantity = takerOrder.SellQuantity.Sub(makerReceiveAmount)
			// remove reduced record
			ob.RemoveRecord(revOBRecord)
			if takerOrder.IsBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration
				app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.SellDenom, takerOrder.SellQuantity))
				return true
			}
			// if nothing to match with add remaining takerOrder
			if ob.IsEmpty() {
				revOB.AddOrder(takerOrder)
				return true
			}
		}

		return false
	})
}

// SendCoin sends coins to sample app accounts.
func (app *App) SendCoin(recipient string, amt sdk.Coin) {
	accountBalances, ok := app.Balances[recipient]
	if !ok {
		accountBalances = make(sdk.Coins, 0)
	}
	app.Balances[recipient] = accountBalances.Add(amt)
	fmt.Printf("Sending coins, recipient: %s, amount:%s\n", recipient, amt.String())
}

// PrintOrderBooks prints order books by keys.
func (app *App) PrintOrderBooks(obKey, revKey string) {
	fmt.Println("---------- Order books: ----------")
	obKeys := []string{
		obKey, revKey,
	}
	// sort to preserve the printed order for better readability
	sort.Strings(obKeys)
	for _, obKey := range obKeys {
		app.OrderBooks[obKey].Print()
	}
}

// PrintBalances prints sample app current balances.
func (app *App) PrintBalances() {
	fmt.Println("---------- Balances: ----------")
	if len(app.Balances) == 0 {
		fmt.Println("Empty...")
		return
	}
	addresses := lo.Keys(app.Balances)
	sort.Strings(addresses)
	for _, address := range addresses {
		fmt.Printf("Account %s: %s\n", address, app.Balances[address].String())
	}
}

// RatAmountToIntRoundDown converts the big.Rat to sdkmath.Int with round down strategy.
func RatAmountToIntRoundDown(amt *big.Rat) sdkmath.Int {
	return sdk.NewIntFromBigInt((&big.Int{}).Quo(amt.Num(), amt.Denom()))
}

// IsAmountLessThanOne returns true is price * quantity < 1.
func IsAmountLessThanOne(price sdkmath.LegacyDec, quantity sdkmath.Int) bool {
	ratPrice := (&big.Rat{}).SetFrac(price.BigInt(), DecPrecisionReuse)
	// price * remainingSellQuantity  < 1
	return (&big.Rat{}).Mul((&big.Rat{}).SetInt(quantity.BigInt()), ratPrice).Cmp(OneRat) == -1
}
