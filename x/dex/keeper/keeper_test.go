package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

func TestKeeper_PlaceOrder_OrderBookIDs(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)

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
		price := types.MustNewPriceFromString("1")
		acc, _ := testApp.GenAccount(sdkCtx)
		order := types.Order{
			Creator:    acc.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  item.baseDenom,
			QuoteDenom: item.quoteDenom,
			Price:      price,
			Quantity:   sdkmath.NewInt(1),
			Side:       types.SIDE_SELL,
		}
		lockedBalance, err := order.ComputeLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))

		require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
		selfOrderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, item.baseDenom, item.quoteDenom)
		require.NoError(t, err)
		oppositeOrderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, item.quoteDenom, item.baseDenom)
		require.NoError(t, err)

		require.Equal(t, item.expectedSelfOrderBookID, selfOrderBookID)
		require.Equal(t, item.expectedOppositeOrderBookID, oppositeOrderBookID)
	}
}

func TestKeeper_PlaceAndGetOrderByID(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	price := types.MustNewPriceFromString("12e-1")
	acc, _ := testApp.GenAccount(sdkCtx)

	sellOrder := types.Order{
		Creator:    acc.String(),
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      price,
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
	}
	lockedBalance, err := sellOrder.ComputeLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))

	// try to place the sellOrder one more time
	err = dexKeeper.PlaceOrder(sdkCtx, sellOrder)
	require.ErrorIs(t, err, types.ErrInvalidInput)
	require.ErrorContains(t, err, "is already created")

	gotOrder, err := dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(sellOrder.Creator), sellOrder.ID,
	)
	require.NoError(t, err)

	// set expected values
	sellOrder.RemainingQuantity = sdkmath.NewInt(10)
	sellOrder.RemainingBalance = sdkmath.NewInt(10)
	require.Equal(t, sellOrder, gotOrder)

	// check same buy with the buy order

	buyOrder := types.Order{
		Creator:    acc.String(),
		ID:         uuid.Generate().String(),
		BaseDenom:  denom2,
		QuoteDenom: denom3,
		Price:      price,
		Quantity:   sdkmath.NewInt(100),
		Side:       types.SIDE_BUY,
	}
	lockedBalance, err = buyOrder.ComputeLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))

	gotOrder, err = dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(buyOrder.Creator), buyOrder.ID,
	)
	require.NoError(t, err)

	// set expected values
	buyOrder.RemainingQuantity = sdkmath.NewInt(100)
	buyOrder.RemainingBalance = sdkmath.NewInt(120)
	require.Equal(t, buyOrder, gotOrder)
}

func TestKeeper_PlaceAndCancelOrder(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper
	assetFTKeeper := testApp.AssetFTKeeper

	acc, _ := testApp.GenAccount(sdkCtx)

	sellOrder := types.Order{
		Creator:    acc.String(),
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      types.MustNewPriceFromString("12e-1"),
		Quantity:   sdkmath.NewInt(1_000),
		Side:       types.SIDE_SELL,
	}
	sellLockedBalance, err := sellOrder.ComputeLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(sellLockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))
	dexLockedBalance := assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.Equal(t, sellLockedBalance.String(), dexLockedBalance.String())

	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, sellOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())

	buyOrder := types.Order{
		Creator:    acc.String(),
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      types.MustNewPriceFromString("13e-1"),
		Quantity:   sdkmath.NewInt(5_000),
		Side:       types.SIDE_BUY,
	}
	buyLockedBalance, err := buyOrder.ComputeLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(buyLockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.Equal(t, buyLockedBalance.String(), dexLockedBalance.String())

	// check unlocking
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())

	// now place both orders to let them match partially
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))

	_, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, sellOrder.ID)
	require.ErrorIs(t, err, types.ErrRecordNotFound)
	buyOrder, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, buyOrder.ID)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(5300).String(), buyOrder.RemainingBalance.String())
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())
}

