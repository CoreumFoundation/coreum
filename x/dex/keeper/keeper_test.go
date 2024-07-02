package keeper_test

import (
	"sort"
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
			"0.1",
			"0.2",
		},
		{
			"0.01",
			"0.3",
		},
		{
			"0.00000000001",
			"0.0000000003",
			"0.1000000000000000001",
			"0.0000000000000000001",
		},
		{
			"0.000000000000000000000001",
		},
		{
			"1.000000000000000000000001",
			"123239123000000000003123123211.000000000000000000000001",
		},
		{
			"3123123211.000000000000000000000001",
			"1.000000000000000000000001",
		},
		{
			"12323123233212300000000000000000000000.12323123233212400000000000000000000001",
			"12323123233212300000000000000000000000.12323123233212300000000000000000000001",
			"12323123233212300000000000000000000000.12323123233212300000000000000000000002",
		},
	}

	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	testApp.EndBlockAndCommit(sdkCtx)
	dexKeeper := testApp.DEXKeeper
	// TODO(dzmitryhil) replace with SDK generator once we implement it
	orderSeq := uint64(0)
	pairID := uint64(1)
	side := types.Side_buy

	expectedPriceOrder := make([]types.Price, 0)
	for _, priceGroup := range priceGroups {
		sdkCtx = testApp.BeginNextBlock(time.Now())
		// check after beginning of a new block
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side, expectedPriceOrder)
		for _, priceStr := range priceGroup {
			price, err := types.NewPriceFromString(priceStr)
			// append the price and sort as expected to be sorted by the store
			expectedPriceOrder = append(expectedPriceOrder, price)
			sort.Slice(expectedPriceOrder, func(i, j int) bool {
				// TODO(dzmitryhil) extend with the orderSeq once we implement the seq on the SDK level
				return expectedPriceOrder[i].Rat().Cmp(expectedPriceOrder[j].Rat()) == -1
			})

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
			assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side, expectedPriceOrder)
		}
		// check before commit
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side, expectedPriceOrder)
		testApp.EndBlockAndCommit(sdkCtx)
		// check after commit
		assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side, expectedPriceOrder)
	}
	// check final state
	assertPriceOrdersOrdering(t, dexKeeper, sdkCtx, pairID, side, expectedPriceOrder)
}

//nolint:unparam // same params since we have single tests now
func assertPriceOrdersOrdering(
	t *testing.T,
	dexKeeper keeper.Keeper,
	sdkCtx sdk.Context,
	pairID uint64,
	side types.Side,
	expectedPriceOrder []types.Price,
) {
	t.Helper()
	records := make([]types.OrderBookRecord, 0)
	require.NoError(t,
		dexKeeper.IterateOrderBook(
			sdkCtx,
			pairID,
			side,
			false,
			func(record types.OrderBookRecord) (bool, error) {
				records = append(records, record)
				return false, nil
			}))
	require.Len(t, records, len(expectedPriceOrder))
	for i := range expectedPriceOrder {
		require.Equal(t, expectedPriceOrder[i].String(), records[i].Price.String())
	}
}
