package keeper_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func FuzzOrderedPriceStore(f *testing.F) {
	f.Add(uint64(0), int8(0))
	f.Add(uint64(123), types.MaxExp)
	f.Add(uint64(4123123123), types.MinExt)
	f.Add(uint64(9999999999999999999), types.MaxExp)
	f.Add(uint64(1), types.MinExt)

	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	testApp.EndBlockAndCommit(sdkCtx)
	dexKeeper := testApp.DEXKeeper

	lock := sync.Mutex{}
	pairID := uint64(1)
	side := types.Side_buy
	orderSeq := uint64(1)

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		lock.Lock()
		defer lock.Unlock()

		// prepare valid price
		var expPart string
		if exp != 0 {
			expPart = types.ExponentSymbol + strconv.Itoa(int(exp))
		}
		numPart := strconv.FormatUint(num, 10)
		if strings.HasSuffix(numPart, "0") || len(numPart) > types.MaxNumLen {
			t.Skip()
		}
		if exp > types.MaxExp || exp < types.MinExt {
			t.Skip()
		}
		priceStr := strconv.FormatUint(num, 10) + expPart
		price, err := types.NewPriceFromString(priceStr)
		require.NoError(t, err)

		// save price to the store
		sdkCtx = testApp.BeginNextBlock(time.Now())
		orderSeq++
		r := types.OrderBookRecord{
			PairID:            pairID,
			Side:              side,
			Price:             price,
			OrderSeq:          orderSeq,
			OrderID:           uuid.Generate().String(),
			AccountID:         "acc",
			RemainingQuantity: sdkmath.NewInt(1),
			RemainingBalance:  sdkmath.NewInt(2),
		}
		require.NoError(t, dexKeeper.SaveOrderBookRecord(sdkCtx, r))
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)

		testApp.EndBlockAndCommit(sdkCtx)
	})
}
