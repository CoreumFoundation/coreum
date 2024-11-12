package keeper

import (
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
	orderBookID, oppositeOrderBookID uint32,
	order types.Order,
) error {
	k.logger(ctx).Debug("Matching order.", "order", order.String())

	mf, err := k.NewMatchingFinder(ctx, orderBookID, oppositeOrderBookID, order)
	if err != nil {
		return err
	}
	defer func() {
		if err := mf.Close(); err != nil {
			k.logger(ctx).Error(err.Error())
		}
	}()

	takerRecord, err := k.initTakerRecord(ctx, accNumber, orderBookID, order)
	if err != nil {
		return err
	}

	mr, err := NewMatchingResult(order)
	if err != nil {
		return err
	}
	for {
		makerRecord, matches, err := mf.Next()
		if err != nil {
			return err
		}
		if !matches {
			break
		}
		stop, err := k.matchRecords(ctx, mr, &takerRecord, &makerRecord, order)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	switch order.Type {
	case types.ORDER_TYPE_LIMIT:
		switch order.TimeInForce {
		case types.TIME_IN_FORCE_GTC:
			// create new order with the updated record
			if err := k.applyMatchingResult(ctx, mr); err != nil {
				return err
			}
			if takerRecord.RemainingBalance.IsPositive() {
				return k.createOrder(ctx, params, order, takerRecord)
			}
			return nil
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr)
		case types.TIME_IN_FORCE_FOK:
			// if the order is not fill fully don't apply the matching result
			if takerRecord.RemainingQuantity.IsPositive() {
				return nil
			}
			return k.applyMatchingResult(ctx, mr)
		default:
			return sdkerrors.Wrapf(types.ErrInvalidInput, "unsupported time in force: %s", order.TimeInForce.String())
		}
	case types.ORDER_TYPE_MARKET:
		return k.applyMatchingResult(ctx, mr)
	default:
		return sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpect order type : %s", order.Type.String(),
		)
	}
}

func (k Keeper) initTakerRecord(
	ctx sdk.Context,
	accNumber uint64,
	orderBookID uint32,
	order types.Order,
) (types.OrderBookRecord, error) {
	remainingAmount, err := k.getInitialRemainingAmount(ctx, order)
	if err != nil {
		return types.OrderBookRecord{}, err
	}

	var price types.Price
	if order.Price != nil {
		price = *order.Price
	}
	return types.OrderBookRecord{
		OrderBookID:       orderBookID,
		Side:              order.Side,
		Price:             price,
		OrderSeq:          0, // set to zero and update only if we need to save it to the state
		OrderID:           order.ID,
		AccountNumber:     accNumber,
		RemainingQuantity: order.Quantity,
		RemainingBalance:  remainingAmount,
	}, nil
}

func (k Keeper) getInitialRemainingAmount(
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
		if order.Side == types.SIDE_BUY {
			remainingBalance = k.assetFTKeeper.GetSpendableBalance(ctx, creatorAddr, order.QuoteDenom)
		} else {
			remainingBalance = sdk.NewCoin(order.BaseDenom, order.Quantity)
		}
	default:
		return sdkmath.Int{}, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpect order type : %s", order.Type.String(),
		)
	}

	k.logger(ctx).Debug("Got initial remaining balance.", "remainingBalance", remainingBalance)

	return remainingBalance.Amount, nil
}

