package keeper

import (
	"math/big"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
)

var MarketOrderPrice = big.NewRat(-1, 1)

type Trade struct {
	Quantity *big.Int
	Price    *big.Rat
}

type OrderSide int

const (
	BuyOrderSide OrderSide = iota
	SellOrderSide
)

type OBRecord struct {
	Side              OrderSide
	Price             *big.Rat
	RemainingQuantity *big.Int
	RemainingBalance  *big.Int
}

type MatchType int

const (
	NoneMatchType MatchType = iota
	CloseTaker
	CloseMaker
	CloseBoth
)

func match(takerRecord, makerRecord OBRecord) (Trade, MatchType, error) {
	if takerRecord.Side == makerRecord.Side {
		return Trade{}, NoneMatchType, nil
	}

	trade := Trade{Price: makerRecord.Price}

	// TODO(ysv): Consider zero quantity.
	takerQuantity := takerRecord.MaxTakerQuantityForPrice(makerRecord.Price)

	matchType := NoneMatchType
	if cbig.IntLT(takerQuantity, makerRecord.RemainingQuantity) {
		matchType = CloseTaker
		trade.Quantity = takerQuantity
	} else if cbig.IntEQ(takerQuantity, makerRecord.RemainingQuantity) {
		matchType = CloseBoth
		trade.Quantity = takerQuantity
	} else {
		matchType = CloseMaker
		trade.Quantity = makerRecord.RemainingQuantity
	}

	return trade, matchType, nil
}

func (obr OBRecord) MaxTakerQuantityForPrice(price *big.Rat) *big.Int {
	// For limit order we execute RemainingQuantity fully.
	if obr.IsLimit() {
		return obr.RemainingQuantity
	}

	// For market sell orders we execute up to RemainingQuantity or RemainingBalance, whichever is filled first.
	if obr.Side == SellOrderSide {
		return cbig.IntMin(obr.RemainingQuantity, obr.RemainingBalance)
	}

	// For market buy orders we execute up to RemainingQuantity or RemainingBalance / Price, whichever is filled first.
	// RemainingBalance / Price = RemainingBalance * Price^-1
	// Reminder is not executable so we just round it down.
	maxQuantityFromBalance, _ := cbig.IntMulRatWithRemainder(obr.RemainingBalance, cbig.RatInv(price))

	return cbig.IntMin(obr.RemainingQuantity, maxQuantityFromBalance)
}

func (obr OBRecord) IsLimit() bool {
	return !cbig.RatEQ(obr.Price, MarketOrderPrice)
}
