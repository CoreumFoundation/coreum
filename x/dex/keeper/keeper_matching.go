package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
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

	mr := NewMatchingResult(accNumber)
	for {
		makerRecord, matches, err := mf.Next()
		if err != nil {
			return err
		}
		if !matches {
			break
		}
		if k.matchRecords(ctx, mr, &takerRecord, &makerRecord, order) {
			break
		}
	}

	switch order.Type {
	case types.ORDER_TYPE_LIMIT:
		switch order.TimeInForce {
		case types.TIME_IN_FORCE_GTC:
			// create new order with the updated record
			if err := k.applyMatchingResult(ctx, mr, order); err != nil {
				return err
			}
			if takerRecord.RemainingBalance.IsPositive() {
				takerAddr, err := sdk.AccAddressFromBech32(order.Creator)
				if err != nil {
					return sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
				}
				// lock the remaining balance
				coinToLock := sdk.NewCoin(order.GetSpendDenom(), takerRecord.RemainingBalance)
				if err := k.lockCoin(ctx, takerAddr, coinToLock, order.GetReceiveDenom()); err != nil {
					return err
				}
				return k.createOrder(ctx, params, order, takerRecord)
			}
			return nil
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr, order)
		default:
			return sdkerrors.Wrapf(types.ErrInvalidInput, "unsupported time in force: %s", order.TimeInForce.String())
		}
	case types.ORDER_TYPE_MARKET:
		return k.applyMatchingResult(ctx, mr, order)
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
) bool {
	recordToClose, recordToReduce := k.getRecordToCloseAndReduce(ctx, takerRecord, makerRecord)
	k.logger(ctx).Debug(
		"Executing orders.",
		"recordToClose", recordToClose.String(),
		"recordToReduce", recordToReduce.String(),
	)

	recordToCloseReceiveCoin, recordToReduceReceiveCoin, recordToReduceReducedQuantity := getRecordsReceiveCoins(
		makerRecord, recordToClose, recordToReduce, order,
	)

	// stop if any record receives more than opposite record balance
	// that situation is possible with the marker order with quantity which doesn't correspond the order balance
	if recordToClose.RemainingBalance.LT(recordToReduceReceiveCoin.Amount) ||
		recordToReduce.RemainingBalance.LT(recordToCloseReceiveCoin.Amount) {
		k.logger(ctx).Debug("Stop matching, order balance is not enough to cover the quantity.")
		return true
	}

	if recordToReduce.IsMaker() {
		mr.RegisterMakerUnlockAndSend(recordToReduce.AccountNumber, recordToCloseReceiveCoin)
	} else {
		mr.RegisterTakerSend(recordToClose.AccountNumber, recordToCloseReceiveCoin)
	}

	if recordToClose.IsMaker() {
		mr.RegisterMakerUnlock(recordToClose.AccountNumber, sdk.NewCoin(
			recordToReduceReceiveCoin.Denom,
			recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount),
		))
	}

	recordToClose.RemainingQuantity = recordToClose.RemainingQuantity.Sub(recordToCloseReceiveCoin.Amount)
	recordToClose.RemainingBalance = sdkmath.ZeroInt()
	k.logger(ctx).Debug("Updated recordToClose.", "recordToClose", recordToClose)

	if recordToClose.IsMaker() {
		mr.RegisterMakerUnlockAndSend(recordToClose.AccountNumber, recordToReduceReceiveCoin)
	} else {
		mr.RegisterTakerSend(recordToReduce.AccountNumber, recordToReduceReceiveCoin)
	}

	recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(recordToReduceReducedQuantity)
	recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)
	k.logger(ctx).Debug("Updated recordToReduce.", "recordToReduce", recordToReduce)

	// remove order only if it's maker, so it was saved before
	if recordToClose.IsMaker() {
		mr.RegisterMakerRemoveRecord(*recordToClose)
		// check if the maker record has what to fill later, or we can cancel the remaining part now
		if recordToReduce.RemainingQuantity.IsZero() {
			k.logger(ctx).Debug("Taker record is filled fully.")
			recordToReduce.RemainingBalance = sdkmath.ZeroInt()
			return true
		}
		k.logger(ctx).Debug("Going to next record in the order book.")
		return false
	}

	mr.RegisterMakerUpdateRecord(*makerRecord)
	return true
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
) (sdk.Coin, sdk.Coin, sdkmath.Int) {
	var (
		recordToCloseReceiveDenom     string
		recordToCloseReceiveAmt       sdkmath.Int
		recordToReduceReceiveDenom    string
		recordToReduceReceiveAmt      sdkmath.Int
		recordToReduceReducedQuantity sdkmath.Int

		executionQuantity         sdkmath.Int
		oppositeExecutionQuantity sdkmath.Int
	)

	if recordToClose.Side != recordToReduce.Side { // self
		executionQuantity, oppositeExecutionQuantity = computeMaxExecutionQuantity(
			makerRecord.Price.Rat(), recordToClose.RemainingQuantity,
		)
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

	return recordToCloseReceiveCoin, recordToReduceReceiveCoin, recordToReduceReducedQuantity
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