func (k Keeper) matchRecords(
	ctx sdk.Context,
	mr *MatchingResult,
	takerRecord, makerRecord *types.OrderBookRecord,
	order types.Order,
) (bool, error) {
	recordToClose, recordToReduce := k.getRecordToCloseAndReduce(ctx, takerRecord, makerRecord)
	k.logger(ctx).Debug(
		"Executing orders.",
		"recordToClose", recordToClose.String(),
		"recordToReduce", recordToReduce.String(),
	)

	recordToCloseReceiveCoin,
		recordToReduceReceiveCoin,
		recordToCloseReducedQuantity,
		recordToReduceReducedQuantity := getRecordsReceiveCoins(makerRecord, recordToClose, recordToReduce, order)

	// stop if any record receives more than opposite record balance
	// that situation is possible with the marker order with quantity which doesn't correspond the order balance
	if recordToClose.RemainingBalance.LT(recordToReduceReceiveCoin.Amount) ||
		recordToReduce.RemainingBalance.LT(recordToCloseReceiveCoin.Amount) {
		k.logger(ctx).Debug("Stop matching, order balance is not enough to cover the quantity.")
		return true, nil
	}

	recordToCloseRemainingQuantity := recordToClose.RemainingQuantity.Sub(recordToCloseReducedQuantity)
	if recordToClose.IsMaker() {
		mr.RegisterTakerCheckLimitsAndSendCoin(
			recordToClose.AccountNumber, recordToClose.OrderID, recordToCloseReceiveCoin, recordToReduceReceiveCoin,
		)
		mr.RegisterMakerUnlockAndSend(
			recordToClose.AccountNumber, recordToClose.OrderID, recordToReduceReceiveCoin, recordToCloseReceiveCoin,
		)
		unlockCoin := sdk.NewCoin(
			recordToReduceReceiveCoin.Denom, recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount),
		)
		expectedToReceiveAmt, err := types.ComputeLimitOrderExpectedToReceiveAmount(
			recordToClose.Side, recordToCloseRemainingQuantity, recordToClose.Price,
		)
		if err != nil {
			return false, err
		}
		mr.RegisterMakerUnlock(
			recordToClose.AccountNumber, unlockCoin, sdk.NewCoin(recordToCloseReceiveCoin.Denom, expectedToReceiveAmt),
		)
		mr.RegisterMakerRemoveRecord(recordToClose)
	} else {
		mr.RegisterTakerCheckLimitsAndSendCoin(
			recordToReduce.AccountNumber, recordToReduce.OrderID, recordToReduceReceiveCoin, recordToCloseReceiveCoin,
		)
		mr.RegisterMakerUnlockAndSend(
			recordToReduce.AccountNumber, recordToReduce.OrderID, recordToCloseReceiveCoin, recordToReduceReceiveCoin,
		)
	}

	recordToClose.RemainingQuantity = recordToCloseRemainingQuantity
	recordToClose.RemainingBalance = sdkmath.ZeroInt()
	k.logger(ctx).Debug("Updated recordToClose.", "recordToClose", recordToClose)

	recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(recordToReduceReducedQuantity)
	recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)
	k.logger(ctx).Debug("Updated recordToReduce.", "recordToReduce", recordToReduce)

	// continue or stop
	if recordToClose.IsMaker() {
		if recordToReduce.RemainingQuantity.IsZero() {
			k.logger(ctx).Debug("Taker record is filled fully.")
			recordToReduce.RemainingBalance = sdkmath.ZeroInt()
			return true, nil
		}
		k.logger(ctx).Debug("Going to next record in the order book.")
		return false, nil
	}

	mr.RegisterMakerUpdateRecord(*makerRecord)
	return true, nil
}

func (k Keeper) getRecordToCloseAndReduce(ctx sdk.Context, takerRecord, makerRecord *types.OrderBookRecord) (
	*types.OrderBookRecord, *types.OrderBookRecord,
) {
	var (
		recordToClose, recordToReduce  *types.OrderBookRecord
		takerVolumeRat, makerVolumeRat *big.Rat
	)

	takerVolumeRat = cbig.NewRatFromBigInt(takerRecord.RemainingQuantity.BigInt())
	if takerRecord.Side != makerRecord.Side { // self
		makerVolumeRat = cbig.NewRatFromBigInt(makerRecord.RemainingQuantity.BigInt())
	} else { // opposite
		makerVolumeRat = cbig.RatMul(cbig.NewRatFromBigInt(makerRecord.RemainingQuantity.BigInt()), makerRecord.Price.Rat())
	}
	k.logger(ctx).Debug("Computed order volumes.", "takerVolume", takerVolumeRat, "makerVolume", makerVolumeRat)

	if cbig.RatGTE(takerVolumeRat, makerVolumeRat) {
		// close maker record
		recordToClose = makerRecord
		recordToReduce = takerRecord
	} else {
		// close taker record
		recordToClose = takerRecord
		recordToReduce = makerRecord
	}

	return recordToClose, recordToReduce
}

