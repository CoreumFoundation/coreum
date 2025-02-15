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
	BaseQuantity *big.Rat
	SpendBalance *big.Rat
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

	takerMaxBaseQuantityRat := takerRecord.MaxBaseQuantityForPrice(trade.Price)
	makerMaxBaseQuantityRat := makerRecord.MaxBaseQuantityForPrice(trade.Price)

	var baseQuantityRat *big.Rat
	closeResult := NoneCloseType

	// Note that we compare max execution quantities for each record as rational.
	// Because if we do it using integers it may cause roudning and rational reminder
	// of a bigger order might be executable with the next order.
	if cbig.RatLT(takerMaxBaseQuantityRat, makerMaxBaseQuantityRat) {
		closeResult = CloseTaker
		baseQuantityRat = takerMaxBaseQuantityRat
	} else if cbig.RatEQ(takerMaxBaseQuantityRat, makerMaxBaseQuantityRat) {
		closeResult = CloseBoth
		baseQuantityRat = takerMaxBaseQuantityRat
	} else {
		closeResult = CloseMaker
		baseQuantityRat = makerMaxBaseQuantityRat
	}

	// But for matching executeion we use integers.
	// TODO(ysv): Take zero quantity into consideration.
	trade.BaseQuantity, trade.QuoteQuantity = ComputeMaxIntExecutionQuantityV2(
		trade.Price,
		cbig.IntQuo(baseQuantityRat.Num(), baseQuantityRat.Denom()),
	)

	if takerRecord.Side == SellOrderSide {
		trade.TakerSpends = trade.BaseQuantity
		trade.TakerReceives = trade.QuoteQuantity
	} else {
		trade.TakerSpends = trade.QuoteQuantity
		trade.TakerReceives = trade.BaseQuantity
	}

	return trade, closeResult, nil
}

func (obr OBRecord) MaxBaseQuantityForPrice(price *big.Rat) *big.Rat {
	// For limit order we execute BaseQuantity fully.
	if obr.IsLimit() {
		return obr.BaseQuantity
	}

	// For market sell orders we execute up to BaseQuantity or SpendBalance, whichever is filled first.
	if obr.Side == SellOrderSide {
		return cbig.RatMin(obr.BaseQuantity, obr.SpendBalance)
	}

	// For market buy orders we execute up to BaseQuantity or SpendBalance / Price, whichever is filled first.
	// SpendBalance / Price = SpendBalance * Price^-1
	maxQuantityFromBalance := cbig.RatMul(obr.SpendBalance, cbig.RatInv(price))

	return cbig.RatMin(obr.BaseQuantity, maxQuantityFromBalance)
}

// func (obr OBRecord) MaxBaseQuantityForPriceLeg(price *big.Rat) *big.Int {
// 	// For limit order we execute BaseQuantity fully.
// 	if obr.IsLimit() {
// 		maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, obr.BaseQuantity)
// 		return maxBaseQuantity
// 	}

// 	// For market sell orders we execute up to BaseQuantity or SpendBalance, whichever is filled first.
// 	if obr.Side == SellOrderSide {
// 		maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, cbig.IntMin(obr.BaseQuantity, obr.SpendBalance))
// 		return maxBaseQuantity
// 	}

// 	// For market buy orders we execute up to BaseQuantity or SpendBalance / Price, whichever is filled first.
// 	// SpendBalance / Price = SpendBalance * Price^-1
// 	maxQuantityFromBalance, _ := cbig.IntMulRatWithRemainder(obr.SpendBalance, cbig.RatInv(price))

// 	maxBaseQuantity, _ := computeMaxIntExecutionQuantityV2(price, cbig.IntMin(obr.BaseQuantity, maxQuantityFromBalance))
// 	return maxBaseQuantity
// }

func ComputeMaxIntExecutionQuantityV2(priceRat *big.Rat, baseQuantity *big.Int) (*big.Int, *big.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()

	n := cbig.IntQuo(baseQuantity, priceDenom)
	baseIntQuantity := cbig.IntMul(n, priceDenom)
	quoteIntQuantity := cbig.IntMul(n, priceNum)

	return baseIntQuantity, quoteIntQuantity
}

func (obr OBRecord) IsLimit() bool {
	return !cbig.RatEQ(obr.Price, MarketOrderPrice)
}
