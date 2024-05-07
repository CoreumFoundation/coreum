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
	// the remaining sell quantity to fill the order
	RemainingSellQuantity sdkmath.Int
	// price = buyQuantity / sellQuantity
	Price sdkmath.LegacyDec
}

// TakerOrderBookKey returns SellDenom/BuyDenom order book key.
func (o Order) TakerOrderBookKey() string {
	return fmt.Sprintf("%s/%s", o.SellDenom, o.BuyDenom)
}

// MakerOrderBookKey returns BuyDenom/SellDenom order book key.
func (o Order) MakerOrderBookKey() string {
	return fmt.Sprintf("%s/%s", o.BuyDenom, o.SellDenom)
}

// IsRemainingBuyQuantityLessThanOne returns true is expected remaining buy quantity is less than zero, so order can't
// be filled correctly.
func (o Order) IsRemainingBuyQuantityLessThanOne() bool {
	ratPrice := (&big.Rat{}).SetFrac(o.Price.BigInt(), DecPrecisionReuse)
	return (&big.Rat{}).Mul((&big.Rat{}).SetInt(o.RemainingSellQuantity.BigInt()), ratPrice).Cmp(OneRat) == -1
}

// String returns string representation of the order.
func (o Order) String() string {
	remainingSellQuantity := o.RemainingSellQuantity
	if remainingSellQuantity.IsNil() {
		remainingSellQuantity = sdk.ZeroInt()
	}

	return fmt.Sprintf(
		"ID:%s | account:%s | sellDenom:%s | buyDenom:%s | sellQuantity:%s | ~buyQuantity:%s | buyPrice:%s | ~sellPrice:%s | remainingSellQuantity:%s | ~remainingBuyQuantity:%s", //nolint:lll // string line.
		o.ID, o.Account, o.SellDenom, o.BuyDenom, o.SellQuantity.String(),
		o.SellQuantity.ToLegacyDec().Mul(o.Price).String(), o.Price.String(), sdkmath.LegacyOneDec().Quo(o.Price).String(),
		remainingSellQuantity.String(), remainingSellQuantity.ToLegacyDec().Mul(o.Price).String(),
	)
}

// ********** App **********

// App is sample matching app.
type App struct {
	// sellDenom/buyDenom[]Order
	OrderBooks map[string][]Order
	Balances   map[string]sdk.Coins
}

// NewApp returns new instance of an app.
func NewApp() *App {
	return &App{
		OrderBooks: make(map[string][]Order),
		Balances:   make(map[string]sdk.Coins),
	}
}

// PlaceOrder places and matches the order into the order book.
func (app *App) PlaceOrder(takerOrder Order) {
	fmt.Printf("\nAdding new taker order: %s\n", takerOrder.String())
	takerOKKey := takerOrder.TakerOrderBookKey()
	makerOBKey := takerOrder.MakerOrderBookKey()
	takerOB, ok := app.OrderBooks[takerOKKey]
	if !ok {
		takerOB = make([]Order, 0)
	}
	makerOB, ok := app.OrderBooks[makerOBKey]
	if !ok {
		makerOB = make([]Order, 0)
	}
	// init remaining takerOrder quantity
	takerOrder.RemainingSellQuantity = takerOrder.SellQuantity
	if takerOrder.IsRemainingBuyQuantityLessThanOne() {
		app.CancelOrder(takerOrder)
		return
	}

	if len(makerOB) == 0 {
		takerOB = append(takerOB, takerOrder)
	} else {
		makerOB, takerOB = app.iterateMakerOrderBook(takerOrder, makerOB, takerOB)
	}

	// sort orders by price
	sort.Slice(takerOB, func(i, j int) bool {
		return takerOB[i].Price.LTE(takerOB[j].Price)
	})
	sort.Slice(makerOB, func(i, j int) bool {
		return makerOB[i].Price.LTE(makerOB[j].Price)
	})

	app.OrderBooks[takerOKKey] = takerOB
	app.OrderBooks[makerOBKey] = makerOB
	app.PrintOrderBooks(takerOKKey, makerOBKey)
	app.PrintBalances()
}

