package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/samber/lo"

	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// RecordToAddress is acc address mapped to record.
type RecordToAddress struct {
	Address sdk.AccAddress
	Record  *types.OrderBookRecord
}

// MatchingResult is the result of a matching operation.
type MatchingResult struct {
	TakerAddress            sdk.AccAddress
	FTActions               assetfttypes.DEXActions
	TakerOrderReducedEvent  types.EventOrderReduced
	MakerOrderReducedEvents []types.EventOrderReduced

	RecordsToRemove []RecordToAddress
	RecordToUpdate  *types.OrderBookRecord
}

// NewMatchingResult creates a new MatchingResult.
func NewMatchingResult(order types.Order) (*MatchingResult, error) {
	takerAddress, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	var orderStrPrice *string
	if order.Price != nil {
		orderStrPrice = lo.ToPtr(order.Price.String())
	}

	return &MatchingResult{
		TakerAddress: takerAddress,
		FTActions: assetfttypes.NewDEXActions(
			assetfttypes.DEXOrder{
				Creator:    takerAddress,
				Type:       order.Type.String(),
				ID:         order.ID,
				Sequence:   order.Sequence,
				BaseDenom:  order.BaseDenom,
				QuoteDenom: order.QuoteDenom,
				Price:      orderStrPrice,
				Quantity:   order.Quantity,
				Side:       order.Side.String(),
			},
		),
		TakerOrderReducedEvent: types.EventOrderReduced{
			Creator:      order.Creator,
			ID:           order.ID,
			Sequence:     order.Sequence,
			SentCoin:     sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt()),
			ReceivedCoin: sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt()),
		},
		MakerOrderReducedEvents: make([]types.EventOrderReduced, 0),
		RecordsToRemove:         make([]RecordToAddress, 0),
		RecordToUpdate:          nil,
	}, nil
}

// TakerSend registers the coin to send from taker to maker.
func (mr *MatchingResult) TakerSend(
	makerAddr sdk.AccAddress, makerOrderID string, makerOrderSequence uint64, coin sdk.Coin,
) {
	if coin.IsZero() {
		return
	}

	mr.FTActions.AddCreatorExpectedToSpend(coin)
	mr.FTActions.AddSend(mr.TakerAddress, makerAddr, coin)
	mr.FTActions.AddDecreaseExpectedToReceive(makerAddr, coin)

	mr.updateTakerSendEvents(makerAddr, makerOrderID, makerOrderSequence, coin)
}

// MakerSend registers the coin to send from maker to taker.
func (mr *MatchingResult) MakerSend(makerAddr sdk.AccAddress, makerOrderID string, coin sdk.Coin) {
	if coin.IsZero() {
		return
	}

	// call `AddCreatorExpectedToReceive` but don't call AddIncreaseExpectedToReceive since
	// `AddIncreaseExpectedToReceive` is used for the state after the matching, but CreatorExpectedToReceive before
	mr.FTActions.AddCreatorExpectedToReceive(coin)
	mr.FTActions.AddDecreaseLocked(makerAddr, coin)
	mr.FTActions.AddSend(makerAddr, mr.TakerAddress, coin)

	mr.updateMakerSendEvents(makerAddr, makerOrderID, coin)
}

// DecreaseMakerLimits registers the coins to unlock and decrease expected to receive.
func (mr *MatchingResult) DecreaseMakerLimits(
	makerAddr sdk.AccAddress,
	lockedCoins sdk.Coins, expectedToReceiveCoin sdk.Coin,
) {
	for _, coin := range lockedCoins {
		if coin.IsZero() {
			continue
		}
		mr.FTActions.AddDecreaseLocked(makerAddr, coin)
	}

	if !expectedToReceiveCoin.IsZero() {
		mr.FTActions.AddDecreaseExpectedToReceive(makerAddr, expectedToReceiveCoin)
	}
}

