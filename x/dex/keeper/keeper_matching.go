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

	makerRecord, matches, err := mf.Next()
	if err != nil {
		return err
	}
	// if no match and record is maker exit, since we don't need neither to lock nor to save the record
	if !matches && order.Type == types.ORDER_TYPE_MARKET {
		return nil
	}

	takerRecord, err := k.initTakerRecord(ctx, accNumber, orderBookID, order)
	if err != nil {
		return err
	}

	accNumberToAddrCache := make(map[uint64]sdk.AccAddress)

	for {
		if !matches {
			break
		}
		var stop bool
		accNumberToAddrCache, stop, err = k.matchRecords(ctx, accNumberToAddrCache, &takerRecord, &makerRecord, order)
		if err != nil {
			return err
		}
		if stop {
			break
		}
		makerRecord, matches, err = mf.Next()
		if err != nil {
			return err
		}
	}

	if takerRecord.RemainingBalance.IsPositive() {
		switch order.Type {
		case types.ORDER_TYPE_LIMIT:
			switch order.TimeInForce {
			case types.TIME_IN_FORCE_GTC:
				// create new order with the updated record
				return k.createOrder(ctx, params, order, takerRecord)
			case types.TIME_IN_FORCE_IOC:
				// unlock the remaining balance
				return k.unlockRemainingBalance(ctx, order, takerRecord)
			default:
				return sdkerrors.Wrapf(types.ErrInvalidInput, "unsupported time in force: %s", order.TimeInForce.String())
			}
		case types.ORDER_TYPE_MARKET:
			return k.unlockRemainingBalance(ctx, order, takerRecord)
		default:
			return sdkerrors.Wrapf(
				types.ErrInvalidInput, "unexpect order type : %s", order.Type.String(),
			)
		}
	}

	return nil
}

func (k Keeper) initTakerRecord(
	ctx sdk.Context,
	accNumber uint64,
	orderBookID uint32,
	order types.Order,
) (types.OrderBookRecord, error) {
	remainingAmount, err := k.lockOrderBalance(ctx, order)
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

func (k Keeper) matchRecords(
	ctx sdk.Context,
	accNumberToAddrCache map[uint64]sdk.AccAddress,
	takerRecord, makerRecord *types.OrderBookRecord,
	order types.Order,
) (map[uint64]sdk.AccAddress, bool, error) {
	recordToClose, recordToReduce, closeMaker := k.getRecordToCloseAndReduce(ctx, takerRecord, makerRecord)
	k.logger(ctx).Debug(
		"Executing orders.",
		"recordToClose", recordToClose.String(),
		"recordToReduce", recordToReduce.String(),
	)

	recordToCloseReceiveCoin, recordToReduceReceiveCoin, recordToReduceReducedQuantity := getRecordsReceiveCoins(
		makerRecord, recordToClose, recordToReduce, order, closeMaker,
	)

	// stop if any record receives more than opposite record balance
	// that situation is possible with the marker order with quantity which doesn't correspond the order balance
	if recordToClose.RemainingBalance.LT(recordToReduceReceiveCoin.Amount) ||
		recordToReduce.RemainingBalance.LT(recordToCloseReceiveCoin.Amount) {
		k.logger(ctx).Debug("Stop matching, order balance is not enough to cover the quantity.")
		return accNumberToAddrCache, true, nil
	}

	var recordToCloseAddr, recordToReduceAddr sdk.AccAddress
	recordToCloseAddr, recordToReduceAddr, accNumberToAddrCache, err := k.getRecordToCloseAndReduceAddresses(
		ctx, recordToClose, recordToReduce, accNumberToAddrCache)
	if err != nil {
		return nil, false, err
	}

	if err := k.unlockAndSendCoin(
		ctx, recordToReduceAddr, recordToCloseAddr, recordToCloseReceiveCoin,
	); err != nil {
		return nil, false, err
	}

	if err := k.unlockCoin(
		ctx,
		recordToCloseAddr,
		sdk.NewCoin(
			recordToReduceReceiveCoin.Denom,
			recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount),
		),
	); err != nil {
		return nil, false, err
	}

	recordToClose.RemainingQuantity = recordToClose.RemainingQuantity.Sub(recordToCloseReceiveCoin.Amount)
	recordToClose.RemainingBalance = sdkmath.ZeroInt()
	k.logger(ctx).Debug("Updated recordToClose.", "recordToClose", recordToClose)

	if err := k.unlockAndSendCoin(
		ctx, recordToCloseAddr, recordToReduceAddr, recordToReduceReceiveCoin,
	); err != nil {
		return nil, false, err
	}

	recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(recordToReduceReducedQuantity)
	recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)
	k.logger(ctx).Debug("Updated recordToReduce.", "recordToReduce", recordToReduce)

	// remove order only if it's maker, so it was saved before
	if closeMaker {
		if err := k.removeOrderByRecordAndUsedDenoms(ctx, *recordToClose, order.Denoms()); err != nil {
			return nil, false, err
		}
		// check if the maker record has what to fill later, or we can cancel the remaining part now
		if recordToReduce.RemainingQuantity.IsZero() {
			k.logger(ctx).Debug("Taker record is filled fully.")
			// unlock remaining balance
			if err := k.unlockCoin(
				ctx,
				recordToReduceAddr,
				sdk.NewCoin(recordToCloseReceiveCoin.Denom, recordToReduce.RemainingBalance),
			); err != nil {
				return nil, false, err
			}
			recordToReduce.RemainingBalance = sdkmath.ZeroInt()
			// stop
			return accNumberToAddrCache, true, nil
		}
		k.logger(ctx).Debug("Going to next record in the order book.")
		// continue
		return accNumberToAddrCache, false, nil
	}
	// update maker order
	if err := k.saveOrderBookRecord(ctx, *makerRecord); err != nil {
		return nil, false, err
	}
	// stop
	return accNumberToAddrCache, true, nil
}

