package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// TakerCheckLimitsAndSendCoin is the taker coin to check limits send.
type TakerCheckLimitsAndSendCoin struct {
	MakerAccNumber             uint64
	MakerOrderID               string
	SendCoin                   sdk.Coin
	CheckExpectedToReceiveCoin sdk.Coin
}

// MakerUnlockAndSendCoin is the maker coin to unlock and send.
type MakerUnlockAndSendCoin struct {
	MakerAccNumber                uint64
	MakerOrderID                  string
	UnlockAndSendCoin             sdk.Coin
	DecreaseExpectedToReceiveCoin sdk.Coin
}

// MakerUnlockCoin is the maker coin to unlock.
type MakerUnlockCoin struct {
	MakerAccNumber                uint64
	UnlockCoin                    sdk.Coin
	DecreaseExpectedToReceiveCoin sdk.Coin
}

// MatchingResult is the result of a matching operation.
type MatchingResult struct {
	TakerAddress            sdk.AccAddress
	TakerOrderReducedEvent  types.EventOrderReduced
	TakerCheckLimitsAndSend []TakerCheckLimitsAndSendCoin
	MakerUnlockAndSend      []MakerUnlockAndSendCoin
	MakerUnlock             []MakerUnlockCoin
	MakerRemoveRecords      []*types.OrderBookRecord
	MakerOrderReducedEvents []types.EventOrderReduced
	MakerUpdateRecord       *types.OrderBookRecord
}

// NewMatchingResult creates a new MatchingResult.
func NewMatchingResult(order types.Order) (*MatchingResult, error) {
	takerAddress, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	return &MatchingResult{
		TakerAddress: takerAddress,
		TakerOrderReducedEvent: types.EventOrderReduced{
			Creator:      order.Creator,
			ID:           order.ID,
			SentCoin:     sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt()),
			ReceivedCoin: sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt()),
		},
		TakerCheckLimitsAndSend: make([]TakerCheckLimitsAndSendCoin, 0),
		MakerUnlockAndSend:      make([]MakerUnlockAndSendCoin, 0),
		MakerUnlock:             make([]MakerUnlockCoin, 0),
		MakerRemoveRecords:      make([]*types.OrderBookRecord, 0),
		MakerOrderReducedEvents: make([]types.EventOrderReduced, 0),
		MakerUpdateRecord:       nil,
	}, nil
}

// RegisterTakerCheckLimitsAndSendCoin sets the taker coin to check limits and send.
func (mr *MatchingResult) RegisterTakerCheckLimitsAndSendCoin(
	makerAccNumber uint64,
	makerOrderID string,
	sendCoin, checkExpectedToReceiveCoin sdk.Coin,
) {
	if sendCoin.IsZero() {
		return
	}

	mr.TakerCheckLimitsAndSend = append(mr.TakerCheckLimitsAndSend, TakerCheckLimitsAndSendCoin{
		MakerAccNumber:             makerAccNumber,
		MakerOrderID:               makerOrderID,
		SendCoin:                   sendCoin,
		CheckExpectedToReceiveCoin: checkExpectedToReceiveCoin,
	})
}

// RegisterMakerUnlockAndSend sets the maker coin to unlock and send.
func (mr *MatchingResult) RegisterMakerUnlockAndSend(
	makerAccNumber uint64,
	makerOrderID string,
	unlockAndSendCoin, decreaseExpectedToReceiveCoin sdk.Coin,
) {
	if unlockAndSendCoin.IsZero() {
		return
	}

	mr.MakerUnlockAndSend = append(mr.MakerUnlockAndSend, MakerUnlockAndSendCoin{
		MakerOrderID:                  makerOrderID,
		MakerAccNumber:                makerAccNumber,
		UnlockAndSendCoin:             unlockAndSendCoin,
		DecreaseExpectedToReceiveCoin: decreaseExpectedToReceiveCoin,
	})
}

// RegisterMakerUnlock sets the maker coin to unlock.
func (mr *MatchingResult) RegisterMakerUnlock(
	makerAccNumber uint64, unlockCoin, decreaseExpectedToReceiveCoin sdk.Coin,
) {
	if unlockCoin.IsZero() {
		return
	}

	mr.MakerUnlock = append(mr.MakerUnlock, MakerUnlockCoin{
		MakerAccNumber:                makerAccNumber,
		UnlockCoin:                    unlockCoin,
		DecreaseExpectedToReceiveCoin: decreaseExpectedToReceiveCoin,
	})
}

// RegisterMakerRemoveRecord sets the record to remove.
func (mr *MatchingResult) RegisterMakerRemoveRecord(record *types.OrderBookRecord) {
	mr.MakerRemoveRecords = append(mr.MakerRemoveRecords, record)
}

// RegisterMakerUpdateRecord sets the record to update.
func (mr *MatchingResult) RegisterMakerUpdateRecord(record types.OrderBookRecord) {
	mr.MakerUpdateRecord = &record
}

type accountToCoinsMapping struct {
	AccAddress sdk.AccAddress
	Coin1      sdk.Coin
	Coin2      sdk.Coin
}

type accountsToCoins struct {
	mapping []accountToCoinsMapping
}

func newAccountsToCoins() *accountsToCoins {
	return &accountsToCoins{
		mapping: make([]accountToCoinsMapping, 0),
	}
}

