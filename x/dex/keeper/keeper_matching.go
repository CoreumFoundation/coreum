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
	orderBookID, invertedOrderBookID uint32,
	order types.Order,
) error {
	k.logger(ctx).Debug("Matching order.", "order", order.String())

	mf, err := k.NewMatchingFinder(ctx, orderBookID, invertedOrderBookID, order)
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
	order.Sequence = takerRecord.OrderSequence

	if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderPlaced{
		Creator:  order.Creator,
		ID:       order.ID,
		Sequence: takerRecord.OrderSequence,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderPlaced: %s", err)
	}

	mr, err := NewMatchingResult(order)
	if err != nil {
		return err
	}
	accNumberToAddrCache := make(map[uint64]sdk.AccAddress)
	for {
		makerRecord, matches, err := mf.Next()
		if err != nil {
			return err
		}
		if !matches {
			break
		}
		stop, err := k.matchRecords(ctx, mr, &takerRecord, &makerRecord, order, accNumberToAddrCache)
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
			if err := mr.IncreaseTakerLimitsForRecord(params, order, &takerRecord); err != nil {
				return err
			}
			// apply matching result and create new order if necessary
			if err := k.applyMatchingResult(ctx, mr); err != nil {
				return err
			}
			if takerRecord.RemainingBalance.IsZero() {
				return nil
			}

			return k.createOrder(ctx, params, order, takerRecord)
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr)
		case types.TIME_IN_FORCE_FOK:
			// ensure full order fill
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
			types.ErrInvalidInput, "unexpected order type: %s", order.Type.String(),
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
			remainingBalance, err = k.assetFTKeeper.GetSpendableBalance(ctx, creatorAddr, order.QuoteDenom)
			if err != nil {
				return sdkmath.Int{}, err
			}
		} else {
			remainingBalance = sdk.NewCoin(order.BaseDenom, order.Quantity)
		}
	default:
		return sdkmath.Int{}, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type : %s", order.Type.String(),
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
	accNumberToAddrCache map[uint64]sdk.AccAddress,
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
	closeMaker := order.Sequence != recordToClose.OrderSequence
	if closeMaker {
		makerAddr, err := k.getAccountAddressWithCache(ctx, recordToClose.AccountNumber, accNumberToAddrCache)
		if err != nil {
			return false, err
		}
		mr.SendFromTaker(
			makerAddr, recordToClose.OrderID, recordToClose.OrderSequence, recordToCloseReceiveCoin,
		)
		mr.SendFromMaker(
			makerAddr, recordToClose.OrderID, recordToReduceReceiveCoin,
		)

		lockedCoins, expectedToReceiveCoin, err := k.getMakerLockedAndExpectedToReceiveCoins(
			ctx, recordToReduceReceiveCoin, recordToClose, recordToCloseRemainingQuantity, recordToCloseReceiveCoin,
		)
		if err != nil {
			return false, err
		}

		mr.DecreaseMakerLimits(makerAddr, lockedCoins, expectedToReceiveCoin)
		mr.RemoveRecord(makerAddr, recordToClose)
	} else {
		makerAddr, err := k.getAccountAddressWithCache(ctx, recordToReduce.AccountNumber, accNumberToAddrCache)
		if err != nil {
			return false, err
		}
		mr.SendFromTaker(
			makerAddr, recordToReduce.OrderID, recordToReduce.OrderSequence, recordToReduceReceiveCoin,
		)
		mr.SendFromMaker(
			makerAddr, recordToReduce.OrderID, recordToCloseReceiveCoin,
		)
	}

	recordToClose.RemainingQuantity = recordToCloseRemainingQuantity
	recordToClose.RemainingBalance = sdkmath.ZeroInt()
	k.logger(ctx).Debug("Updated recordToClose.", "recordToClose", recordToClose)

	recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(recordToReduceReducedQuantity)
	recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)
	k.logger(ctx).Debug("Updated recordToReduce.", "recordToReduce", recordToReduce)

	// continue or stop
	if closeMaker {
		if recordToReduce.RemainingQuantity.IsZero() {
			k.logger(ctx).Debug("Taker record is filled fully.")
			recordToReduce.RemainingBalance = sdkmath.ZeroInt()
			return true, nil
		}
		k.logger(ctx).Debug("Going to next record in the order book.")
		return false, nil
	}

	mr.UpdateRecord(*makerRecord)
	return true, nil
}

func (k Keeper) getMakerLockedAndExpectedToReceiveCoins(
	ctx sdk.Context,
	recordToReduceReceiveCoin sdk.Coin,
	recordToClose *types.OrderBookRecord,
	recordToCloseRemainingQuantity sdkmath.Int,
	recordToCloseReceiveCoin sdk.Coin,
) (sdk.Coins, sdk.Coin, error) {
	lockedCoins := sdk.NewCoins(sdk.NewCoin(
		recordToReduceReceiveCoin.Denom, recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount),
	))
	// get the record data to unlock the reserve if present
	recordToCloseOrderData, err := k.getOrderData(ctx, recordToClose.OrderSequence)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	if recordToCloseOrderData.Reserve.IsPositive() {
		lockedCoins = lockedCoins.Add(recordToCloseOrderData.Reserve)
	}

	expectedToReceiveAmt, err := types.ComputeLimitOrderExpectedToReceiveAmount(
		recordToClose.Side, recordToCloseRemainingQuantity, recordToClose.Price,
	)
	if err != nil {
		return nil, sdk.Coin{}, err
	}
	expectedToReceiveCoin := sdk.NewCoin(recordToCloseReceiveCoin.Denom, expectedToReceiveAmt)

	return lockedCoins, expectedToReceiveCoin, nil
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

	closeMaker := order.Sequence != recordToClose.OrderSequence
	if recordToClose.Side != recordToReduce.Side { // self
		executionQuantity, oppositeExecutionQuantity = computeMaxExecutionQuantity(
			makerRecord.Price.Rat(), recordToClose.RemainingQuantity,
		)
		recordToCloseReducedQuantity = executionQuantity
		recordToReduceReducedQuantity = executionQuantity
	} else {
		// if closeMaker is true we find max execution quantity with its price,
		// else with inverse price
		if closeMaker {
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

	if closeMaker {
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
