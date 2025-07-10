package matchingengine

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v6/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
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

// MatchOrder matches an incoming order against the orders present in the order book storage.
func (me MatchingEngine) MatchOrder(
	ctx sdk.Context,
	accNumber uint64,
	orderBookID uint32,
	takerOrder types.Order,
	initialRemainingBalance sdkmath.Int,
) (MatchingResult, error) {
	mr, err := NewMatchingResult(takerOrder)
	if err != nil {
		return MatchingResult{}, err
	}

	takerRecord := convertOrderToOrderBookRecord(accNumber, orderBookID, takerOrder, initialRemainingBalance)

	takerIsFilled := false
	for {
		makerRecord, matches, err := me.obq.Next()
		if err != nil {
			return MatchingResult{}, err
		}
		if !matches {
			break
		}
		takerIsFilled, err = me.matchRecords(ctx, &mr, &takerRecord, &makerRecord, takerOrder)
		if err != nil {
			return MatchingResult{}, err
		}
		if takerIsFilled {
			break
		}
	}

	mr.TakerIsFilled = takerIsFilled
	mr.TakerRecord = takerRecord

	return mr, nil
}

func (me MatchingEngine) matchRecords(
	ctx sdk.Context,
	mr *MatchingResult,
	takerRecord, makerRecord *types.OrderBookRecord,
	takerOrder types.Order,
) (bool, error) {
	me.logger.Debug(
		"Matching OB records.",
		"takerRecord", takerRecord.String(),
		"makerRecord", makerRecord.String(),
	)

	takerReceivesDenom, takerSpendsDenom := takerOrder.BaseDenom, takerOrder.QuoteDenom
	if takerOrder.Side == types.SIDE_SELL {
		takerReceivesDenom, takerSpendsDenom = takerOrder.QuoteDenom, takerOrder.BaseDenom
	}

	isMakerInverted := takerRecord.Side == makerRecord.Side

	takerRecordForMatching := newMatchingOBRecord(takerRecord, false)
	makerRecordForMatching := newMatchingOBRecord(makerRecord, isMakerInverted)
	trade, closeResult := match(takerRecordForMatching, makerRecordForMatching)
	me.logger.Debug(
		"Matching result.",
		"trade", trade,
		"closeResult", closeResult.String(),
	)

	// Send funds
	makerAddr, err := me.ak.GetAccountAddress(ctx, makerRecord.AccountNumber)
	if err != nil {
		return false, err
	}
	mr.SendFromTaker(
		makerAddr,
		makerRecord.OrderID,
		makerRecord.OrderSequence,
		sdk.NewCoin(takerSpendsDenom, sdkmath.NewIntFromBigInt(trade.TakerSpends)),
	)
	mr.SendFromMaker(
		makerAddr,
		makerRecord.OrderID,
		sdk.NewCoin(takerReceivesDenom, sdkmath.NewIntFromBigInt(trade.TakerReceives)),
	)

	// Reduce taker
	takerRecord.RemainingBaseQuantity = takerRecord.RemainingBaseQuantity.Sub(
		sdkmath.NewIntFromBigInt(trade.BaseQuantity))
	takerRecord.RemainingSpendableBalance = takerRecord.RemainingSpendableBalance.Sub(
		sdkmath.NewIntFromBigInt(trade.TakerSpends))

	// Reduce maker
	if !isMakerInverted {
		makerRecord.RemainingBaseQuantity = makerRecord.RemainingBaseQuantity.Sub(
			sdkmath.NewIntFromBigInt(trade.BaseQuantity))
		makerRecord.RemainingSpendableBalance = makerRecord.RemainingSpendableBalance.Sub(
			sdkmath.NewIntFromBigInt(trade.TakerReceives))
	} else {
		makerRecord.RemainingBaseQuantity = makerRecord.RemainingBaseQuantity.Sub(
			sdkmath.NewIntFromBigInt(trade.QuoteQuantity))
		makerRecord.RemainingSpendableBalance = makerRecord.RemainingSpendableBalance.Sub(
			sdkmath.NewIntFromBigInt(trade.TakerReceives))
	}

	me.logger.Debug(
		"Matched OB records after reduction.",
		"takerRecord", takerRecord.String(),
		"makerRecord", makerRecord.String(),
	)

	// Close or update maker record
	if closeResult == closeMaker || closeResult == closeBoth || !isOrderRecordExecutableAsMaker(makerRecord) {
		lockedCoins, expectedToReceiveCoin, err := me.getMakerLockedAndExpectedToReceiveCoins(
			ctx,
			makerRecord,
			takerReceivesDenom,
			takerSpendsDenom,
		)
		if err != nil {
			return false, err
		}

		mr.DecreaseMakerLimits(makerAddr, lockedCoins, expectedToReceiveCoin)
		mr.RemoveRecord(makerAddr, makerRecord)
	} else {
		mr.UpdateRecord(*makerRecord)
	}

	// We continue only if closeResult shouldn't close the taker record
	return closeResult == closeTaker || closeResult == closeBoth, nil
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
	// Because if we do it using integers it may cause rounding and rational reminder
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

	// But for matching execution we use integers.
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

func (me MatchingEngine) getMakerLockedAndExpectedToReceiveCoins(
	ctx sdk.Context,
	makerRecord *types.OrderBookRecord,
	makerSpendsDenom, makerReceivesDenom string,
) (sdk.Coins, sdk.Coin, error) {
	// Return non-executed balance
	lockedCoins := sdk.NewCoins(
		sdk.NewCoin(makerSpendsDenom, makerRecord.RemainingSpendableBalance),
	)
	// TODO(milad): move GetOrderData out of the matching engine.
	recordToCloseOrderData, err := me.dexKeeper.GetOrderData(ctx, makerRecord.OrderSequence)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	// Return order reserve if any
	if recordToCloseOrderData.Reserve.IsPositive() {
		lockedCoins = lockedCoins.Add(recordToCloseOrderData.Reserve)
	}

	expectedToReceiveAmt, err := types.ComputeLimitOrderExpectedToReceiveAmount(
		makerRecord.Side, makerRecord.RemainingBaseQuantity, makerRecord.Price,
	)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	expectedToReceiveCoin := sdk.NewCoin(makerReceivesDenom, expectedToReceiveAmt)

	return lockedCoins, expectedToReceiveCoin, nil
}

// newMatchingOBRecord initializes struct for matching engine and inverts order if needed.
//
// E.g.:
//
// original order: market=USD/BTC buy 50 USD for 0.04 BTC per USD
// RemainingBaseQuantity: 50 USD
// RemainingSpendBalance: 2 BTC
//
// inverted order: market=BTC/USD sell 2 BTC for 25 USD per BTC
// RemainingBaseQuantity: 2 BTC
// RemainingSpendBalance: 2 BTC

// original order: market=USD/BTC sell 50 USD for 0.04 BTC per USD
// RemainingBaseQuantity: 50 USD
// RemainingSpendBalance: 50 USD
//
// inverted order: market=BTC/USD buy 2 BTC for 25 USD per BTC
// RemainingBaseQuantity: 2 BTC
// RemainingSpendBalance: 50 USD.
func newMatchingOBRecord(obRecord *types.OrderBookRecord, inverted bool) OBRecord {
	side := sellOrderSide
	if obRecord.Side == types.SIDE_BUY {
		side = buyOrderSide
	}

	price := obRecord.Price.Rat()
	if cbig.RatIsZero(price) {
		price = marketOrderPrice
	}

	if !inverted {
		return OBRecord{
			Side:         side,
			Price:        price,
			BaseQuantity: cbig.NewRatFromBigInt(obRecord.RemainingBaseQuantity.BigInt()),
			SpendBalance: cbig.NewRatFromBigInt(obRecord.RemainingSpendableBalance.BigInt()),
		}
	}

	baseQuantity := cbig.RatMul(cbig.NewRatFromBigInt(obRecord.RemainingBaseQuantity.BigInt()), price)
	return OBRecord{
		Side:         side.Opposite(),
		Price:        cbig.RatInv(price),
		BaseQuantity: baseQuantity,
		SpendBalance: cbig.NewRatFromBigInt(obRecord.RemainingSpendableBalance.BigInt()),
	}
}

// OBRecord is an order book record for matching engine.
type OBRecord struct {
	Side         OrderSide
	Price        *big.Rat
	BaseQuantity *big.Rat
	SpendBalance *big.Rat
}

// MaxBaseQuantityForPrice returns maximum base quantity order can execute for a given price.
// Price is expected to be better or equal to OBRecord.Price.
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

// isOrderRecordExecutableAsMaker returns true if RemainingBaseQuantity inside order is executable with order price.
// Order with RemainingBaseQuantity: 101 and Price: 0.397 is not executable as maker:
// Qa' = floor(Qa / pd) * pd = floor(101 / 397) * 1000 = 0.
//
// Order with RemainingBaseQuantity: 101 and Price: 0.39 is executable:
// Qa' = floor(Qa / pd) * pd = floor(101 / 39) * 100 > 0.
//
// This func logic might be revised if we introduce proper ticks for price & quantity.
func isOrderRecordExecutableAsMaker(obRecord *types.OrderBookRecord) bool {
	baseQuantity, _ := computeMaxIntExecutionQuantity(obRecord.Price.Rat(), obRecord.RemainingBaseQuantity.BigInt())
	return !cbig.IntEqZero(baseQuantity)
}

func computeMaxIntExecutionQuantity(priceRat *big.Rat, baseQuantity *big.Int) (*big.Int, *big.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()

	n := cbig.IntQuo(baseQuantity, priceDenom)
	baseQuantityInt := cbig.IntMul(n, priceDenom)
	quoteQuantityInt := cbig.IntMul(n, priceNum)

	return baseQuantityInt, quoteQuantityInt
}
