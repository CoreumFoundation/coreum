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

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

func TestKeeper_PlaceOrder_OrderBookIDs(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	type denomsToOrderBookIDs struct {
		baseDenom                   string
		quoteDenom                  string
		expectedSelfOrderBookID     uint32
		expectedOppositeOrderBookID uint32
	}

	for _, item := range []denomsToOrderBookIDs{
		// save with asc denoms ordering
		{
			baseDenom:                   denom1,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(0),
			expectedOppositeOrderBookID: uint32(1),
		},
		// save one more time to check that returns the same
		{
			baseDenom:                   denom1,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(0),
			expectedOppositeOrderBookID: uint32(1),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom1,
			expectedSelfOrderBookID:     uint32(1),
			expectedOppositeOrderBookID: uint32(0),
		},
		// save with desc denoms ordering
		{
			baseDenom:                   denom3,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(3),
			expectedOppositeOrderBookID: uint32(2),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom3,
			expectedSelfOrderBookID:     uint32(2),
			expectedOppositeOrderBookID: uint32(3),
		},
	} {
		price, err := types.NewPriceFromString("1")
		require.NoError(t, err)
		acc, _ := testApp.GenAccount(sdkCtx)
		order := types.Order{
			Account:    acc.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  item.baseDenom,
			QuoteDenom: item.quoteDenom,
			Price:      price,
			Quantity:   sdkmath.NewInt(1),
			Side:       types.Side_sell,
		}

		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
		selfOrderBookID, found, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, item.baseDenom, item.quoteDenom)
		require.NoError(t, err)
		require.True(t, found)
		oppositeOrderBookID, found, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, item.quoteDenom, item.baseDenom)
		require.NoError(t, err)
		require.True(t, found)

		require.Equal(t, item.expectedSelfOrderBookID, selfOrderBookID)
		require.Equal(t, item.expectedOppositeOrderBookID, oppositeOrderBookID)
	}
}

func TestKeeper_PlaceAndGetOrderByID(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	denom1 := denom1
	price1, err := types.NewPriceFromString("1")
	require.NoError(t, err)

	acc1, _ := testApp.GenAccount(sdkCtx)

	order1 := types.Order{
		Account:    acc1.String(),
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      price1,
		Quantity:   sdkmath.NewInt(1),
		Side:       types.Side_sell,
	}

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order1))
	gotOrder, found, err := dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(order1.Account), order1.ID,
	)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, order1, gotOrder)

	// try to place the order one more time
	require.ErrorContains(t, dexKeeper.PlaceOrder(sdkCtx, order1), "is already created")
}

func TestKeeper_PlaceOrder_Ordering(t *testing.T) {
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

	baseDenom := denom1
	quoteDenom := denom2

	side := types.Side_buy
	quantity := sdkmath.NewInt(1)

	var (
		orderBookID        uint32
		orderBookIsCreated bool
	)
	for _, priceGroup := range priceGroups {
		sdkCtx = testApp.BeginNextBlock(time.Now())
		if orderBookIsCreated {
			// check after beginning of a new block
			assertOrdersOrdering(t, dexKeeper, sdkCtx, orderBookID, side)
		}
		for _, priceStr := range priceGroup {
			price, err := types.NewPriceFromString(priceStr)
			require.NoError(t, err)
			acc, _ := testApp.GenAccount(sdkCtx)
			r := types.Order{
				Account:    acc.String(),
				ID:         uuid.Generate().String(),
				BaseDenom:  baseDenom,
				QuoteDenom: quoteDenom,
				Price:      price,
				Quantity:   quantity,
				Side:       side,
			}
			require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, r))

			var found bool
			orderBookID, found, err = dexKeeper.GetOrderBookIDByDenoms(sdkCtx, baseDenom, quoteDenom)
			require.NoError(t, err)
			require.True(t, found)
			orderBookIsCreated = true

			// check just after saving
			assertOrdersOrdering(t, dexKeeper, sdkCtx, orderBookID, side)
		}
		// check before commit
		assertOrdersOrdering(t, dexKeeper, sdkCtx, orderBookID, side)
		testApp.EndBlockAndCommit(sdkCtx)
		// check after commit
		assertOrdersOrdering(t, dexKeeper, sdkCtx, orderBookID, side)
	}
	// check final state
	assertOrdersOrdering(t, dexKeeper, sdkCtx, orderBookID, side)
}

func assertOrdersOrdering(
	t *testing.T,
	dexKeeper keeper.Keeper,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) {
	t.Helper()
	storedRecords := make([]types.OrderBookRecord, 0)
	require.NoError(t,
		dexKeeper.IterateOrderBook(
			sdkCtx,
			orderBookID,
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