func (a *accountsToCoins) Add(acc sdk.AccAddress, coin1, coin2 sdk.Coin) {
	for i := range a.mapping {
		if a.mapping[i].AccAddress.String() == acc.String() {
			a.mapping[i].Coin1 = a.mapping[i].Coin1.Add(coin1)
			a.mapping[i].Coin2 = a.mapping[i].Coin2.Add(coin2)
			return
		}
	}
	a.mapping = append(a.mapping, accountToCoinsMapping{
		AccAddress: acc,
		Coin1:      coin1,
		Coin2:      coin2,
	})
}

func (k Keeper) applyMatchingResult(ctx sdk.Context, mr *MatchingResult) error {
	accCache := make(map[uint64]sdk.AccAddress)

	if err := k.applyMatchingResultTakerCheckLimitsAndSend(ctx, mr, accCache); err != nil {
		return err
	}

	if err := k.applyMatchingResultMakerUnlockAndSend(ctx, mr, accCache); err != nil {
		return err
	}

	if err := k.applyMatchingResultMakerUnlock(ctx, mr, accCache); err != nil {
		return err
	}

	if err := k.applyMatchingResultMakerRemoveRecords(ctx, mr, accCache); err != nil {
		return err
	}

	if mr.MakerUpdateRecord != nil {
		if err := k.saveOrderBookRecord(ctx, *mr.MakerUpdateRecord); err != nil {
			return err
		}
	}

	return k.publishMatchingEvents(ctx, mr)
}

func (k Keeper) applyMatchingResultTakerCheckLimitsAndSend(
	ctx sdk.Context,
	mr *MatchingResult,
	accCache map[uint64]sdk.AccAddress,
) error {
	accsToCoins := newAccountsToCoins()
	for _, item := range mr.TakerCheckLimitsAndSend {
		makerAddr, err := k.getAccountAddressWithCache(ctx, item.MakerAccNumber, accCache)
		if err != nil {
			return err
		}
		accsToCoins.Add(makerAddr, item.SendCoin, item.CheckExpectedToReceiveCoin)

		// init event
		mr.MakerOrderReducedEvents = append(mr.MakerOrderReducedEvents, types.EventOrderReduced{
			Creator:      makerAddr.String(),
			ID:           item.MakerOrderID,
			ReceivedCoin: item.SendCoin,
		})
		mr.TakerOrderReducedEvent.SentCoin = mr.TakerOrderReducedEvent.SentCoin.Add(item.SendCoin)
	}
	for _, accToCoins := range accsToCoins.mapping {
		if err := k.checkFTLimitsAndSend(
			ctx, mr.TakerAddress, accToCoins.AccAddress, accToCoins.Coin1, accToCoins.Coin2,
		); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applyMatchingResultMakerUnlockAndSend(
	ctx sdk.Context,
	mr *MatchingResult,
	accCache map[uint64]sdk.AccAddress,
) error {
	accsToCoins := newAccountsToCoins()
	for _, item := range mr.MakerUnlockAndSend {
		makerAddr, err := k.getAccountAddressWithCache(ctx, item.MakerAccNumber, accCache)
		if err != nil {
			return err
		}
		accsToCoins.Add(makerAddr, item.UnlockAndSendCoin, item.DecreaseExpectedToReceiveCoin)

		// add sent part
		for i := range mr.MakerOrderReducedEvents {
			if mr.MakerOrderReducedEvents[i].Creator == makerAddr.String() &&
				mr.MakerOrderReducedEvents[i].ID == item.MakerOrderID {
				mr.MakerOrderReducedEvents[i].SentCoin = item.UnlockAndSendCoin
			}
		}

		mr.TakerOrderReducedEvent.ReceivedCoin = mr.TakerOrderReducedEvent.ReceivedCoin.Add(item.UnlockAndSendCoin)
	}

	for _, accToCoins := range accsToCoins.mapping {
		if err := k.decreaseFTLimitsAndSend(
			ctx, accToCoins.AccAddress, mr.TakerAddress, accToCoins.Coin1, accToCoins.Coin2,
		); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applyMatchingResultMakerUnlock(
	ctx sdk.Context,
	mr *MatchingResult,
	accCache map[uint64]sdk.AccAddress,
) error {
	accsToCoins := newAccountsToCoins()
	for _, item := range mr.MakerUnlock {
		makerAddr, err := k.getAccountAddressWithCache(ctx, item.MakerAccNumber, accCache)
		if err != nil {
			return err
		}
		accsToCoins.Add(makerAddr, item.UnlockCoin, item.DecreaseExpectedToReceiveCoin)
	}

	for _, accToCoins := range accsToCoins.mapping {
		if err := k.decreaseFTLimits(
			ctx, accToCoins.AccAddress, accToCoins.Coin1, accToCoins.Coin2,
		); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applyMatchingResultMakerRemoveRecords(
	ctx sdk.Context,
	mr *MatchingResult,
	accCache map[uint64]sdk.AccAddress,
) error {
	for _, item := range mr.MakerRemoveRecords {
		makerAddr, err := k.getAccountAddressWithCache(ctx, item.AccountNumber, accCache)
		if err != nil {
			return err
		}
		if err := k.removeOrderByRecord(ctx, makerAddr, *item); err != nil {
			return err
		}
	}
	return nil
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
		evt := evt
		if err := ctx.EventManager().EmitTypedEvent(&evt); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderReduced: %s", err)
		}
	}

	return nil
}
