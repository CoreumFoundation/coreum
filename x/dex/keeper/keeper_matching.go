package keeper

import (
	"fmt"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) matchOrder(
	ctx sdk.Context,
	params types.Params,
	accNumber uint64,
	orderBookID, invertedOrderBookID uint32,
	takerOrder types.Order,
) error {
	k.logger(ctx).Debug("Matching order.", "order", takerOrder.String())

	mf, err := k.NewMatchingFinder(ctx, orderBookID, invertedOrderBookID, takerOrder)
	if err != nil {
		return err
	}
	defer func() {
		if err := mf.Close(); err != nil {
			k.logger(ctx).Error(err.Error())
		}
	}()

	takerRecord, err := k.initTakerRecord(ctx, accNumber, orderBookID, takerOrder)
	if err != nil {
		return err
	}
	takerOrder.Sequence = takerRecord.OrderSequence

	if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderPlaced{
		Creator:  takerOrder.Creator,
		ID:       takerOrder.ID,
		Sequence: takerRecord.OrderSequence,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderPlaced: %s", err)
	}

	mr, err := NewMatchingResult(takerOrder)
	if err != nil {
		return err
	}

	cak := newCachedAccountKeeper(k.accountKeeper, k.accountQueryServer)

	takerIsFilled := false
	for {
		makerRecord, matches, err := mf.Next()
		if err != nil {
			return err
		}
		if !matches {
			break
		}
		takerIsFilled, err = k.mathcRecordsV2(ctx, cak, mr, &takerRecord, &makerRecord, takerOrder)
		if err != nil {
			return err
		}
		if takerIsFilled {
			break
		}
	}

	switch takerOrder.Type {
	case types.ORDER_TYPE_LIMIT:
		switch takerOrder.TimeInForce {
		case types.TIME_IN_FORCE_GTC:
			// If taker order is filled fully or not executable as maker we just apply matching result and return.
			if takerIsFilled || !isOrderRecordExecutableAsMaker(&takerRecord) {
				return k.applyMatchingResult(ctx, mr)
			}

			// If taker orders is not filled fully we need to:
			// - increase taker limits for record for remaining amount
			// - apply matching result
			// - add remaining order to the order book
			if err := mr.IncreaseTakerLimitsForRecord(params, takerOrder, &takerRecord); err != nil {
				return err
			}

			if err := k.applyMatchingResult(ctx, mr); err != nil {
				return err
			}

			return k.createOrder(ctx, params, takerOrder, takerRecord)
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr)
		case types.TIME_IN_FORCE_FOK:
			// ensure full order fill
			if takerRecord.RemainingQuantity.IsPositive() {
				return nil
			}
			return k.applyMatchingResult(ctx, mr)
		default:
			return sdkerrors.Wrapf(types.ErrInvalidInput, "unsupported time in force: %s", takerOrder.TimeInForce.String())
		}
	case types.ORDER_TYPE_MARKET:
		return k.applyMatchingResult(ctx, mr)
	default:
		return sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type: %s", takerOrder.Type.String(),
		)
	}
}

func (k Keeper) initTakerRecord(
	ctx sdk.Context,
	accNumber uint64,
	orderBookID uint32,
	order types.Order,
) (types.OrderBookRecord, error) {
	remainingBalance, err := k.getInitialRemainingBalance(ctx, order)
	if err != nil {
		return types.OrderBookRecord{}, err
	}

	var price types.Price
	if order.Price != nil {
		price = *order.Price
	}

	orderSequence, err := k.genNextOrderSequence(ctx)
	if err != nil {
		return types.OrderBookRecord{}, err
	}

	return types.OrderBookRecord{
		OrderBookID:       orderBookID,
		Side:              order.Side,
		Price:             price,
		OrderSequence:     orderSequence,
		OrderID:           order.ID,
		AccountNumber:     accNumber,
		RemainingQuantity: order.Quantity,
		RemainingBalance:  remainingBalance,
	}, nil
}