func (app *App) iterateMakerOrderBook(takerOrder Order, makerOB, takerOB []Order) ([]Order, []Order) {
	makerOBIndexesToRemove := make(map[int]struct{})

LOOP:
	for i, makerOrder := range makerOB {
		takerBuyPrice := (&big.Rat{}).SetFrac(DecPrecisionReuse, takerOrder.Price.BigInt())
		makerSellPrice := (&big.Rat{}).SetFrac(makerOrder.Price.BigInt(), DecPrecisionReuse)

		if takerBuyPrice.Cmp(makerSellPrice) == -1 {
			takerOB = append(takerOB, takerOrder)
			break
		}

		// this amount uses the maker price since it's better or equal
		takerBuyAmount := (&big.Rat{}).Quo((&big.Rat{}).SetInt(takerOrder.RemainingSellQuantity.BigInt()), makerSellPrice)
		makerSellAmount := (&big.Rat{}).SetInt(makerOrder.RemainingSellQuantity.BigInt())

		fmt.Printf(
			"Match (%s/%s): takerBuyPrice:%s >= makerSellPrice:%s | takerBuyAmount: %s | makerSellAmount:%s \n",
			takerOrder.ID, makerOrder.ID, takerBuyPrice.FloatString(10), makerSellPrice.FloatString(10),
			takerBuyAmount.FloatString(10), makerSellAmount.FloatString(10),
		)

		switch takerBuyAmount.Cmp(makerSellAmount) {
		case -1: // the maker order remains, the taker is reduced fully
			// taker receives the sold by maker price tokens
			takerReceiveAmount := RatAmountToIntRoundDown(takerBuyAmount)
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))
			// maker receives the taker quantity
			makerReceiveAmount := takerOrder.RemainingSellQuantity
			app.SendCoin(makerOrder.Account, sdk.NewCoin(makerOrder.BuyDenom, makerReceiveAmount))
			// update state
			makerOrder.RemainingSellQuantity = makerOrder.RemainingSellQuantity.Sub(takerReceiveAmount)
			if makerOrder.IsRemainingBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration and remove
				app.CancelOrder(makerOrder)
				makerOBIndexesToRemove[i] = struct{}{}
			} else {
				makerOB[i] = makerOrder
			}
			break LOOP
		case 0: // both orders are reduced
			app.SendCoin(makerOrder.Account, sdk.NewCoin(makerOrder.BuyDenom, takerOrder.RemainingSellQuantity))
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, makerOrder.RemainingSellQuantity))
			// remove reduced record
			makerOBIndexesToRemove[i] = struct{}{}
			break LOOP
		case 1: // the taker order remains and will go to the next loop, the maker is reduced fully
			// taker receives the amount maker sells
			takerReceiveAmount := sdk.NewIntFromBigInt(makerOrder.RemainingSellQuantity.BigInt())
			app.SendCoin(takerOrder.Account, sdk.NewCoin(takerOrder.BuyDenom, takerReceiveAmount))
			// maker receive the amount
			makerReceiveAmount := RatAmountToIntRoundDown((&big.Rat{}).Mul(makerSellAmount, makerSellPrice))
			app.SendCoin(makerOrder.Account, sdk.NewCoin(makerOrder.BuyDenom, makerReceiveAmount))
			// update state
			takerOrder.RemainingSellQuantity = takerOrder.RemainingSellQuantity.Sub(makerReceiveAmount)
			// remove reduced record
			makerOBIndexesToRemove[i] = struct{}{}

			if takerOrder.IsRemainingBuyQuantityLessThanOne() {
				// cancel since nothing to use for the next iteration
				app.CancelOrder(takerOrder)
				break LOOP
			}
			// if nothing to match with add remaining taker order
			if len(makerOB) == len(makerOBIndexesToRemove) {
				takerOB = append(takerOB, takerOrder)
			}
		}
	}
	updatedMakerOB := make([]Order, 0)
	for i, order := range makerOB {
		if _, ok := makerOBIndexesToRemove[i]; ok {
			continue
		}
		updatedMakerOB = append(updatedMakerOB, order)
	}

	return updatedMakerOB, takerOB
}

// CancelOrder sends order coins to the creator.
func (app *App) CancelOrder(order Order) {
	fmt.Printf("\nRemaining buy quantity is less than one, order canceled: %s\n", order.String())
	app.SendCoin(order.Account, sdk.NewCoin(order.SellDenom, order.RemainingSellQuantity))
}

// SendCoin sends coins to sample app accounts.
func (app *App) SendCoin(recipient string, amt sdk.Coin) {
	accountBalances, ok := app.Balances[recipient]
	if !ok {
		accountBalances = make(sdk.Coins, 0)
	}
	app.Balances[recipient] = accountBalances.Add(amt)
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
		if !ok || len(ob) == 0 {
			fmt.Println("Empty...")
			continue
		}
		for i, order := range ob {
			fmt.Printf("Order [%d]: %s\n", i, order.String())
		}
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
