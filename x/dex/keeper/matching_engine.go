package keeper

import (
	"math/big"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
)

var marketOrderPrice = big.NewRat(-1, 1)

// Trade is a matching result of two orders.
type Trade struct {
	BaseQuantity  *big.Int
	QuoteQuantity *big.Int
	Price         *big.Rat

	TakerReceives *big.Int
	TakerSpends   *big.Int
}

// OrderSide is a side of an order.
type OrderSide int

const (
	buyOrderSide OrderSide = iota
	sellOrderSide
)

// Opposite returns the opposite side of an order.
func (o OrderSide) Opposite() OrderSide {
	if o == sellOrderSide {
		return buyOrderSide
	}

	return sellOrderSide
}

// OBRecord is an order book record for matching engine.
type OBRecord struct {
	Side         OrderSide
	Price        *big.Rat
	BaseQuantity *big.Rat
	SpendBalance *big.Rat
}

// CloseResult is a result of order matching which specifies which order should be closed.
type CloseResult int

const (
	noneCloseType CloseResult = iota
	closeTaker
	closeMaker
	closeBoth
)

// String returns a string representation of CloseResult.
func (cr CloseResult) String() string {
	switch cr {
	case closeTaker:
		return "CloseTaker"
	case closeMaker:
		return "CloseMaker"
	case closeBoth:
		return "CloseBoth"
	default:
		return "None"
	}
}

func match(takerRecord, makerRecord OBRecord) (Trade, CloseResult) {
	if takerRecord.Side == makerRecord.Side {
		return Trade{}, noneCloseType
	}

	trade := Trade{Price: makerRecord.Price}

	takerMaxBaseQuantityRat := takerRecord.MaxBaseQuantityForPrice(trade.Price)
	makerMaxBaseQuantityRat := makerRecord.MaxBaseQuantityForPrice(trade.Price)

	var baseQuantityRat *big.Rat
	var closeRes CloseResult

	// Note that we compare max execution quantities for each record as rational.
	// Because if we do it using integers it may cause roudning and rational reminder
	// of a bigger order might be executable with the next order.
	switch cmp := takerMaxBaseQuantityRat.Cmp(makerMaxBaseQuantityRat); cmp {
	case -1:
		closeRes = closeTaker
		baseQuantityRat = takerMaxBaseQuantityRat
	case 0:
		closeRes = closeBoth
		baseQuantityRat = takerMaxBaseQuantityRat
	case 1:
		closeRes = closeMaker
		baseQuantityRat = makerMaxBaseQuantityRat
	}

	// But for matching executeion we use integers.
	// TODO(ysv): Take zero quantity into consideration.
	trade.BaseQuantity, trade.QuoteQuantity = computeMaxIntExecutionQuantity(
		trade.Price,
		cbig.IntQuo(baseQuantityRat.Num(), baseQuantityRat.Denom()),
	)

	if takerRecord.Side == sellOrderSide {
		trade.TakerSpends = trade.BaseQuantity
		trade.TakerReceives = trade.QuoteQuantity
	} else {
		trade.TakerSpends = trade.QuoteQuantity
		trade.TakerReceives = trade.BaseQuantity
	}

	return trade, closeRes
}

// MaxBaseQuantityForPrice returns maximum base quantity order can execute for a given price.
// Price is expected to be better or equal ot OBRecord.Price.
func (obr OBRecord) MaxBaseQuantityForPrice(price *big.Rat) *big.Rat {
	// For limit order we execute BaseQuantity fully.
	if obr.IsLimit() {
		return obr.BaseQuantity
	}

	// For market sell orders we execute up to BaseQuantity or SpendBalance, whichever is filled first.
	if obr.Side == sellOrderSide {
		return cbig.RatMin(obr.BaseQuantity, obr.SpendBalance)
	}

	// For market buy orders we execute up to BaseQuantity or SpendBalance / Price, whichever is filled first.
	// SpendBalance / Price = SpendBalance * Price^-1
	maxQuantityFromBalance := cbig.RatMul(obr.SpendBalance, cbig.RatInv(price))

	return cbig.RatMin(obr.BaseQuantity, maxQuantityFromBalance)
}

// IsLimit returns true if order is a limit order.
func (obr OBRecord) IsLimit() bool {
	return !cbig.RatEQ(obr.Price, marketOrderPrice)
}