func (k Keeper) getInitialRemainingBalance(
	ctx sdk.Context,
	order types.Order,
) (sdkmath.Int, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdkmath.Int{}, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	var remainingBalance sdk.Coin
	switch order.Type {
	case types.ORDER_TYPE_LIMIT:
		var err error
		remainingBalance, err = order.ComputeLimitOrderLockedBalance()
		if err != nil {
			return sdkmath.Int{}, err
		}
	case types.ORDER_TYPE_MARKET:
		spendableBalance, err := k.assetFTKeeper.GetSpendableBalance(ctx, creatorAddr, order.GetSpendDenom())
		if err != nil {
			return sdkmath.Int{}, err
		}

		// For market buy order we lock whole spendable balance.
		remainingBalance = spendableBalance

		// For market sell order we lock min of spendable balance or order quantity.
		if order.Side == types.SIDE_SELL && order.Quantity.LT(spendableBalance.Amount) {
			remainingBalance = sdk.NewCoin(remainingBalance.Denom, order.Quantity)
		}
	default:
		return sdkmath.Int{}, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type : %s", order.Type.String(),
		)
	}

	k.logger(ctx).Debug("Got initial remaining balance.", "remainingBalance", remainingBalance)

	return remainingBalance.Amount, nil
}

func (k Keeper) mathcRecordsV2(
	ctx sdk.Context,
	cak cachedAccountKeeper,
	mr *MatchingResult,
	takerRecord, makerRecord *types.OrderBookRecord,
	takerOrder types.Order,
) (bool, error) {
	fmt.Printf("matching records: \ntakerRecord: %v \nmakerRecord: %v\n", takerRecord.String(), makerRecord.String())
	takerReceivesDenom, takerSpendsDenom := takerOrder.BaseDenom, takerOrder.QuoteDenom
	if takerOrder.Side == types.SIDE_SELL {
		takerReceivesDenom, takerSpendsDenom = takerOrder.QuoteDenom, takerOrder.BaseDenom
	}

	isMakerInverted := takerRecord.Side == makerRecord.Side

	takerRecordForMatching := newMatchingOBRecord(takerRecord, false)
	makerRecordForMatching := newMatchingOBRecord(makerRecord, isMakerInverted)
	trade, closeResult, err := match(takerRecordForMatching, makerRecordForMatching)
	if err != nil {
		return false, err
	}
	fmt.Printf("resulting closeResult: %v \ntrade: %+v\n", closeResult.String(), trade)

	// Exchange funds
	makerAddr, err := cak.getAccountAddressWithCache(ctx, makerRecord.AccountNumber)
	if err != nil {
		return false, err
	}
	mr.SendFromTaker(
		makerAddr, makerRecord.OrderID, makerRecord.OrderSequence, sdk.NewCoin(takerSpendsDenom, sdkmath.NewIntFromBigInt(trade.TakerSpends)),
	)
	mr.SendFromMaker(
		makerAddr, makerRecord.OrderID, sdk.NewCoin(takerReceivesDenom, sdkmath.NewIntFromBigInt(trade.TakerReceives)),
	)

	// Reduce taker
	takerRecord.RemainingQuantity = takerRecord.RemainingQuantity.Sub(sdkmath.NewIntFromBigInt(trade.BaseQuantity))
	takerRecord.RemainingBalance = takerRecord.RemainingBalance.Sub(sdkmath.NewIntFromBigInt(trade.TakerSpends))

	// Reduce maker
	if !isMakerInverted {
		makerRecord.RemainingQuantity = makerRecord.RemainingQuantity.Sub(sdkmath.NewIntFromBigInt(trade.BaseQuantity))
		makerRecord.RemainingBalance = makerRecord.RemainingBalance.Sub(sdkmath.NewIntFromBigInt(trade.TakerReceives))
	} else {
		makerRecord.RemainingQuantity = makerRecord.RemainingQuantity.Sub(sdkmath.NewIntFromBigInt(trade.QuoteQuantity))
		makerRecord.RemainingBalance = makerRecord.RemainingBalance.Sub(sdkmath.NewIntFromBigInt(trade.TakerReceives))
	}

	fmt.Printf("records after reducing: \ntakerRecord: %v \nmakerRecord: %v\n", takerRecord.String(), makerRecord.String())

	// Close or update maker record
	if closeResult == CloseMaker || closeResult == CloseBoth || !isOrderRecordExecutableAsMaker(makerRecord) {
		lockedCoins, expectedToReceiveCoin, err := k.getMakerLockedAndExpectedToReceiveCoinsV2(ctx, makerRecord, takerReceivesDenom, takerSpendsDenom)
		if err != nil {
			return false, err
		}

		mr.DecreaseMakerLimits(makerAddr, lockedCoins, expectedToReceiveCoin)
		mr.RemoveRecord(makerAddr, makerRecord)
	} else {
		mr.UpdateRecord(*makerRecord)
	}

	// We continue only if closeResult shouldn't close the taker record
	return closeResult == CloseTaker || closeResult == CloseBoth, nil
}

