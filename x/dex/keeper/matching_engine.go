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

func (cr CloseResult) String() string {
	switch cr {
	case CloseTaker:
		return "CloseTaker"
	case CloseMaker:
		return "CloseMaker"
	case CloseBoth:
		return "CloseBoth"
	default:
		return "None"
	}
}

func match(takerRecord, makerRecord OBRecord) (Trade, CloseResult, error) {
	if takerRecord.Side == makerRecord.Side {
		return Trade{}, NoneCloseType, nil
	}

	trade := Trade{Price: makerRecord.Price}

	takerMaxBaseQuantity := takerRecord.MaxBaseQuantityForPrice(trade.Price)
	makerMaxBaseQuantity := makerRecord.MaxBaseQuantityForPrice(trade.Price)
	closeResult := NoneCloseType

	// TODO(ysv): Consider zero quantity.
	if cbig.IntLT(takerMaxBaseQuantity, makerMaxBaseQuantity) {
		closeResult = CloseTaker
		trade.BaseQuantity = takerMaxBaseQuantity
	} else if cbig.IntEQ(takerMaxBaseQuantity, makerMaxBaseQuantity) {
		closeResult = CloseBoth
		trade.BaseQuantity = takerMaxBaseQuantity
	} else {
		closeResult = CloseMaker
		trade.BaseQuantity = makerMaxBaseQuantity
	}

	// BaseQuantity is calculated in the way so that QuoteQuantity = BaseQuantity * Price is always integer.
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

func (obr OBRecord) MaxBaseQuantityForPrice(price *big.Rat) *big.Int {
	// For limit order we execute BaseQuantity fully.
	if obr.IsLimit() {
		maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, obr.BaseQuantity)
		return maxBaseQuantity
	}

	// For market sell orders we execute up to BaseQuantity or SpendBalance, whichever is filled first.
	if obr.Side == SellOrderSide {
		maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, cbig.IntMin(obr.BaseQuantity, obr.SpendBalance))
		return maxBaseQuantity
	}

	// For market buy orders we execute up to BaseQuantity or SpendBalance / Price, whichever is filled first.
	// RemainingBalance / Price = RemainingBalance * Price^-1
	maxQuantityFromBalance, _ := cbig.IntMulRatWithRemainder(obr.SpendBalance, cbig.RatInv(price))

	maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, cbig.IntMin(obr.BaseQuantity, maxQuantityFromBalance))
	return maxBaseQuantity
}

func computeMaxIntExecutionQuantityV2(priceRat *big.Rat, baseQuantity *big.Int) (*big.Int, *big.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()

	n := cbig.IntQuo(baseQuantity, priceDenom)
	baseMaxQuantity := cbig.IntMul(n, priceDenom)
	quoteMaxQuantity := cbig.IntMul(n, priceNum)

	return baseMaxQuantity, quoteMaxQuantity
}

func (obr OBRecord) IsLimit() bool {
	return !cbig.RatEQ(obr.Price, MarketOrderPrice)
}
