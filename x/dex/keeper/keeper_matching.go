package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

//nolint:funlen // reducing the function length will lead to the worse readability
func (k Keeper) matchOrder(ctx sdk.Context, accNumber uint64, orderBookID uint32, order types.Order) error {
	k.logger(ctx).Debug("Matching order.", "order", order.String())
	oppositeSide, err := order.Side.Opposite()
	if err != nil {
		return err
	}

	lockedBalance, err := k.lockOrderBalance(ctx, order)
	if err != nil {
		return err
	}

	takerRecord := types.OrderBookRecord{
		OrderBookID:       orderBookID,
		Side:              order.Side,
		Price:             order.Price,
		OrderSeq:          0, // set to zero and update only if we need to save it to the state
		OrderID:           order.ID,
		AccountNumber:     accNumber,
		RemainingQuantity: order.Quantity,
		RemainingBalance:  lockedBalance.Amount,
	}

	oppositeSideOrderBookIterator := k.NewOrderBookSideIterator(ctx, orderBookID, oppositeSide)
	defer oppositeSideOrderBookIterator.Close()

	for {
		oppositeSideRecord, exist, err := oppositeSideOrderBookIterator.Next()
		if err != nil {
			return err
		}
		// if nothing to match with, stop the execution
		if !exist {
			k.logger(ctx).Debug("Reached the end of the order book.")
			break
		}

		makerRecord := oppositeSideRecord
		k.logger(ctx).Debug(
			"Finding best match in self order book.",
			"takerRecord", takerRecord.String(),
			"makerRecord", makerRecord.String(),
		)
		// compare the price
		if !isOppositeSideRecordMatches(takerRecord, oppositeSideRecord) {
			k.logger(ctx).Debug("Taker record doesn't match maker record.")
			break
		}

		recordToClose, recordToReduce, closeMaker := getRecordToCloseAndReduce(&takerRecord, &makerRecord)
		k.logger(ctx).Debug(
			"Executing orders.",
			"recordToClose", recordToClose.String(),
			"recordToReduce", recordToReduce.String(),
		)

		// the executionQuantity is the quantity we use based on the order with the lower volume
		executionQuantity, oppositeExecutionQuantity := computeMaxExecutionQuantity(
			makerRecord.Price.Rat(), recordToClose.RemainingQuantity,
		)
		recordToCloseReceiveCoin, recordToReduceReceiveCoin := getRecordToCloseAndReceiveCoins(
			recordToClose, order, executionQuantity, oppositeExecutionQuantity,
		)
		// send coins to account of the record to close
		if err := k.sendCoinsFromDEX(
			ctx, recordToClose.AccountNumber,
			sdk.NewCoins(
				recordToCloseReceiveCoin,
				// refund the remaining balance
				sdk.NewCoin(recordToReduceReceiveCoin.Denom, recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount)),
			),
		); err != nil {
			return err
		}
		recordToClose.RemainingQuantity = recordToClose.RemainingQuantity.Sub(recordToCloseReceiveCoin.Amount)
		recordToClose.RemainingBalance = sdkmath.ZeroInt()

		if recordToClose.RemainingQuantity.IsPositive() {
			k.logger(ctx).Debug(
				"Closing with not zero remaining quantity.",
				"remainingQuantity", recordToClose.RemainingQuantity.String(),
			)
		}
		// send coins to account of the record to reduce
		if err := k.sendCoinsFromDEX(
			ctx, recordToReduce.AccountNumber, sdk.NewCoins(recordToReduceReceiveCoin),
		); err != nil {
			return err
		}
		recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(executionQuantity)
		recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)

		// remove order only if it's maker, so it was saved before
		if closeMaker {
			if err := k.removeOrderByRecord(ctx, *recordToClose); err != nil {
				return err
			}
			// check if the maker record has what to fill later, or we can cancel the remaining part now
			if recordToReduce.RemainingQuantity.IsZero() {
				k.logger(ctx).Debug("Taker record is filled fully.")
				// send remaining coins
				if err := k.sendCoinsFromDEX(ctx, recordToReduce.AccountNumber, sdk.NewCoins(
					sdk.NewCoin(recordToCloseReceiveCoin.Denom, recordToReduce.RemainingBalance),
				)); err != nil {
					return err
				}
				recordToReduce.RemainingBalance = sdkmath.ZeroInt()
				break
			}
			k.logger(ctx).Debug("Going to next record in the order book.")
			continue
		}
		// update maker order
		if err := k.saveOrderBookRecord(ctx, makerRecord); err != nil {
			return err
		}
	}
	// create new order with the updated record
	if takerRecord.RemainingBalance.IsPositive() {
		if err := k.createOrder(ctx, order, takerRecord); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) lockOrderBalance(ctx sdk.Context, order types.Order) (sdk.Coin, error) {
	lockedBalance, err := order.ComputeLockedBalance()
	if err != nil {
		return sdk.Coin{}, err
	}
	// currently we send the coins to module and from, but later we will implement the locking instead
	if err := k.sendCoinToDEX(ctx, order.Creator, lockedBalance); err != nil {
		return sdk.Coin{}, err
	}

	k.logger(ctx).Debug("Locked order balance.", "lockedBalance", lockedBalance)

	return lockedBalance, nil
}

