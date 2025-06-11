package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"

	cbig "github.com/CoreumFoundation/coreum/v6/pkg/math/big"
	matchingengine "github.com/CoreumFoundation/coreum/v6/x/dex/matching-engine"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

//nolint:funlen
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

	remainingBalance, err := k.getInitialRemainingBalance(ctx, takerOrder)
	if err != nil {
		return err
	}

	orderSequence, err := k.genNextOrderSequence(ctx)
	if err != nil {
		return err
	}
	takerOrder.Sequence = orderSequence

	cachedAccKeeper := newCachedAccountKeeper(k.accountKeeper, k.accountQueryServer)
	engine := matchingengine.NewMatchingEngine(mf, cachedAccKeeper, k.logger(ctx), k)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderPlaced{
		Creator:  takerOrder.Creator,
		ID:       takerOrder.ID,
		Sequence: orderSequence,
	}); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrIO, "failed to emit event EventOrderPlaced: %s", err)
	}

	mr, err := engine.MatchOrder(ctx, accNumber, orderBookID, takerOrder, remainingBalance)
	if err != nil {
		return err
	}

	switch takerOrder.Type {
	case types.ORDER_TYPE_LIMIT:
		switch takerOrder.TimeInForce {
		case types.TIME_IN_FORCE_GTC:
			// If taker order is filled fully or not executable as maker we just apply matching result and return.
			if mr.TakerIsFilled || !isOrderRecordExecutableAsMaker(&mr.TakerRecord) {
				return k.applyMatchingResult(ctx, mr)
			}

			// If taker orders is not filled fully we need to:
			// - increase taker limits for record for remaining amount
			// - apply matching result
			// - add remaining order to the order book
			if err := mr.IncreaseTakerLimitsForRecord(params, takerOrder, &mr.TakerRecord); err != nil {
				return err
			}

			// In partial match case, we should create an order for the remaining part, and it makes sense to happen
			// after finalizing the match, but since a call to smart contract happens in applyMatchingResult, it will be
			// created before that. (The reason is explained inside DEXExecuteActions function) So, smart contract will
			// see the match already happened, and it can calculate what the state was before the match, with the data
			// passed to it.
			if err := k.createOrder(ctx, params, takerOrder, mr.TakerRecord); err != nil {
				return err
			}

			return k.applyMatchingResult(ctx, mr)
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr)
		case types.TIME_IN_FORCE_FOK:
			// ensure full order fill
			if mr.TakerRecord.RemainingBaseQuantity.IsPositive() {
				return nil
			}
			return k.applyMatchingResult(ctx, mr)
		default:
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"unsupported time in force: %s for limit order",
				takerOrder.TimeInForce.String())
		}
	case types.ORDER_TYPE_MARKET:
		switch takerOrder.TimeInForce {
		case types.TIME_IN_FORCE_IOC:
			return k.applyMatchingResult(ctx, mr)
		default:
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"unsupported time in force: %s for market order",
				takerOrder.TimeInForce.String())
		}
	default:
		return sdkerrors.Wrapf(
			types.ErrInvalidInput, "unexpected order type: %s", takerOrder.Type.String(),
		)
	}
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
