package keeper

import (
	"fmt"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// MatchingFinder is responsible for finding orders with the best price and priority.
type MatchingFinder struct {
	log log.Logger

	directOBIterator   *OrderBookIterator
	invertedOBIterator *OrderBookIterator

	order types.Order

	directOBRecord   *types.OrderBookRecord
	invertedOBRecord *types.OrderBookRecord
}

// NewMatchingFinder returns new instance of the MatchingFinder.
func (k Keeper) NewMatchingFinder(
	ctx sdk.Context,
	orderBookID, invertedOrderBookID uint32,
	order types.Order,
) (*MatchingFinder, error) {
	oppositeSide, err := order.Side.Opposite()
	if err != nil {
		return nil, err
	}

	return &MatchingFinder{
		log: k.logger(ctx),

		directOBIterator:   k.NewOrderBookSideIterator(ctx, orderBookID, oppositeSide),
		invertedOBIterator: k.NewOrderBookSideIterator(ctx, invertedOrderBookID, order.Side),

		order: order,
	}, nil
}

// Next returns the next order book record with the best price and priority and flag that indicates whether it matches
// the taker record.
func (mf *MatchingFinder) Next() (types.OrderBookRecord, bool, error) {
	if err := mf.loadOrders(); err != nil {
		return types.OrderBookRecord{}, false, err
	}

	var directMatches, invertedMatches bool
	switch mf.order.Type {
	case types.ORDER_TYPE_LIMIT:
		directMatches = mf.isDirectRecordMatches()
		invertedMatches = mf.isInvertedRecordMatches()
	case types.ORDER_TYPE_MARKET:
		if mf.directOBRecord != nil {
			directMatches = true
		}
		if mf.invertedOBRecord != nil {
			invertedMatches = true
		}
	default:
		return types.OrderBookRecord{}, false, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type : %s", mf.order.Type.String(),
		)
	}

	// no match
	if !directMatches && !invertedMatches {
		mf.log.Debug("Both maker records don't match taker record.")
		return types.OrderBookRecord{}, false, nil
	}
	// matches
	if mf.isDirectRecordBestMatch(directMatches, invertedMatches) {
		mf.log.Debug("Direct OB record is best match.")
		record := *mf.directOBRecord
		mf.directOBRecord = nil
		return record, true, nil
	}

	mf.log.Debug("Inverted OB record is best match.")
	record := *mf.invertedOBRecord
	mf.invertedOBRecord = nil
	return record, true, nil
}

// Close closes used iterators for the MatchingFinder.
func (mf *MatchingFinder) Close() error {
	if err := mf.directOBIterator.Close(); err != nil {
		return sdkerrors.Wrapf(err, "failed to close directOBIterator")
	}
	if err := mf.invertedOBIterator.Close(); err != nil {
		return sdkerrors.Wrapf(err, "failed to close invertedOBIterator")
	}

	return nil
}

func (mf *MatchingFinder) loadOrders() error {
	if mf.directOBRecord == nil {
		directOBRecord, found, err := mf.directOBIterator.Next()
		if err != nil {
			return err
		}
		if found {
			mf.directOBRecord = &directOBRecord
		}
	}

	if mf.invertedOBRecord == nil {
		invertedOBRecord, found, err := mf.invertedOBIterator.Next()
		if err != nil {
			return err
		}
		if found {
			mf.invertedOBRecord = &invertedOBRecord
		}
	}

	return nil
}

func (mf *MatchingFinder) isDirectRecordBestMatch(directMatches, invertedMatches bool) bool {
	if directMatches && !invertedMatches {
		return true
	}
	if !directMatches && invertedMatches {
		return false
	}
	// both matches, find best
	directOBPriceRat := mf.directOBRecord.Price.Rat()
	invertedOBInvPriceRat := cbig.RatInv(mf.invertedOBRecord.Price.Rat())

	// if both prices are the same then FIFO by OrderSequence wins
	if directOBPriceRat.Cmp(invertedOBInvPriceRat) == 0 {
		return mf.directOBRecord.OrderSequence < mf.invertedOBRecord.OrderSequence
	}

	if mf.order.Side == types.SIDE_BUY {
		// find best sell - lower wins
		return cbig.RatGTE(invertedOBInvPriceRat, directOBPriceRat)
	}

	// find best buy - greater wins
	return cbig.RatGTE(directOBPriceRat, invertedOBInvPriceRat)
}

func (mf *MatchingFinder) isDirectRecordMatches() bool {
	if mf.directOBRecord == nil {
		mf.log.Debug("Direct order book is finished.")
		return false
	}
	matches := mf.isPriceMatches(mf.directOBRecord.Price.Rat())
	mf.log.Debug(
		"Compared direct OB maker order",
		"matches", matches,
		"takerPrice", mf.order.Price,
		"takerSide", mf.order.Side,
		"makerOrderID", mf.directOBRecord.OrderID,
		"makerPrice", mf.directOBRecord.Price,
	)

	return matches
}

func (mf *MatchingFinder) isInvertedRecordMatches() bool {
	if mf.invertedOBRecord == nil {
		mf.log.Debug("Inverted order book is finished.")
		return false
	}
	// use inversed price for the inverted OB
	matches := mf.isPriceMatches(cbig.RatInv(mf.invertedOBRecord.Price.Rat()))
	mf.log.Debug(
		"Compared inverted OB maker order",
		"matches", matches,
		"takerPrice", mf.order.Price,
		"takerSide", mf.order.Side,
		"makerOrderID", mf.invertedOBRecord.OrderID,
		"makerInvPrice", fmt.Sprintf("1/%s", mf.invertedOBRecord.Price),
	)

	return matches
}

func (mf *MatchingFinder) isPriceMatches(priceRat *big.Rat) bool {
	if mf.order.Side == types.SIDE_BUY {
		return cbig.RatGTE(mf.order.Price.Rat(), priceRat)
	}

	return cbig.RatLTE(mf.order.Price.Rat(), priceRat)
}