func (k Keeper) getRecordToCloseAndReduceAddresses(
	ctx sdk.Context,
	recordToClose, recordToReduce *types.OrderBookRecord,
	accNumberToAddr map[uint64]sdk.AccAddress,
) (sdk.AccAddress, sdk.AccAddress, map[uint64]sdk.AccAddress, error) {
	var (
		recordToCloseAddr sdk.AccAddress
		err               error
	)
	recordToCloseAddr, accNumberToAddr, err = k.getAccountAddressWithCache(
		ctx, recordToClose.AccountNumber, accNumberToAddr,
	)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, nil, err
	}

	var recordToReduceAddr sdk.AccAddress
	recordToReduceAddr, accNumberToAddr, err = k.getAccountAddressWithCache(
		ctx, recordToReduce.AccountNumber, accNumberToAddr,
	)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, nil, err
	}

	return recordToCloseAddr, recordToReduceAddr, accNumberToAddr, nil
}

func (k Keeper) getRecordToCloseAndReduce(ctx sdk.Context, takerRecord, makerRecord *types.OrderBookRecord) (
	*types.OrderBookRecord, *types.OrderBookRecord, bool,
) {
	var (
		recordToClose, recordToReduce *types.OrderBookRecord
		closeMaker                    bool

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
		closeMaker = true
	} else {
		// close taker record
		recordToClose = takerRecord
		recordToReduce = makerRecord
	}

	return recordToClose, recordToReduce, closeMaker
}

func (k Keeper) unlockRemainingBalance(ctx sdk.Context, order types.Order, takerRecord types.OrderBookRecord) error {
	creatorAddr, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	return k.unlockCoin(
		ctx,
		creatorAddr,
		sdk.NewCoin(order.GetSpendDenom(), takerRecord.RemainingBalance),
	)
}

func getRecordsReceiveCoins(
	makerRecord, recordToClose, recordToReduce *types.OrderBookRecord,
	order types.Order,
	closeMaker bool,
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
		if closeMaker {
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

	if closeMaker { // recordToClose is maker
		recordToCloseReceiveDenom = order.GetSpendDenom()
		recordToReduceReceiveDenom = order.GetReceiveDenom()
	} else { // recordToClose is taker
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
