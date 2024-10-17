package dex_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()

	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	const (
		denom1 = "denom1"
		denom2 = "denom2"
	)
	settings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "SMB",
		Subunit:       "sut",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(100),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_order_cancellation,
		},
	}
	denom3, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
	requireT.NoError(err)

	prams := types.DefaultParams()
	genState := types.GenesisState{
		Params: prams,
		OrderBooks: []types.OrderBookDataWithID{
			{
				ID: 0,
				Data: types.OrderBookData{
					BaseDenom:  denom1,
					QuoteDenom: denom2,
				},
			},
			{
				ID: 1,
				Data: types.OrderBookData{
					BaseDenom:  denom2,
					QuoteDenom: denom1,
				},
			},
			{
				ID: 2,
				Data: types.OrderBookData{
					BaseDenom:  denom2,
					QuoteDenom: denom3,
				},
			},
			{
				ID: 3,
				Data: types.OrderBookData{
					BaseDenom:  denom3,
					QuoteDenom: denom2,
				},
			},
		},
		Orders: []types.OrderWithSequence{
			{
				Sequence: 0,
				Order: types.Order{
					Creator:    acc1.String(),
					Type:       types.ORDER_TYPE_LIMIT,
					ID:         "id1",
					BaseDenom:  denom1,
					QuoteDenom: denom2,
					Price:      lo.ToPtr(types.MustNewPriceFromString("1e-2")),
					Quantity:   sdkmath.NewInt(100),
					Side:       types.SIDE_BUY,
					GoodTil: &types.GoodTil{
						GoodTilBlockHeight: 1000,
					},
					TimeInForce:       types.TIME_IN_FORCE_GTC,
					RemainingQuantity: sdkmath.NewInt(90),
					RemainingBalance:  sdkmath.NewInt(90),
					Reserve:           prams.OrderReserve,
				},
			},
			{
				Sequence: 1,
				Order: types.Order{
					Creator:           acc2.String(),
					Type:              types.ORDER_TYPE_LIMIT,
					ID:                "id2",
					BaseDenom:         denom2,
					QuoteDenom:        denom1,
					Price:             lo.ToPtr(types.MustNewPriceFromString("3e3")),
					Quantity:          sdkmath.NewInt(100),
					Side:              types.SIDE_SELL,
					TimeInForce:       types.TIME_IN_FORCE_GTC,
					RemainingQuantity: sdkmath.NewInt(90),
					RemainingBalance:  sdkmath.NewInt(90),
					Reserve:           prams.OrderReserve,
				},
			},
			{
				Sequence: 2,
				Order: types.Order{
					Creator:    acc2.String(),
					Type:       types.ORDER_TYPE_LIMIT,
					ID:         "id3",
					BaseDenom:  denom2,
					QuoteDenom: denom3,
					Price:      lo.ToPtr(types.MustNewPriceFromString("11111e12")),
					Quantity:   sdkmath.NewInt(10000000),
					Side:       types.SIDE_BUY,
					GoodTil: &types.GoodTil{
						GoodTilBlockHeight: 323,
					},
					TimeInForce:       types.TIME_IN_FORCE_GTC,
					RemainingQuantity: sdkmath.NewInt(70000),
					RemainingBalance:  sdkmath.NewInt(50),
					Reserve:           prams.OrderReserve,
				},
			},
		},
	}

	accountDenomToAccountDenomOrdersCount := make(map[string]types.AccountDenomOrdersCount, 0)
	for _, orderWithSeq := range genState.Orders {
		creator := sdk.MustAccAddressFromBech32(orderWithSeq.Order.Creator)
		accNum := testApp.AccountKeeper.GetAccount(sdkCtx, creator).GetAccountNumber()
		for _, denom := range orderWithSeq.Order.Denoms() {
			key := fmt.Sprintf("%d%s", accNum, denom)
			count, ok := accountDenomToAccountDenomOrdersCount[key]
			if !ok {
				count = types.AccountDenomOrdersCount{
					AccountNumber: accNum,
					Denom:         denom,
					OrdersCount:   0,
				}
			}
			count.OrdersCount++
			accountDenomToAccountDenomOrdersCount[key] = count
		}
		// emulate asset FT locking, expecting that the asset FT imports state before the DEX
		lockedBalance, err := orderWithSeq.Order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(lockedBalance))
		require.NoError(t, testApp.AssetFTKeeper.DEXLock(
			sdkCtx, creator, lockedBalance,
		))
		testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(prams.OrderReserve))
		require.NoError(t, testApp.AssetFTKeeper.DEXLock(
			sdkCtx, creator, orderWithSeq.Order.Reserve,
		))
	}
	// set the correct state
	genState.AccountsDenomsOrdersCounts = lo.Values(accountDenomToAccountDenomOrdersCount)

	// init the keeper
	dex.InitGenesis(sdkCtx, dexKeeper, testApp.AccountKeeper, genState)

	// check imported state
	params := dexKeeper.GetParams(sdkCtx)
	requireT.EqualValues(prams, params)

	// check that export is equal import
	exportedGenState := dex.ExportGenesis(sdkCtx, dexKeeper)
	require.NoError(t, exportedGenState.Validate())

	requireT.EqualValues(genState.Params, exportedGenState.Params)
	requireT.EqualValues(genState.OrderBooks, exportedGenState.OrderBooks)
	requireT.EqualValues(genState.Orders, exportedGenState.Orders)

	// check that imported state is valid

	denom2Count, err := dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denom2)
	require.NoError(t, err)
	require.Equal(t, uint64(2), denom2Count)
	denom3Count, err := dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denom3)
	require.NoError(t, err)
	require.Equal(t, uint64(1), denom3Count)

	// place an order with the existing order book
	orderWithExisingOrderBook := types.Order{
		Creator:     acc2.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          "id4",
		BaseDenom:   denom2,
		QuoteDenom:  denom3,
		Price:       lo.ToPtr(types.MustNewPriceFromString("4e3")),
		Quantity:    sdkmath.NewInt(10000000),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := orderWithExisingOrderBook.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc2, sdk.NewCoins(lockedBalance))
	testApp.MintAndSendCoin(t, sdkCtx, acc2, sdk.NewCoins(params.OrderReserve))
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, orderWithExisingOrderBook))

	// set the expected state
	orderWithExisingOrderBook.RemainingQuantity = sdkmath.NewInt(10000000)
	orderWithExisingOrderBook.RemainingBalance = sdkmath.NewInt(40000000000)
	orderWithExisingOrderBook.Reserve = params.OrderReserve

	// check that denom orders counters are incremented
	denom2Count, err = dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denom2)
	require.NoError(t, err)
	require.Equal(t, uint64(3), denom2Count)
	denom3Count, err = dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denom3)
	require.NoError(t, err)
	require.Equal(t, uint64(2), denom3Count)

	// check that this order sequence is next
	ordersWithSeq, _, err := dexKeeper.GetOrdersWithSequence(
		sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit},
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
	orderWithNewOrderBook := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          "id5",
		BaseDenom:   denom1,
		QuoteDenom:  denom3,
		Price:       lo.ToPtr(types.MustNewPriceFromString("4e3")),
		Quantity:    sdkmath.NewInt(10000000),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err = orderWithNewOrderBook.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(lockedBalance))
	testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(params.OrderReserve))
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, orderWithNewOrderBook))

	// check that order books are generated correctly
	denom1ToDenom3OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, denom1, denom3)
	require.NoError(t, err)
	require.Equal(t, uint32(4), denom1ToDenom3OrderBookID)
	denom3ToDenom1OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, denom3, denom1)
	require.NoError(t, err)
	require.Equal(t, uint32(5), denom3ToDenom1OrderBookID)

	// cancel orders by denom to be sure that the acc-denom-orderSeq mapping is saved
	acc1Orders, _, err := dexKeeper.GetOrders(sdkCtx, acc2, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)
	require.Len(t, acc1Orders, 3)

	require.NoError(t, dexKeeper.CancelOrdersByDenom(sdkCtx, issuer, acc2, denom3))

	acc1Orders, _, err = dexKeeper.GetOrders(sdkCtx, acc2, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)
	require.Len(t, acc1Orders, 1)
}
