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
	fmt.Printf("Adding record to the order book, %s/%s, record:%s",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	ob.Records = append(ob.Records, record)
	ob.Sort()
}

// UpdateRecord updates order book record.
func (ob *OrderBook) UpdateRecord(record OrderBookRecord) {
	fmt.Printf("Updating record in the order book, %s/%s, record:%s",
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
	fmt.Printf("Removing record from the order book, %s/%s, record:%s",
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
}

// NewApp returns new instance of an app.
func NewApp() *App {
	return &App{
		OrderBooks: make(map[string]*OrderBook),
		Balances:   make(map[string]sdk.Coins),
	}
}

// PlaceOrder places and matches the order into the order book.
func (app *App) PlaceOrder(takerOrder Order) {
	fmt.Printf("\nAdding new taker order: %s\n", takerOrder.String())

	// init remaining takerOrder quantity
	if takerOrder.IsBuyQuantityLessThanOne() {
		fmt.Printf("\nOrder cancelled, buy quantity < 1, %s\n", takerOrder.String())
		app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.SellDenom, takerOrder.SellQuantity))
		return
	}

	takerOKKey := takerOrder.OrderBookKey()
	makerOBKey := takerOrder.ReversedOrderBookKey()
	takerOB, ok := app.OrderBooks[takerOKKey]
	if !ok {
		takerOB = NewOrderBook(takerOrder.SellDenom, takerOrder.BuyDenom)
		app.OrderBooks[takerOKKey] = takerOB
	}
	makerOB, ok := app.OrderBooks[makerOBKey]
	if !ok {
		makerOB = NewOrderBook(takerOrder.BuyDenom, takerOrder.SellDenom)
		app.OrderBooks[makerOBKey] = makerOB
	}
	if makerOB.IsEmpty() {
		takerOB.AddOrder(takerOrder)
	} else {
		app.matchOrder(takerOrder, makerOB, takerOB)
	}

	app.PrintOrderBooks(takerOKKey, makerOBKey)
	app.PrintBalances()
}

func (app *App) matchOrder(takerOrder Order, makerOB, takerOB *OrderBook) {
	makerOB.Iterate(func(makerOBRecord OrderBookRecord) bool {
		takerBuyPrice := (&big.Rat{}).SetFrac(DecPrecisionReuse, takerOrder.Price.BigInt())
		makerSellPrice := (&big.Rat{}).SetFrac(makerOBRecord.Price.BigInt(), DecPrecisionReuse)

		if takerBuyPrice.Cmp(makerSellPrice) == -1 {
			takerOB.AddOrder(takerOrder)
			return true
		}

		// this amount uses the maker price since it's better or equal
		takerBuyAmount := (&big.Rat{}).Quo((&big.Rat{}).SetInt(takerOrder.SellQuantity.BigInt()), makerSellPrice)
		makerSellAmount := (&big.Rat{}).SetInt(makerOBRecord.RemainingSellQuantity.BigInt())

		fmt.Printf(
			"Match (%s/%s): takerBuyPrice:%s >= makerSellPrice:%s | takerBuyAmount: %s | makerSellAmount:%s \n",
			takerOrder.ID, makerOBRecord.OrderID, takerBuyPrice.FloatString(10), makerSellPrice.FloatString(10),
			takerBuyAmount.FloatString(10), makerSellAmount.FloatString(10),
		)

		switch takerBuyAmount.Cmp(makerSellAmount) {
		case -1: // the maker order remains, the taker is reduced fully
			// taker receives the sold by maker price tokens
			takerReceiveAmount := RatAmountToIntRoundDown(takerBuyAmount)
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))
			// maker receives the taker quantity
			makerReceiveAmount := takerOrder.SellQuantity
			app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.BuyDenom, makerReceiveAmount))
			// update state
			makerOBRecord.RemainingSellQuantity = makerOBRecord.RemainingSellQuantity.Sub(takerReceiveAmount)
			if makerOBRecord.IsRemainingBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration and remove
				app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.SellDenom, makerOBRecord.RemainingSellQuantity))
				makerOB.RemoveRecord(makerOBRecord)
			} else {
				makerOB.UpdateRecord(makerOBRecord)
			}
			return true
		case 0: // both orders are reduced
			app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.BuyDenom, takerOrder.SellQuantity))
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, makerOBRecord.RemainingSellQuantity))
			// remove reduced record
			makerOB.RemoveRecord(makerOBRecord)
			return true
		case 1: // the taker order remains and will go to the next loop, the maker is reduced fully
			// taker receives the amount maker sells
			takerReceiveAmount := sdk.NewIntFromBigInt(makerOBRecord.RemainingSellQuantity.BigInt())
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))
			// maker receive the amount
			makerReceiveAmount := RatAmountToIntRoundDown((&big.Rat{}).Mul(makerSellAmount, makerSellPrice))
			app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.BuyDenom, makerReceiveAmount))
			// update state
			takerOrder.SellQuantity = takerOrder.SellQuantity.Sub(makerReceiveAmount)
			// remove reduced record
			makerOB.RemoveRecord(makerOBRecord)
			if takerOrder.IsBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration
				app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.SellDenom, takerOrder.SellQuantity))
				return true
			}
			// if nothing to match with add remaining taker order
			if makerOB.IsEmpty() {
				takerOB.AddOrder(takerOrder)
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
func (app *App) PrintOrderBooks(key, revKey string) {
	obKeys := []string{
		key, revKey,
	}
	// sort to preserve the printed order for better readability
	sort.Strings(obKeys)
	for _, obKey := range obKeys {
		fmt.Printf("---------- Order book:%s ----------\n", obKey)
		ob, ok := app.OrderBooks[obKey]
		if !ok || ob.IsEmpty() {
			fmt.Println("Empty...")
			continue
		}
		i := 0
		ob.Iterate(func(record OrderBookRecord) bool {
			fmt.Printf("OrderBookRecord [%d]: %s\n", i, record.String())
			i++
			return false
		})
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