func isOppositeSideRecordMatches(takerRecord, oppositeSideRecord types.OrderBookRecord) bool {
	if takerRecord.Side == types.Side_buy {
		return cbig.RatGTE(takerRecord.Price.Rat(), oppositeSideRecord.Price.Rat())
	}

	return cbig.RatLTE(takerRecord.Price.Rat(), oppositeSideRecord.Price.Rat())
}

func getRecordToCloseAndReduce(takerRecord, makerRecord *types.OrderBookRecord) (
	*types.OrderBookRecord, *types.OrderBookRecord, bool,
) {
	var (
		recordToClose, recordToReduce *types.OrderBookRecord
		closeMaker                    bool
	)
	// find the order with greater volume
	if takerRecord.RemainingQuantity.GTE(makerRecord.RemainingQuantity) {
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

func computeMaxExecutionQuantity(priceRat *big.Rat, remainingQuantity sdkmath.Int) (sdkmath.Int, sdkmath.Int) {
	priceNum := priceRat.Num()
	priceDenom := priceRat.Denom()
	n := cbig.IntQuo(remainingQuantity.BigInt(), priceDenom)
	maxExecutionQuantity := cbig.IntMul(n, priceDenom)
	oppositeExecutionQuantity := cbig.IntMul(n, priceNum)

	return sdkmath.NewIntFromBigInt(maxExecutionQuantity),
		sdkmath.NewIntFromBigInt(oppositeExecutionQuantity)
}

func getRecordToCloseAndReceiveCoins(
	recordToClose *types.OrderBookRecord,
	order types.Order,
	executionQuantity sdkmath.Int,
	oppositeExecutionQuantity sdkmath.Int,
) (sdk.Coin, sdk.Coin) {
	var (
		recordToCloseReceiveDenom  string
		recordToCloseReceiveAmt    sdkmath.Int
		recordToReduceReceiveDenom string
		recordToReduceReceiveAmt   sdkmath.Int
	)
	if recordToClose.Side == types.Side_buy {
		recordToCloseReceiveDenom = order.BaseDenom
		recordToCloseReceiveAmt = executionQuantity
		recordToReduceReceiveDenom = order.QuoteDenom
		recordToReduceReceiveAmt = oppositeExecutionQuantity
	} else {
		recordToCloseReceiveDenom = order.QuoteDenom
		recordToCloseReceiveAmt = oppositeExecutionQuantity
		recordToReduceReceiveDenom = order.BaseDenom
		recordToReduceReceiveAmt = executionQuantity
	}

	recordToCloseReceiveCoin := sdk.NewCoin(recordToCloseReceiveDenom, recordToCloseReceiveAmt)
	recordToReduceReceiveCoin := sdk.NewCoin(recordToReduceReceiveDenom, recordToReduceReceiveAmt)

	return recordToCloseReceiveCoin, recordToReduceReceiveCoin
}
