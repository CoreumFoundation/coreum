package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestKeeper_SaveOrderBookRecordAndCheckReadingOrdering(t *testing.T) {
	priceGroups := [][]string{
		{
			"2e-1",
			"1e-1",
		},
		{
			"1e-2",
			"3e-1",
		},
		{
			"1e-10",
			"3e-10",
			"1000000000000000001e-19",
			"1e-19",
		},
		{
			"1e-25",
		},
		{
			"1230000000000000004e-90",
			"1231231241245135243e-90",
		},
		{
			"1230000000000000004e90",
			"1231231241245135243e90",
		},
		{
			"1e10",
			"101e10",
		},
		{
			"12e21",
			"101e-10",
			"1e-9",
		},
		{
			"1e-100",
			"9999999999999999999e100",
		},
		{
			"1e-100",
			"1231239e-32",
		},
	}

	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	testApp.EndBlockAndCommit(sdkCtx)
	dexKeeper := testApp.DEXKeeper
	// TODO(dzmitryhil) replace with SDK generator once we implement it
	orderSeq := uint64(1)
	pairID := uint64(1)
	side := types.Side_buy

	for _, priceGroup := range priceGroups {
		sdkCtx = testApp.BeginNextBlock(time.Now())
		// check after beginning of a new block
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)
		for _, priceStr := range priceGroup {
			price, err := types.NewPriceFromString(priceStr)
			require.NoError(t, err)
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
			orderSeq++
			// check just after saving
			assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)
		}
		// check before commit
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)
		testApp.EndBlockAndCommit(sdkCtx)
		// check after commit
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)
	}
	// check final state
	assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side)
}

func assertPriceOrdersOrdering(
	t *testing.T,
	dexKeeper keeper.Keeper,
	sdkCtx sdk.Context,
	pairID uint64,
	side types.Side,
) {
	t.Helper()
	storedRecords := make([]types.OrderBookRecord, 0)
	require.NoError(t,
		dexKeeper.IterateOrderBook(
			sdkCtx,
			pairID,
			side,
			false,
			func(record types.OrderBookRecord) (bool, error) {
				storedRecords = append(storedRecords, record)
				return false, nil
			}))

	// assert ordering
	for i := 0; i < len(storedRecords)-1; i++ {
		left := storedRecords[i]
		right := storedRecords[i+1]
		require.True(t, //nolint:testifylint // require.NotEqual shouldn't be used here
			// left.Price <= right.Price
			left.Price.Rat().Cmp(right.Price.Rat()) != 1,
			fmt.Sprintf("left price: %s > right price: %s", left.Price.String(), right.Price.String()),
		)
		if left.Price.Rat().Cmp(right.Price.Rat()) == 0 {
			require.Less(t,
				left.OrderSeq, right.OrderSeq,
				fmt.Sprintf("left order seq: %d >= right order seq: %d", left.OrderSeq, right.OrderSeq),
			)
		}
	}
}