func TestKeeper_GetOrdersAndOrderBookOrders(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)

	orders := []types.Order{
		{
			Creator:    acc1.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("13e-1"),
			Quantity:   sdkmath.NewInt(2000),
			Side:       types.SIDE_SELL,
		},
		{
			Creator:    acc1.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom3,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("14e-1"),
			Quantity:   sdkmath.NewInt(3000),
			Side:       types.SIDE_BUY,
		},
		{
			Creator:    acc1.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("12e-1"),
			Quantity:   sdkmath.NewInt(1000),
			Side:       types.SIDE_SELL,
		},
		{
			Creator:    acc2.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("11e-1"),
			Quantity:   sdkmath.NewInt(100),
			Side:       types.SIDE_BUY,
		},
	}

	for i, order := range orders {
		lockedBalance, err := order.ComputeLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))
		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
		// fill order with the remaining quantity for assertions
		order.RemainingQuantity = order.Quantity
		order.RemainingBalance = lockedBalance.Amount
		orders[i] = order
	}

	// get account orders
	acc1Orders, pageRes, err := testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Offset:     0,
		Limit:      2,
		CountTotal: true,
	})
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(3), pageRes.Total)
	require.Len(t, acc1Orders, 2)

	acc1Orders, _, err = testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Limit: query.PaginationMaxLimit,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []types.Order{
		orders[0], orders[1], orders[2],
	}, acc1Orders)

	// get order book orders
	denom1To2Orders, pageRes, err := testApp.DEXKeeper.GetOrderBookOrders(
		sdkCtx,
		denom1,
		denom2,
		types.SIDE_SELL,
		&query.PageRequest{
			Offset:     0,
			Limit:      1,
			CountTotal: true,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(2), pageRes.Total)
	require.Len(t, denom1To2Orders, 1)

	denom1To2Orders, _, err = testApp.DEXKeeper.GetOrderBookOrders(
		sdkCtx,
		denom1,
		denom2,
		types.SIDE_SELL,
		&query.PageRequest{
			Limit: query.PaginationMaxLimit,
		},
	)
	require.NoError(t, err)
	require.ElementsMatch(t, []types.Order{
		orders[0], orders[2],
	}, denom1To2Orders)
}

func TestKeeper_GetOrderBooks(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)

	orders := []types.Order{
		{
			Creator:    acc1.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("12e-1"),
			Quantity:   sdkmath.NewInt(10),
			Side:       types.SIDE_SELL,
		},
		{
			Creator:    acc1.String(),
			ID:         uuid.Generate().String(),
			BaseDenom:  denom3,
			QuoteDenom: denom2,
			Price:      types.MustNewPriceFromString("13e-1"),
			Quantity:   sdkmath.NewInt(10),
			Side:       types.SIDE_BUY,
		},
	}

	for _, order := range orders {
		lockedBalance, err := order.ComputeLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))
		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
	}

	orderBooks, pageRes, err := testApp.DEXKeeper.GetOrderBooks(sdkCtx, &query.PageRequest{
		Offset:     0,
		Limit:      3,
		CountTotal: true,
	})
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(4), pageRes.Total)
	require.Equal(t, []types.OrderBookData{
		{
			BaseDenom:  denom1,
			QuoteDenom: denom2,
		},
		{
			BaseDenom:  denom2,
			QuoteDenom: denom1,
		},
		{
			BaseDenom:  denom2,
			QuoteDenom: denom3,
		},
	}, orderBooks)
}

func getSorterOrderBookOrders(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) []types.Order {
	records := getSorterOrderBookRecords(t, testApp, sdkCtx, orderBookID, side)
	orders := make([]types.Order, 0, len(records))
	authQueryServer := authkeeper.NewQueryServer(testApp.AccountKeeper)
	for _, record := range records {
		resp, err := authQueryServer.AccountAddressByID(
			sdkCtx,
			&authtypes.QueryAccountAddressByIDRequest{AccountId: record.AccountNumber},
		)
		require.NoError(t, err)
		addr := sdk.MustAccAddressFromBech32(resp.AccountAddress)
		order, err := testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, addr, record.OrderID)
		require.NoError(t, err)
		orders = append(orders, order)
	}

	return orders
}

func getSorterOrderBookRecords(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) []types.OrderBookRecord {
	records := make([]types.OrderBookRecord, 0)
	iterator := testApp.DEXKeeper.NewOrderBookSideIterator(sdkCtx, orderBookID, side)
	defer iterator.Close()

	for {
		record, found, err := iterator.Next()
		require.NoError(t, err)
		if !found {
			break
		}
		records = append(records, record)
	}

	return records
}
