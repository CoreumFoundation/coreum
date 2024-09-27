package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// CoinToAccNumber is a coin to account number struct.
type CoinToAccNumber struct {
	AccNumber uint64
	Coin      sdk.Coin
}

// MatchingResult is the result of a matching operation.
type MatchingResult struct {
	TakerAccNumber     uint64
	TakerSend          []CoinToAccNumber
	MakerUnlockAndSend []CoinToAccNumber
	MakerUnlock        []CoinToAccNumber
	MakerRemoveRecords []types.OrderBookRecord
	MakerUpdateRecord  *types.OrderBookRecord
}

// NewMatchingResult creates a new MatchingResult.
func NewMatchingResult(takerAccNumber uint64) *MatchingResult {
	return &MatchingResult{
		TakerAccNumber:     takerAccNumber,
		TakerSend:          make([]CoinToAccNumber, 0),
		MakerUnlockAndSend: make([]CoinToAccNumber, 0),
		MakerUnlock:        make([]CoinToAccNumber, 0),
		MakerRemoveRecords: make([]types.OrderBookRecord, 0),
		MakerUpdateRecord:  nil,
	}
}

// RegisterTakerSend sets the coin to send.
func (mr *MatchingResult) RegisterTakerSend(makerAccNumber uint64, coin sdk.Coin) {
	if coin.IsZero() {
		return
	}
	mr.TakerSend = appendOrAddCoinToAccNumber(mr.TakerSend, CoinToAccNumber{
		AccNumber: makerAccNumber,
		Coin:      coin,
	})
}

// RegisterMakerUnlockAndSend sets the coin to unlock and send.
func (mr *MatchingResult) RegisterMakerUnlockAndSend(makerAccNumber uint64, coin sdk.Coin) {
	if coin.IsZero() {
		return
	}

	mr.MakerUnlockAndSend = appendOrAddCoinToAccNumber(mr.MakerUnlockAndSend, CoinToAccNumber{
		AccNumber: makerAccNumber,
		Coin:      coin,
	})
}

// RegisterMakerUnlock sets the coin to unlock.
func (mr *MatchingResult) RegisterMakerUnlock(makerAccNumber uint64, coin sdk.Coin) {
	if coin.IsZero() {
		return
	}

	mr.MakerUnlock = appendOrAddCoinToAccNumber(mr.MakerUnlock, CoinToAccNumber{
		AccNumber: makerAccNumber,
		Coin:      coin,
	})
}

// RegisterMakerRemoveRecord sets the record to remove.
func (mr *MatchingResult) RegisterMakerRemoveRecord(record types.OrderBookRecord) {
	mr.MakerRemoveRecords = append(mr.MakerRemoveRecords, record)
}

// RegisterMakerUpdateRecord sets the record to update.
func (mr *MatchingResult) RegisterMakerUpdateRecord(record types.OrderBookRecord) {
	mr.MakerUpdateRecord = &record
}

func appendOrAddCoinToAccNumber(coins []CoinToAccNumber, coin CoinToAccNumber) []CoinToAccNumber {
	for i := range coins {
		if coins[i].AccNumber == coin.AccNumber {
			coins[i].Coin = coins[i].Coin.Add(coin.Coin)
			return coins
		}
	}

	return append(coins, coin)
}

func (k Keeper) applyMatchingResult(ctx sdk.Context, mr *MatchingResult, order types.Order) error {
	accCache := make(map[uint64]sdk.AccAddress)
	takerAddr, err := k.getAccountAddressWithCache(ctx, mr.TakerAccNumber, accCache)
	if err != nil {
		return err
	}
	for _, s := range mr.TakerSend {
		var makerAddr sdk.AccAddress
		makerAddr, err = k.getAccountAddressWithCache(ctx, s.AccNumber, accCache)
		if err != nil {
			return err
		}
		if err := k.sendCoinWithLockCheck(ctx, takerAddr, makerAddr, s.Coin, order.GetReceiveDenom()); err != nil {
			return err
		}
	}
	for _, us := range mr.MakerUnlockAndSend {
		var makerAddr sdk.AccAddress
		makerAddr, err = k.getAccountAddressWithCache(ctx, us.AccNumber, accCache)
		if err != nil {
			return err
		}
		if err := k.unlockAndSendCoin(ctx, makerAddr, takerAddr, us.Coin); err != nil {
			return err
		}
	}
	for _, u := range mr.MakerUnlock {
		var makerAddr sdk.AccAddress
		makerAddr, err = k.getAccountAddressWithCache(ctx, u.AccNumber, accCache)
		if err != nil {
			return err
		}
		if err := k.unlockCoin(ctx, makerAddr, u.Coin); err != nil {
			return err
		}
	}
	for _, record := range mr.MakerRemoveRecords {
		if err := k.removeOrderByRecordAndUsedDenoms(ctx, record, order.Denoms()); err != nil {
			return err
		}
	}
	if mr.MakerUpdateRecord != nil {
		if err := k.saveOrderBookRecord(ctx, *mr.MakerUpdateRecord); err != nil {
			return err
		}
	}

	return nil
}