// isOrderRecordExecutableAsMaker returns true if RemainingQuantity inside order is executable with order price.
// Order with RemainingQuantity: 101 and Price: 0.397 is not executable as maker:
// Qa' = floor(Qa / pd) * pd = floor(101 / 397) * 1000 = 0.
//
// Order with RemainingQuantity: 101 and Price: 0.39 is executable:
// Qa' = floor(Qa / pd) * pd = floor(101 / 39) * 100 > 0.
//
// This func logic might be rewised if we introduce proper ticks for price & quantity.
func isOrderRecordExecutableAsMaker(obRecord *types.OrderBookRecord) bool {
	baseQuantity, _ := computeMaxIntExecutionQuantity(obRecord.Price.Rat(), obRecord.RemainingQuantity.BigInt())
	return !cbig.IntEqZero(baseQuantity)
}

//
// original order: market=USD/BTC buy 50 USD for 0.04 BTC per USD
// RemeaningBaseQuantity: 50 USD
// RemeaningSpendBalance: 2 BTC
//
// inverted order: market=BTC/USD sell 2 BTC for 25 USD per BTC
// RemeaningBaseQuantity: 2 BTC
// RemeaningSpendBalance: 2 BTC

// original order: market=USD/BTC sell 50 USD for 0.04 BTC per USD
// RemeaningBaseQuantity: 50 USD
// RemeaningSpendBalance: 50 USD
//
// inverted order: market=BTC/USD buy 2 BTC for 25 USD per BTC
// RemeaningBaseQuantity: 2 BTC
// RemeaningSpendBalance: 50 USD
func newMatchingOBRecord(obRecord *types.OrderBookRecord, inverted bool) OBRecord {
	side := SellOrderSide
	if obRecord.Side == types.SIDE_BUY {
		side = BuyOrderSide
	}

	price := obRecord.Price.Rat()
	if cbig.RatIsZero(price) {
		price = MarketOrderPrice
	}

	if !inverted {
		return OBRecord{
			Side:         side,
			Price:        price,
			BaseQuantity: cbig.NewRatFromBigInt(obRecord.RemainingQuantity.BigInt()),
			SpendBalance: cbig.NewRatFromBigInt(obRecord.RemainingBalance.BigInt()),
		}
	}

	baseQuantity := cbig.RatMul(cbig.NewRatFromBigInt(obRecord.RemainingQuantity.BigInt()), price)
	return OBRecord{
		Side:         side.Opposite(),
		Price:        cbig.RatInv(price),
		BaseQuantity: baseQuantity,
		SpendBalance: cbig.NewRatFromBigInt(obRecord.RemainingBalance.BigInt()),
	}
}

func (k Keeper) getMakerLockedAndExpectedToReceiveCoinsV2(
	ctx sdk.Context,
	makerRecord *types.OrderBookRecord,
	makerSpendsDenom, makerReceivesDenom string,
) (sdk.Coins, sdk.Coin, error) {
	// Return non-executed balance
	lockedCoins := sdk.NewCoins(
		sdk.NewCoin(makerSpendsDenom, makerRecord.RemainingBalance),
	)
	recordToCloseOrderData, err := k.getOrderData(ctx, makerRecord.OrderSequence)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	// Return order reserve if any
	if recordToCloseOrderData.Reserve.IsPositive() {
		lockedCoins = lockedCoins.Add(recordToCloseOrderData.Reserve)
	}

	// TODO(ysv): Not sure if this part is correct. Double check.
	expectedToReceiveAmt, err := types.ComputeLimitOrderExpectedToReceiveAmount(
		makerRecord.Side, makerRecord.RemainingQuantity, makerRecord.Price,
	)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	expectedToReceiveCoin := sdk.NewCoin(makerReceivesDenom, expectedToReceiveAmt)

	return lockedCoins, expectedToReceiveCoin, nil
}

func computeMaxIntExecutionQuantity(priceRat *big.Rat, baseQuantity *big.Int) (*big.Int, *big.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()

	n := cbig.IntQuo(baseQuantity, priceDenom)
	baseIntQuantity := cbig.IntMul(n, priceDenom)
	quoteIntQuantity := cbig.IntMul(n, priceNum)

	return baseIntQuantity, quoteIntQuantity
}
