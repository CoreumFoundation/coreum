package keeper

import (
	"fmt"
	"math/big"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"
)

// Order is matching sample ordrer.
type Order struct {
	ID        string
	Account   string
	SellDenom string
	BuyDenom  string
	// `quantity` here is the amount you want to sell.
	Quantity sdkmath.Int
	// `price` here is the amount you want to get for each token you sell.
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

// String returns string representation of the order.
func (o Order) String() string {
	return fmt.Sprintf(
		"ID:%s | account:%s | sellDenom:%s | buyDenom:%s | quantity:%s | ~revQuantity:%s | price:%s | ~revPrice:%s", //nolint:lll // string line.
		o.ID, o.Account, o.SellDenom, o.BuyDenom, o.Quantity.String(),
		o.Quantity.ToLegacyDec().Mul(o.Price).String(), o.Price.String(), sdkmath.LegacyOneDec().Quo(o.Price).String(),
	)
}

// OrderBookRecord is order record.
type OrderBookRecord struct {
	Account           string
	OrderID           string
	RemainingQuantity sdkmath.Int
	Price             sdkmath.LegacyDec
}

// String returns string representation of the order.
func (r *OrderBookRecord) String() string {
	return fmt.Sprintf(
		"OrderID:%s | account:%s | remainingQuantity:%s | ~remainingRevQuantity:%s | price:%s | ~revPrice:%s", //nolint:lll // string line.
		r.OrderID, r.Account, r.RemainingQuantity.String(),
		r.RemainingQuantity.ToLegacyDec().Mul(r.Price).String(), r.Price.String(),
		sdkmath.LegacyOneDec().Quo(r.Price).String(),
	)
}

// IsRevQuantityIsLessThanOne returns true is remaining rev quantity is less than one.
func (r *OrderBookRecord) IsRevQuantityIsLessThanOne() bool {
	// quantity * price < 1
	return BigRatLTOne(BigRatMul(NewBigRatFromSDKInt(r.RemainingQuantity), NewBigRatFromSDKDec(r.Price)))
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

// AddRecordFromOrder add order book record from the order.
func (ob *OrderBook) AddRecordFromOrder(order Order) bool {
	i := ob.findRecordIndex(order.Account, order.ID)
	if i >= 0 {
		panic(fmt.Sprintf("Record with the same account and orderID already exists in the order book, order:%s", order))
	}

	record := OrderBookRecord{
		Account:           order.Account,
		OrderID:           order.ID,
		RemainingQuantity: order.Quantity,
		Price:             order.Price,
	}
	// we don't allow to store record with the quantity * price < 1
	if record.IsRevQuantityIsLessThanOne() {
		fmt.Printf("The record won't be added to the order book, since remaining buy quantity is less than one, %s/%s, record:%s\n", //nolint:lll // breaking down this string will make it less readable
			ob.SellDenom, ob.BuyDenom, record.String(),
		)
		return false
	}

	fmt.Printf("Adding record to the order book, %s/%s, record:%s\n",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	ob.Records = append(ob.Records, record)
	ob.Sort()

	return true
}

// UpdateRecord updates order book record, if the record is not added the method returns false.
func (ob *OrderBook) UpdateRecord(record OrderBookRecord) bool {
	// we don't allow to store record with the quantity * price < 1
	if record.IsRevQuantityIsLessThanOne() {
		fmt.Printf("The record won't be updated, since remaining buy quantity is less than one, %s/%s, record:%s\n",
			ob.SellDenom, ob.BuyDenom, record.String(),
		)
		return false
	}

	fmt.Printf("Updating record in the order book, %s/%s, record:%s\n",
		ob.SellDenom, ob.BuyDenom, record.String(),
	)
	i := ob.findRecordIndex(record.Account, record.OrderID)
	if i < 0 {
		panic(fmt.Sprintf("Failed to find record to update in order book: %s", record))
	}
	ob.Records[i] = record

	return true
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

// GetRecordByAccountAndOrderID returns order book record by account and orderID.
func (ob *OrderBook) GetRecordByAccountAndOrderID(account, orderBookID string) (bool, OrderBookRecord) {
	i := ob.findRecordIndex(account, orderBookID)
	if i < 0 {
		return false, OrderBookRecord{}
	}

	return true, ob.Records[i]
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
	TickMultiplier         *big.Rat
	DenomSignificantAmount map[string]int64
	// sellDenom/buyDenom[]Order
	OrderBooks map[string]*OrderBook
	Balances   map[string]sdk.Coins
}

// NewApp returns new instance of an app.
func NewApp(tickMultiplier *big.Rat, denomAmountIncrement map[string]int64) *App {
	return &App{
		TickMultiplier:         tickMultiplier,
		DenomSignificantAmount: denomAmountIncrement,
		OrderBooks:             make(map[string]*OrderBook),
		Balances:               make(map[string]sdk.Coins),
	}
}

// PlaceOrder places and matches the order into the order book.
func (app *App) PlaceOrder(order Order) {
	fmt.Printf("\n---------- New Order ----------\n")
	fmt.Printf("Order: %s\n", order.String())

	tickSize := app.GetTickSize(order.SellDenom, order.BuyDenom)
	fmt.Printf("Tick size %s/%s: %s\n", order.SellDenom, order.BuyDenom, tickSize.FloatString(sdkmath.LegacyPrecision))
	app.ValidatePriceAgainstTickSize(order.SellDenom, order.BuyDenom, order.Price)

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

func (app *App) matchOrder(order Order, makerOB, ob *OrderBook) {
	if makerOB.IsEmpty() {
		app.addRecordToOBOrCancelOrder(ob, order)
		return
	}

	newOrderRevPriceRat := BigRatQuo(NewBigRatFromInt64(1), NewBigRatFromSDKDec(order.Price))
	makerOB.Iterate(func(makerOBRecord OrderBookRecord) bool {
		makerOBRecordPriceRat := NewBigRatFromSDKDec(makerOBRecord.Price)
		if BigRatLT(newOrderRevPriceRat, makerOBRecordPriceRat) {
			app.addRecordToOBOrCancelOrder(ob, order)
			return true
		}

		makerOBRecordExpectedReceiveQuantityRat := BigRatMul(
			NewBigRatFromSDKInt(makerOBRecord.RemainingQuantity), makerOBRecordPriceRat,
		)
		takerQuantityRat := NewBigRatFromSDKInt(order.Quantity)

		fmt.Printf(
			"\nMatch (%s/%s): newOrderRevPriceRat:%s >= makerOBRecordPriceRat:%s | makerOBRecordExpectedQuantity: %s | newOrderQuanriry:%s\n", //nolint:lll // breaking down this string will make it less readable
			order.ID, makerOBRecord.OrderID, newOrderRevPriceRat.FloatString(sdkmath.LegacyPrecision),
			makerOBRecordPriceRat.FloatString(sdkmath.LegacyPrecision),
			makerOBRecordExpectedReceiveQuantityRat.FloatString(sdkmath.LegacyPrecision),
			takerQuantityRat.FloatString(sdkmath.LegacyPrecision),
		)

		if BigRatGT(makerOBRecordExpectedReceiveQuantityRat, takerQuantityRat) {
			// the rev order remains, the taker is reduced fully
			makerOBRecordRevPriceRat := BigRatQuo(NewBigRatFromInt64(1), NewBigRatFromSDKDec(makerOBRecord.Price))
			maxExecutionQuantity, revMaxExecutionQuantity, remainder := FindMaxExecutionQuantity(
				order.Quantity.BigInt(), makerOBRecordRevPriceRat,
			)
			// taker receives
			app.SendCoin(order.Account, sdk.NewCoin(order.SellDenom, sdk.NewIntFromBigInt(remainder)))
			app.SendCoin(order.Account, sdk.NewCoin(order.BuyDenom, sdkmath.NewIntFromBigInt(revMaxExecutionQuantity)))
			// maker receives
			app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.BuyDenom, sdk.NewIntFromBigInt(maxExecutionQuantity)))
			// update state
			makerOBRecord.RemainingQuantity = makerOBRecord.RemainingQuantity.Sub(sdk.NewIntFromBigInt(revMaxExecutionQuantity))
			if !makerOB.UpdateRecord(makerOBRecord) {
				// is the `true` is returned the record wasn't update, and we cancel the remaining part
				app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.SellDenom, makerOBRecord.RemainingQuantity))
				makerOB.RemoveRecord(makerOBRecord)
				return true
			}

			return true
		}

		maxExecutionQuantity, revMaxExecutionQuantity, remainder := FindMaxExecutionQuantity(
			makerOBRecord.RemainingQuantity.BigInt(), makerOBRecordPriceRat,
		)
		// maker receives
		app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.BuyDenom, sdk.NewIntFromBigInt(revMaxExecutionQuantity)))
		app.SendCoin(makerOBRecord.Account, sdk.NewCoin(makerOB.SellDenom, sdk.NewIntFromBigInt(remainder)))
		// taker receives
		app.SendCoin(order.Account, sdk.NewCoin(order.BuyDenom, sdk.NewIntFromBigInt(maxExecutionQuantity)))
		// remove reduced record
		makerOB.RemoveRecord(makerOBRecord)
		// update state
		order.Quantity = order.Quantity.Sub(sdk.NewIntFromBigInt(revMaxExecutionQuantity))
		if BigIntEqZero(order.Quantity.BigInt()) {
			return true
		}
		// if nothing to match with add remaining order
		if makerOB.IsEmpty() {
			app.addRecordToOBOrCancelOrder(ob, order)
			return true
		}

		return false
	})
}

