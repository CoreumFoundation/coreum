package keeper

import (
	"math/big"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
)

var MarketOrderPrice = big.NewRat(-1, 1)

type Trade struct {
	BaseQuantity  *big.Int
	QuoteQuantity *big.Int
	Price         *big.Rat

	TakerReceives *big.Int
	TakerSpends   *big.Int
}

type OrderSide int

const (
	BuyOrderSide OrderSide = iota
	SellOrderSide
)

func (o OrderSide) Opposite() OrderSide {
	if o == SellOrderSide {
		return BuyOrderSide
	}

	return SellOrderSide
}

type OBRecord struct {
	Side         OrderSide
	Price        *big.Rat
	BaseQuantity *big.Int
	SpendBalance *big.Int
}

type CloseResult int

const (
	NoneCloseType CloseResult = iota
	CloseTaker
	CloseMaker
	CloseBoth
)

func match(takerRecord, makerRecord OBRecord) (Trade, CloseResult, error) {
	if takerRecord.Side == makerRecord.Side {
		return Trade{}, NoneCloseType, nil
	}

	trade := Trade{Price: makerRecord.Price}

	takerBaseQuantity := takerRecord.MaxTakerBaseQuantityForPrice(trade.Price)
	closeResult := NoneCloseType
	// TODO(ysv): Consider zero quantity.
	if cbig.IntLT(takerBaseQuantity, makerRecord.BaseQuantity) {
		closeResult = CloseTaker
		trade.BaseQuantity = takerBaseQuantity
	} else if cbig.IntEQ(takerBaseQuantity, makerRecord.BaseQuantity) {
		closeResult = CloseBoth
		trade.BaseQuantity = takerBaseQuantity
	} else {
		closeResult = CloseMaker
		trade.BaseQuantity = makerRecord.BaseQuantity
	}

	trade.QuoteQuantity, _ = cbig.IntMulRatWithRemainder(trade.BaseQuantity, trade.Price)

	if takerRecord.Side == SellOrderSide {
		trade.TakerSpends = trade.BaseQuantity
		trade.TakerReceives = trade.QuoteQuantity
	} else {
		trade.TakerSpends = trade.QuoteQuantity
		trade.TakerReceives = trade.BaseQuantity
	}

	return trade, closeResult, nil
}

func (obr OBRecord) MaxTakerBaseQuantityForPrice(price *big.Rat) *big.Int {
	// For limit order we execute RemainingQuantity fully.
	if obr.IsLimit() {
		return obr.BaseQuantity
	}

	// For market sell orders we execute up to RemainingQuantity or RemainingBalance, whichever is filled first.
	if obr.Side == SellOrderSide {
		return cbig.IntMin(obr.BaseQuantity, obr.SpendBalance)
	}

	// For market buy orders we execute up to RemainingQuantity or RemainingBalance / Price, whichever is filled first.
	// RemainingBalance / Price = RemainingBalance * Price^-1
	// Reminder is not executable so we just round it down.
	maxQuantityFromBalance, _ := cbig.IntMulRatWithRemainder(obr.SpendBalance, cbig.RatInv(price))

	return cbig.IntMin(obr.BaseQuantity, maxQuantityFromBalance)
}

func (obr OBRecord) IsLimit() bool {
	return !cbig.RatEQ(obr.Price, MarketOrderPrice)
}