// IncreaseTakerLimitsForRecord increase required limits for the taker record.
func (mr *MatchingResult) IncreaseTakerLimitsForRecord(
	params types.Params,
	order types.Order,
	takerRecord *types.OrderBookRecord,
) error {
	// if the order is filled fully
	if takerRecord.RemainingBalance.IsZero() {
		return nil
	}

	lockedCoin, err := types.ComputeLimitOrderLockedBalance(
		order.Side, order.BaseDenom, order.QuoteDenom, takerRecord.RemainingQuantity, *order.Price,
	)
	if err != nil {
		return err
	}
	// update taker record with the remaining balance
	takerRecord.RemainingBalance = lockedCoin.Amount

	mr.FTActions.AddCreatorExpectedToSpend(lockedCoin)
	mr.FTActions.AddIncreaseLocked(mr.TakerAddress, lockedCoin)

	expectedToReceiveCoin, err := types.ComputeLimitOrderExpectedToReceiveBalance(
		order.Side, order.BaseDenom, order.QuoteDenom, takerRecord.RemainingQuantity, *order.Price,
	)
	if err != nil {
		return err
	}
	mr.FTActions.AddCreatorExpectedToReceive(expectedToReceiveCoin)
	mr.FTActions.AddIncreaseExpectedToReceive(mr.TakerAddress, expectedToReceiveCoin)

	// lock reserve if is set
	if params.OrderReserve.IsPositive() {
		// lock but don't increase expected to receive
		mr.FTActions.AddIncreaseLocked(mr.TakerAddress, params.OrderReserve)
	}

	return nil
}

// RemoveRecord registers the record to remove.
func (mr *MatchingResult) RemoveRecord(creator sdk.AccAddress, record *types.OrderBookRecord) {
	mr.RecordsToRemove = append(mr.RecordsToRemove, RecordToAddress{
		Address: creator,
		Record:  record,
	})
}

// UpdateRecord registers the record to update.
func (mr *MatchingResult) UpdateRecord(record types.OrderBookRecord) {
	mr.RecordToUpdate = &record
}

func (mr *MatchingResult) updateTakerSendEvents(
	makerAddr sdk.AccAddress,
	makerOrderID string,
	makerOrderSequence uint64,
	coin sdk.Coin,
) {
	mr.TakerOrderReducedEvent.SentCoin = mr.TakerOrderReducedEvent.SentCoin.Add(coin)
	mr.MakerOrderReducedEvents = append(mr.MakerOrderReducedEvents, types.EventOrderReduced{
		Creator:      makerAddr.String(),
		ID:           makerOrderID,
		Sequence:     makerOrderSequence,
		ReceivedCoin: coin,
	})
}

func (mr *MatchingResult) updateMakerSendEvents(
	makerAddr sdk.AccAddress,
	makerOrderID string,
	coin sdk.Coin,
) {
	mr.TakerOrderReducedEvent.ReceivedCoin = mr.TakerOrderReducedEvent.ReceivedCoin.Add(coin)
	for i := range mr.MakerOrderReducedEvents {
		// find corresponding event created by `updateTakerSendEvents`
		if mr.MakerOrderReducedEvents[i].Creator == makerAddr.String() && mr.MakerOrderReducedEvents[i].ID == makerOrderID {
			mr.MakerOrderReducedEvents[i].SentCoin = coin
			break
		}
	}
}

func (k Keeper) applyMatchingResult(ctx sdk.Context, mr *MatchingResult) error {
	// if matched passed but no changes are applied return
	if mr.FTActions.CreatorExpectedToSpend.IsNil() {
		return nil
	}

	if err := k.assetFTKeeper.DEXExecuteActions(ctx, mr.FTActions); err != nil {
		return err
	}

	for _, item := range mr.RecordsToRemove {
		if err := k.removeOrderByRecord(ctx, item.Address, *item.Record); err != nil {
			return err
		}
	}

	if mr.RecordToUpdate != nil {
		if err := k.saveOrderBookRecord(ctx, *mr.RecordToUpdate); err != nil {
			return err
		}
	}

	return k.publishMatchingEvents(ctx, mr)
}

func (k Keeper) publishMatchingEvents(
	ctx sdk.Context,
	mr *MatchingResult,
) error {
	events := mr.MakerOrderReducedEvents
	if !mr.TakerOrderReducedEvent.SentCoin.IsZero() {
		events = append(events, mr.TakerOrderReducedEvent)
	}

	for _, evt := range events {
		if err := ctx.EventManager().EmitTypedEvent(&evt); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderReduced: %s", err)
		}
	}

	return nil
}
