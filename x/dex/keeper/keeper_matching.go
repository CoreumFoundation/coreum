package keeper

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

//nolint:funlen // reducing the function length will lead to the worse readability
func (k Keeper) matchOrder(
	ctx sdk.Context,
	accNumber uint64,
	orderBookID, oppositeOrderBookID uint32,
	order types.Order,
) error {
	k.logger(ctx).Debug("Matching order.", "order", order.String())

	mf, err := k.NewMatchingFinder(ctx, orderBookID, oppositeOrderBookID, order.Side, order.Price)
	if err != nil {
		return err
	}
	defer func() {
		if err := mf.Close(); err != nil {
			k.logger(ctx).Error(err.Error())
		}
	}()

	accNumberToAddrCache := make(map[uint64]sdk.AccAddress)
	takerRecord, err := k.initTakerRecord(ctx, accNumber, orderBookID, order)
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

		recordToClose, recordToReduce, closeMaker := k.getRecordToCloseAndReduce(ctx, &takerRecord, &makerRecord)
		k.logger(ctx).Debug(
			"Executing orders.",
			"recordToClose", recordToClose.String(),
			"recordToReduce", recordToReduce.String(),
		)

		recordToCloseReceiveCoin, recordToReduceReceiveCoin, recordToReduceReducedQuantity := getRecordsReceiveCoins(
			&makerRecord, recordToClose, recordToReduce, order, closeMaker,
		)

		var recordToCloseAddr, recordToReduceAddr sdk.AccAddress
		recordToCloseAddr, recordToReduceAddr, accNumberToAddrCache, err = k.getRecordToCloseAndReduceAddresses(
			ctx, recordToClose, recordToReduce, accNumberToAddrCache)
		if err != nil {
			return err
		}

		if err := k.unlockAndSendCoin(
			ctx, recordToReduceAddr, recordToCloseAddr, recordToCloseReceiveCoin,
		); err != nil {
			return err
		}

		if err := k.unlockCoin(
			ctx,
			recordToCloseAddr,
			sdk.NewCoin(
				recordToReduceReceiveCoin.Denom,
				recordToClose.RemainingBalance.Sub(recordToReduceReceiveCoin.Amount),
			),
		); err != nil {
			return err
		}

		recordToClose.RemainingQuantity = recordToClose.RemainingQuantity.Sub(recordToCloseReceiveCoin.Amount)
		recordToClose.RemainingBalance = sdkmath.ZeroInt()
		k.logger(ctx).Debug("Updated recordToClose.", "recordToClose", recordToClose)

		if err := k.unlockAndSendCoin(
			ctx, recordToCloseAddr, recordToReduceAddr, recordToReduceReceiveCoin,
		); err != nil {
			return err
		}

		recordToReduce.RemainingQuantity = recordToReduce.RemainingQuantity.Sub(recordToReduceReducedQuantity)
		recordToReduce.RemainingBalance = recordToReduce.RemainingBalance.Sub(recordToCloseReceiveCoin.Amount)
		k.logger(ctx).Debug("Updated recordToReduce.", "recordToReduce", recordToReduce)

		// remove order only if it's maker, so it was saved before
		if closeMaker {
			if err := k.removeOrderByRecord(ctx, *recordToClose); err != nil {
				return err
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
		break
	}
	// create new order with the updated record
	if takerRecord.RemainingBalance.IsPositive() {
		if err := k.createOrder(ctx, order, takerRecord); err != nil {
			return err
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
	lockedBalance, err := k.lockOrderBalance(ctx, order)
	if err != nil {
		return types.OrderBookRecord{}, err
	}
	return types.OrderBookRecord{
		OrderBookID:       orderBookID,
		Side:              order.Side,
		Price:             order.Price,
		OrderSeq:          0, // set to zero and update only if we need to save it to the state
		OrderID:           order.ID,
		AccountNumber:     accNumber,
		RemainingQuantity: order.Quantity,
		RemainingBalance:  lockedBalance.Amount,
	}, nil
}

func (k Keeper) getRecordToCloseAndReduceAddresses(
	ctx sdk.Context,
	recordToClose, recordToReduce *types.OrderBookRecord,
	accountNumberToAddr map[uint64]sdk.AccAddress,
) (sdk.AccAddress, sdk.AccAddress, map[uint64]sdk.AccAddress, error) {
	var (
		recordToCloseAddr sdk.AccAddress
		err               error
	)
	recordToCloseAddr, accountNumberToAddr, err = k.getAccountAddressWithCache(
		ctx, recordToClose.AccountNumber, accountNumberToAddr,
	)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, nil, err
	}

	var recordToReduceAddr sdk.AccAddress
	recordToReduceAddr, accountNumberToAddr, err = k.getAccountAddressWithCache(
		ctx, recordToReduce.AccountNumber, accountNumberToAddr,
	)
	if err != nil {
		return sdk.AccAddress{}, sdk.AccAddress{}, nil, err
	}

	return recordToCloseAddr, recordToReduceAddr, accountNumberToAddr, nil
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

	if recordToClose.Side == types.Side_buy {
		recordToCloseReceiveAmt = executionQuantity
		recordToReduceReceiveAmt = oppositeExecutionQuantity
	} else {
		recordToCloseReceiveAmt = oppositeExecutionQuantity
		recordToReduceReceiveAmt = executionQuantity
	}

	if closeMaker { // recordToClose is maker
		recordToCloseReceiveDenom = order.GetBalanceDenom()
		recordToReduceReceiveDenom = order.GetOppositeFromBalanceDenom()
	} else { // recordToClose is taker
		recordToCloseReceiveDenom = order.GetOppositeFromBalanceDenom()
		recordToReduceReceiveDenom = order.GetBalanceDenom()
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