func (k Keeper) lockRequiredBalances(
	ctx sdk.Context, params types.Params, order types.Order, takerRecord types.OrderBookRecord,
) (types.OrderBookRecord, error) {
	creatorAddr, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return types.OrderBookRecord{}, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}
	// recompute the min required balance to be locked based on record remaining quantity
	lockCoin, err := types.ComputeLimitOrderLockedBalance(
		order.Side, order.BaseDenom, order.QuoteDenom, takerRecord.RemainingQuantity, *order.Price,
	)
	if err != nil {
		return types.OrderBookRecord{}, err
	}
	expectedToReceiveCoin, err := types.ComputeLimitOrderExpectedToReceiveBalance(
		order.Side, order.BaseDenom, order.QuoteDenom, takerRecord.RemainingQuantity, *order.Price,
	)
	if err != nil {
		return types.OrderBookRecord{}, err
	}

	// lock reserve if positive
	if params.OrderReserve.IsPositive() {
		// don't check for the DEX FT limits since it's independent of the trading limits
		if err := k.lockFT(ctx, creatorAddr, params.OrderReserve); err != nil {
			return types.OrderBookRecord{},
				sdkerrors.Wrapf(err, "failed to lock order reserve: %s", params.OrderReserve.String())
		}
	}

	if err := k.increaseFTLimits(
		ctx,
		creatorAddr,
		lockCoin,
		expectedToReceiveCoin,
	); err != nil {
		return types.OrderBookRecord{}, err
	}
	takerRecord.RemainingBalance = lockCoin.Amount

	return takerRecord, nil
}

func getRecordsReceiveCoins(
	makerRecord, recordToClose, recordToReduce *types.OrderBookRecord,
	order types.Order,
) (sdk.Coin, sdk.Coin, sdkmath.Int, sdkmath.Int) {
	var (
		recordToCloseReceiveDenom     string
		recordToCloseReceiveAmt       sdkmath.Int
		recordToReduceReceiveDenom    string
		recordToReduceReceiveAmt      sdkmath.Int
		recordToCloseReducedQuantity  sdkmath.Int
		recordToReduceReducedQuantity sdkmath.Int

		executionQuantity         sdkmath.Int
		oppositeExecutionQuantity sdkmath.Int
	)

	if recordToClose.Side != recordToReduce.Side { // self
		executionQuantity, oppositeExecutionQuantity = computeMaxExecutionQuantity(
			makerRecord.Price.Rat(), recordToClose.RemainingQuantity,
		)
		recordToCloseReducedQuantity = executionQuantity
		recordToReduceReducedQuantity = executionQuantity
	} else {
		// if closeMaker is true we find max execution quantity with its price,
		// else with inverse price
		if recordToClose.IsMaker() {
			executionQuantity, oppositeExecutionQuantity = computeMaxExecutionQuantity(
				makerRecord.Price.Rat(), recordToClose.RemainingQuantity,
			)
		} else {
			executionQuantity, oppositeExecutionQuantity = computeMaxExecutionQuantity(
				cbig.RatInv(makerRecord.Price.Rat()), recordToClose.RemainingQuantity,
			)
		}

		recordToCloseReducedQuantity = executionQuantity
		recordToReduceReducedQuantity = oppositeExecutionQuantity
	}

	if recordToClose.Side == types.SIDE_BUY {
		recordToCloseReceiveAmt = executionQuantity
		recordToReduceReceiveAmt = oppositeExecutionQuantity
	} else {
		recordToCloseReceiveAmt = oppositeExecutionQuantity
		recordToReduceReceiveAmt = executionQuantity
	}

	if recordToClose.IsMaker() {
		recordToCloseReceiveDenom = order.GetSpendDenom()
		recordToReduceReceiveDenom = order.GetReceiveDenom()
	} else {
		recordToCloseReceiveDenom = order.GetReceiveDenom()
		recordToReduceReceiveDenom = order.GetSpendDenom()
	}

	recordToCloseReceiveCoin := sdk.NewCoin(recordToCloseReceiveDenom, recordToCloseReceiveAmt)
	recordToReduceReceiveCoin := sdk.NewCoin(recordToReduceReceiveDenom, recordToReduceReceiveAmt)

	return recordToCloseReceiveCoin, recordToReduceReceiveCoin, recordToCloseReducedQuantity, recordToReduceReducedQuantity
}

func computeMaxExecutionQuantity(priceRat *big.Rat, remainingQuantity sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()
	n := cbig.IntQuo(remainingQuantity.BigInt(), priceDenom)
	maxExecutionQuantity := cbig.IntMul(n, priceDenom)
	oppositeExecutionQuantity := cbig.IntMul(n, priceNum)

	return sdkmath.NewIntFromBigInt(maxExecutionQuantity),
		sdkmath.NewIntFromBigInt(oppositeExecutionQuantity)
}
