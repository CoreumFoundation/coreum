package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestKeeper_SaveOrderAndReadWithOrderBookIterator(t *testing.T) {
	tests := []struct {
		name        string
		priceGroups [][]string
		side        types.Side
	}{
		{
			name:        "sell_no_record",
			priceGroups: [][]string{},
			side:        types.SIDE_SELL,
		},
		{
			name: "sell_one_record",
			priceGroups: [][]string{
				{
					"99e-3",
				},
			},
			side: types.SIDE_SELL,
		},
		{
			name: "sell_three_records_with_different_prices",
			priceGroups: [][]string{
				{
					"99e-3",
					"9e-3",
					"1e2",
				},
			},
			side: types.SIDE_SELL,
		},
		{
			name: "sell_combined",
			priceGroups: [][]string{
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
			},
			side: types.SIDE_SELL,
		},
		{
			name:        "buy_no_record",
			priceGroups: [][]string{},
			side:        types.SIDE_BUY,
		},
		{
			name: "buy_one_record",
			priceGroups: [][]string{
				{
					"99e-3",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_three_records_with_different_prices",
			priceGroups: [][]string{
				{
					"99e-3",
					"9e-3",
					"1e2",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_multiple_prices_two_same",
			priceGroups: [][]string{
				{
					"1e-1",
					"2e-1",
					"3e-1",
					"2e-1",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_two_same_prices",
			priceGroups: [][]string{
				{
					"1e-1",
					"1e-1",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_tree_same_prices",
			priceGroups: [][]string{
				{
					"1e-1",
					"1e-1",
					"1e-1",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_tree_same_prices_after_different",
			priceGroups: [][]string{
				{
					"1",
					"1e-1",
					"1e-1",
					"1e-1",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_tree_same_prices_before_another",
			priceGroups: [][]string{
				{
					"1e-1",
					"1e-1",
					"1e-1",
					"1e-2",
				},
			},
			side: types.SIDE_BUY,
		},
		{
			name: "buy_combined",
			priceGroups: [][]string{
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
					"1000000000000000001e19",
					"1e-19",
				},
				{
					"1e-25",
				},
				{
					"1230000000000000004e-12",
					"1000000000000000001e19",
				},
				{
					"1230000000000000004e-12",
					"1231231241245135243e23",
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
					"1e-13",
					"1230000000000000004e-12",
				},
				{
					"1e-13",
					"1231239e-32",
				},
			},
			side: types.SIDE_BUY,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testApp := simapp.New()
			sdkCtx := testApp.BaseApp.NewContext(false)
			_, err := testApp.EndBlocker(sdkCtx)
			require.NoError(t, err)

			// don't limit the price tick
			params := testApp.DEXKeeper.GetParams(sdkCtx)
			params.PriceTickExponent = int32(types.MinExt)
			require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

			baseDenom := denom1
			quoteDenom := denom2
			var (
				orderBookID        uint32
				orderBookIsCreated bool
			)
			for _, priceGroup := range tt.priceGroups {
				sdkCtx, _, _ = testApp.BeginNextBlock()
				if orderBookIsCreated {
					// check after beginning of a new block
					assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, tt.side)
				}
				for _, priceStr := range priceGroup {
					price := types.MustNewPriceFromString(priceStr)
					acc, _ := testApp.GenAccount(sdkCtx)

					var quantity sdkmath.Int
					if tt.side == types.SIDE_BUY {
						// make the locked balance as Int for any side
						quantity = sdkmath.NewIntFromBigInt(price.Rat().Denom())
					} else {
						// for the sell side we use constant to test the min and max price
						quantity = sdkmath.NewInt(1)
					}
					order := types.Order{
						Creator:     acc.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          uuid.Generate().String(),
						BaseDenom:   baseDenom,
						QuoteDenom:  quoteDenom,
						Price:       &price,
						Quantity:    quantity,
						Side:        tt.side,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					}

					lockedBalance, err := order.ComputeLimitOrderLockedBalance()
					require.NoError(t, err)
					testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))
					require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

					orderBookID, err = testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, baseDenom, quoteDenom)
					require.NoError(t, err)
					orderBookIsCreated = true

					// check just after saving
					assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, tt.side)
				}
				// check before commit
				assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, tt.side)
				_, err = testApp.EndBlocker(sdkCtx)
				require.NoError(t, err)
				// check after commit
				assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, tt.side)
			}
			// check final state
			assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, tt.side)
		})
	}
}

func assertOrdersOrdering(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) {
	t.Helper()
	storedRecords := getSorterOrderBookRecords(t, testApp, sdkCtx, orderBookID, side)
	if side == types.SIDE_BUY {
		// buy - price desc + order sec asc
		for i := 0; i < len(storedRecords)-1; i++ {
			left := storedRecords[i]
			right := storedRecords[i+1]
			require.True(t, //nolint:testifylint // require.NotEqual shouldn't be used here
				// left.Price >= right.Price
				left.Price.Rat().Cmp(right.Price.Rat()) != -1,
				fmt.Sprintf("left price: %s < right price: %s", left.Price.String(), right.Price.String()),
			)
			if left.Price.Rat().Cmp(right.Price.Rat()) == 0 {
				require.Less(t,
					left.OrderSeq, right.OrderSeq,
					fmt.Sprintf("left order seq: %d >= right order seq: %d", left.OrderSeq, right.OrderSeq),
				)
			}
		}
		return
	}
	// sell - price asc + order sec asc
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
