package dex_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

func TestInitAndExportGenesis(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(ctx)
	acc2, _ := testApp.GenAccount(ctx)

	prams := dextypes.DefaultParams()
	genState := dextypes.GenesisState{
		Params: prams,
		OrderBooks: []dextypes.OrderBookDataWithID{
			{
				ID: 0,
				Data: dextypes.OrderBookData{
					BaseDenom:  denom1,
					QuoteDenom: denom2,
				},
			},
			{
				ID: 1,
				Data: dextypes.OrderBookData{
					BaseDenom:  denom2,
					QuoteDenom: denom1,
				},
			},
			{
				ID: 2,
				Data: dextypes.OrderBookData{
					BaseDenom:  denom2,
					QuoteDenom: denom3,
				},
			},
			{
				ID: 3,
				Data: dextypes.OrderBookData{
					BaseDenom:  denom3,
					QuoteDenom: denom2,
				},
			},
		},
		Orders: []dextypes.OrderWithSequence{
			{
				Sequence: 0,
				Order: dextypes.Order{
					Creator:           acc1.String(),
					Type:              dextypes.ORDER_TYPE_LIMIT,
					ID:                "id1",
					BaseDenom:         denom1,
					QuoteDenom:        denom2,
					Price:             lo.ToPtr(dextypes.MustNewPriceFromString("1e-2")),
					Quantity:          sdkmath.NewInt(100),
					Side:              dextypes.SIDE_BUY,
					RemainingQuantity: sdkmath.NewInt(90),
					RemainingBalance:  sdkmath.NewInt(90),
				},
			},
			{
				Sequence: 1,
				Order: dextypes.Order{
					Creator:           acc2.String(),
					Type:              dextypes.ORDER_TYPE_LIMIT,
					ID:                "id2",
					BaseDenom:         denom2,
					QuoteDenom:        denom1,
					Price:             lo.ToPtr(dextypes.MustNewPriceFromString("3e3")),
					Quantity:          sdkmath.NewInt(100),
					Side:              dextypes.SIDE_SELL,
					RemainingQuantity: sdkmath.NewInt(90),
					RemainingBalance:  sdkmath.NewInt(90),
				},
			},
			{
				Sequence: 2,
				Order: dextypes.Order{
					Creator:           acc2.String(),
					Type:              dextypes.ORDER_TYPE_LIMIT,
					ID:                "id3",
					BaseDenom:         denom2,
					QuoteDenom:        denom3,
					Price:             lo.ToPtr(dextypes.MustNewPriceFromString("11111e12")),
					Quantity:          sdkmath.NewInt(10000000),
					Side:              dextypes.SIDE_BUY,
					RemainingQuantity: sdkmath.NewInt(70000),
					RemainingBalance:  sdkmath.NewInt(50),
				},
			},
		},
	}
	// init the keeper
	dex.InitGenesis(ctx, dexKeeper, testApp.AccountKeeper, genState)

	// check imported state
	params := dexKeeper.GetParams(ctx)
	requireT.EqualValues(prams, params)

	// check that export is equal import
	exportedGenState := dex.ExportGenesis(ctx, dexKeeper)

	requireT.EqualValues(genState.Params, exportedGenState.Params)
	requireT.EqualValues(genState.OrderBooks, exportedGenState.OrderBooks)
	requireT.EqualValues(genState.Orders, exportedGenState.Orders)

	// check that imported state is valid

	// place an order with the existing order book
	orderWithExisingOrderBook := dextypes.Order{
		Creator:    acc2.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id4",
		BaseDenom:  denom2,
		QuoteDenom: denom3,
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("4e3")),
		Quantity:   sdkmath.NewInt(10000000),
		Side:       dextypes.SIDE_BUY,
	}
	lockedBalance, err := orderWithExisingOrderBook.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, ctx, acc2, sdk.NewCoins(lockedBalance))
	require.NoError(t, dexKeeper.PlaceOrder(ctx, orderWithExisingOrderBook))

	// set the expected state
	orderWithExisingOrderBook.RemainingQuantity = sdkmath.NewInt(10000000)
	orderWithExisingOrderBook.RemainingBalance = sdkmath.NewInt(40000000000)

	// check that this order sequence is next
	ordersWithSeq, _, err := dexKeeper.GetPaginatedOrdersWithSequence(
		ctx, &query.PageRequest{Limit: query.PaginationMaxLimit},
	)
	require.NoError(t, err)

	orderFound := false
	for _, orderWithSeq := range ordersWithSeq {
		if orderWithSeq.Order.Creator == acc2.String() && orderWithSeq.Order.ID == orderWithExisingOrderBook.ID {
			orderFound = true
			// check that next seq is max from imported + 1
			requireT.Equal(uint64(3), orderWithSeq.Sequence)
			requireT.EqualValues(orderWithExisingOrderBook, orderWithSeq.Order)
		}
	}
	require.True(t, orderFound)

	// place an order in the new order book
	orderWithNewOrderBook := dextypes.Order{
		Creator:    acc1.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id5",
		BaseDenom:  denom1,
		QuoteDenom: denom3,
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("4e3")),
		Quantity:   sdkmath.NewInt(10000000),
		Side:       dextypes.SIDE_BUY,
	}
	lockedBalance, err = orderWithNewOrderBook.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, ctx, acc1, sdk.NewCoins(lockedBalance))
	require.NoError(t, dexKeeper.PlaceOrder(ctx, orderWithNewOrderBook))

	// check that order books are generated correctly
	denom1ToDenom3OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(ctx, denom1, denom3)
	require.NoError(t, err)
	require.Equal(t, uint32(4), denom1ToDenom3OrderBookID)
	denom3ToDenom1OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(ctx, denom3, denom1)
	require.NoError(t, err)
	require.Equal(t, uint32(5), denom3ToDenom1OrderBookID)
}
