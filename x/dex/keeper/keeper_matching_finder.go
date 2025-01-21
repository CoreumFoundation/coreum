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

	selfIterator     *OrderBookIterator
	oppositeIterator *OrderBookIterator

	order types.Order

	selfRecord     *types.OrderBookRecord
	oppositeRecord *types.OrderBookRecord
}

// NewMatchingFinder returns new instance of the MatchingFinder.
func (k Keeper) NewMatchingFinder(
	ctx sdk.Context,
	orderBookID, oppositeOrderBookID uint32,
	order types.Order,
) (*MatchingFinder, error) {
	oppositeSide, err := order.Side.Opposite()
	if err != nil {
		return nil, err
	}

	return &MatchingFinder{
		log: k.logger(ctx),

		selfIterator:     k.NewOrderBookSideIterator(ctx, orderBookID, oppositeSide),
		oppositeIterator: k.NewOrderBookSideIterator(ctx, oppositeOrderBookID, order.Side),

		order: order,
	}, nil
}

// Next returns the next order book record with the best price and priority and flag that indicates whether it matches
// the taker record.
func (mf *MatchingFinder) Next() (types.OrderBookRecord, bool, error) {
	if err := mf.loadOrders(); err != nil {
		return types.OrderBookRecord{}, false, err
	}

	var selfMatches, oppositeMatches bool
	switch mf.order.Type {
	case types.ORDER_TYPE_LIMIT:
		selfMatches = mf.isSelfRecordMatches()
		oppositeMatches = mf.isOppositeRecordMatches()
	case types.ORDER_TYPE_MARKET:
		if mf.selfRecord != nil {
			selfMatches = true
		}
		if mf.oppositeRecord != nil {
			oppositeMatches = true
		}
	default:
		return types.OrderBookRecord{}, false, sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type : %s", mf.order.Type.String(),
		)
	}

	// no match
	if !selfMatches && !oppositeMatches {
		mf.log.Debug("Both maker records don't match taker record.")
		return types.OrderBookRecord{}, false, nil
	}
	// matches
	if mf.isSelfRecordBestMatch(selfMatches, oppositeMatches) {
		mf.log.Debug("Self record is best match.")
		record := *mf.selfRecord
		mf.selfRecord = nil
		return record, true, nil
	}

	mf.log.Debug("Opposite record is best match.")
	record := *mf.oppositeRecord
	mf.oppositeRecord = nil
	return record, true, nil
}

// Close closes used iterators for the MatchingFinder.
func (mf *MatchingFinder) Close() error {
	if err := mf.selfIterator.Close(); err != nil {
		return sdkerrors.Wrapf(err, "failed to close selfIterator")
	}
	if err := mf.oppositeIterator.Close(); err != nil {
		return sdkerrors.Wrapf(err, "failed to close oppositeIterator")
	}

	return nil
}

func (mf *MatchingFinder) loadOrders() error {
	if mf.selfRecord == nil {
		selfRecord, found, err := mf.selfIterator.Next()
		if err != nil {
			return err
		}
		if found {
			mf.selfRecord = &selfRecord
		}
	}

	if mf.oppositeRecord == nil {
		oppositeRecord, found, err := mf.oppositeIterator.Next()
		if err != nil {
			return err
		}
		if found {
			mf.oppositeRecord = &oppositeRecord
		}
	}

	return nil
}

func (mf *MatchingFinder) isSelfRecordBestMatch(selfMatches, oppositeMatches bool) bool {
	if selfMatches && !oppositeMatches {
		return true
	}
	if !selfMatches && oppositeMatches {
		return false
	}
	// both matches, find best
	selfPriceRat := mf.selfRecord.Price.Rat()
	oppositeInvPriceRat := cbig.RatInv(mf.oppositeRecord.Price.Rat())

	// the same price is impossible because in that case the opposite order would have been matched

	if mf.order.Side == types.SIDE_BUY {
		// find best sell - lower wins
		return cbig.RatGTE(oppositeInvPriceRat, selfPriceRat)
	}

	// find best buy - greater wins
	return cbig.RatGTE(selfPriceRat, oppositeInvPriceRat)
}

func (mf *MatchingFinder) isSelfRecordMatches() bool {
	if mf.selfRecord == nil {
		mf.log.Debug("Self order book is finished.")
		return false
	}
	matches := mf.isPriceMatches(mf.selfRecord.Price.Rat())
	mf.log.Debug(
		"Compared self maker order",
		"matches", matches,
		"takerPrice", mf.order.Price,
		"takerSide", mf.order.Side,
		"makerOrderID", mf.selfRecord.OrderID,
		"makerPrice", mf.selfRecord.Price,
	)

	return matches
}

func (mf *MatchingFinder) isOppositeRecordMatches() bool {
	if mf.oppositeRecord == nil {
		mf.log.Debug("Opposite order book is finished.")
		return false
	}
	// use inverse price for the opposite
	matches := mf.isPriceMatches(cbig.RatInv(mf.oppositeRecord.Price.Rat()))
	mf.log.Debug(
		"Compared opposite maker order",
		"matches", matches,
		"takerPrice", mf.order.Price,
		"takerSide", mf.order.Side,
		"makerOrderID", mf.oppositeRecord.OrderID,
		"makerInvPrice", fmt.Sprintf("1/%s", mf.oppositeRecord.Price),
	)

	return matches
}

func (mf *MatchingFinder) isPriceMatches(priceRat *big.Rat) bool {
	if mf.order.Side == types.SIDE_BUY {
		return cbig.RatGTE(mf.order.Price.Rat(), priceRat)
	}

	return cbig.RatLTE(mf.order.Price.Rat(), priceRat)
}