func (app *App) addRecordToOBOrCancelOrder(ob *OrderBook, order Order) {
	if !ob.AddRecordFromOrder(order) {
		app.SendCoin(order.Account, sdk.NewCoin(order.SellDenom, order.Quantity))
	}
}

// SendCoin sends coins to sample app accounts.
func (app *App) SendCoin(recipient string, amt sdk.Coin) {
	if amt.IsZero() {
		return
	}
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

// ValidatePriceAgainstTickSize validates the price against the tick size.
func (app *App) ValidatePriceAgainstTickSize(denom1, denom2 string, price sdkmath.LegacyDec) {
	tickSize := app.GetTickSize(denom1, denom2)
	priceRat := NewBigRatFromSDKDec(price)
	_, remainder := BigRatToIntWithRemainder(BigRatQuo(priceRat, tickSize))
	if !BigIntEqZero(remainder) {
		panic(
			fmt.Sprintf(
				"Invalid price: %s, doesn't match price tick: %s",
				price.String(), tickSize.FloatString(sdkmath.LegacyPrecision),
			),
		)
	}
}

// GetTickSize returns the tick size for the provided denoms.
func (app *App) GetTickSize(denom1, denom2 string) *big.Rat {
	amountIncrement1, ok := app.DenomSignificantAmount[denom1]
	if !ok {
		panic(fmt.Sprintf("denom %s not found in the registry", denom1))
	}

	amountIncrement2, ok := app.DenomSignificantAmount[denom2]
	if !ok {
		panic(fmt.Sprintf("denom %s not found in the registry", denom2))
	}

	// tick_size(denom1/denom2) = tick_multiplier * significant_amount(denom2) / significant_amount(denom1)
	return BigRatMul(
		app.TickMultiplier,
		BigRatQuo(NewBigRatFromInt64(amountIncrement2), NewBigRatFromInt64(amountIncrement1)),
	)
}

// FindMaxExecutionQuantity returns max execution quantity that gives int when we multiply quantity by price,
// max reversed execution quantity which is max execution quantity multiplied by price and remainder.
func FindMaxExecutionQuantity(quantity *big.Int, price *big.Rat) (*big.Int, *big.Int, *big.Int) {
	priceDenom := price.Denom()
	// truncate(quantity / priceDenom) * priceDenom
	maxExecutionQuantity := BigIntMul(BigIntQuo(quantity, priceDenom), priceDenom)
	// maxExecutionQuantity * price is always integer here
	revMaxExecutionQuantity := BigRatToInt(BigRatMul(NewBigRatFromBigInt(maxExecutionQuantity), price))
	remainder := BigIntSub(quantity, maxExecutionQuantity)

	return maxExecutionQuantity, revMaxExecutionQuantity, remainder
}

// ********** Math **********

var (
	// DecPrecisionReuse is  DecPrecisionReuse constant from sdkmath.Dec.
	DecPrecisionReuse = new(big.Int).Exp(big.NewInt(10), big.NewInt(sdkmath.LegacyPrecision), nil)
	// OneRat defines one in big.Rat.
	OneRat = (&big.Rat{}).SetFrac(big.NewInt(1), big.NewInt(1))
	// ZeroInt defins zer in big.Int.
	ZeroInt = big.NewInt(0)
)

// ********** Rat **********

// NewBigRatFromInt64 returns *big.Rat with int64 value.
func NewBigRatFromInt64(x int64) *big.Rat {
	return (&big.Rat{}).SetInt(big.NewInt(x))
}

// NewBigRatFromBigInt returns *big.Rat with *big.Int value.
func NewBigRatFromBigInt(x *big.Int) *big.Rat {
	return (&big.Rat{}).SetInt(x)
}

// NewBigRatFromSDKDec converts sdkmath.LegacyDec to *big.Rat.
func NewBigRatFromSDKDec(sdkDec sdkmath.LegacyDec) *big.Rat {
	return (&big.Rat{}).SetFrac(sdkDec.BigInt(), DecPrecisionReuse)
}

// NewBigRatFromSDKInt converts sdkmath.Int to *big.Rat.
func NewBigRatFromSDKInt(sdkInt sdkmath.Int) *big.Rat {
	return NewBigRatFromBigInt(sdkInt.BigInt())
}

// BigRatToIntWithRemainder converts *big.Rat to *big.Int integer part and reminder.
func BigRatToIntWithRemainder(rat *big.Rat) (*big.Int, *big.Int) {
	num := rat.Num()
	denom := rat.Denom()
	intPart := BigIntQuo(num, denom)
	return intPart, BigIntSub(num, BigIntMul(intPart, denom))
}

// BigRatToInt converts *big.Rat to *big.Int integer or panics if remainder is not zero.
func BigRatToInt(rat *big.Rat) *big.Int {
	intPart, remainder := BigRatToIntWithRemainder(rat)
	if !BigIntEqZero(remainder) {
		panic(fmt.Sprintf("faild to conver rat:%s to int without remainder", rat.FloatString(sdkmath.LegacyPrecision)))
	}
	return intPart
}

// BigRatQuo divides Rat with Rat.
func BigRatQuo(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Quo(x, y)
}

// BigRatMul multiplies Rat with Rat.
func BigRatMul(x, y *big.Rat) *big.Rat {
	return (&big.Rat{}).Mul(x, y)
}

// BigRatLTOne returns true if value lower than one.
func BigRatLTOne(x *big.Rat) bool {
	return x.Cmp(OneRat) == -1
}

// BigRatLT returns true if x is lower than y.
func BigRatLT(x, y *big.Rat) bool {
	return x.Cmp(y) == -1
}

// BigRatGT returns true if x is greater than y.
func BigRatGT(x, y *big.Rat) bool {
	return x.Cmp(y) == 1
}

// BigRatGTE returns true if x is greater or equal to y.
func BigRatGTE(x, y *big.Rat) bool {
	return x.Cmp(y) != -1
}

// ********** Int **********

// BigIntSub substitute Int with Int.
func BigIntSub(x, y *big.Int) *big.Int {
	return (&big.Int{}).Sub(x, y)
}

// BigIntQuo divides Int with Int.
func BigIntQuo(x, y *big.Int) *big.Int {
	return (&big.Int{}).Quo(x, y)
}

// BigIntMul multiplies Int with Int.
func BigIntMul(x, y *big.Int) *big.Int {
	return (&big.Int{}).Mul(x, y)
}

// BigIntEqZero returns true if value is equal to zero.
func BigIntEqZero(x *big.Int) bool {
	return x.Cmp(ZeroInt) == 0
}
