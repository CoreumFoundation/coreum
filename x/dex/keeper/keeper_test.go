package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})

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
			Side:       types.Side_sell,
		}
		lockedBalance, err := order.ComputeLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

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
	sdkCtx := testApp.BaseApp.NewContext(false, tmproto.Header{})
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
		Side:       types.Side_sell,
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
		Side:       types.Side_buy,
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

func getOrderBookOrders(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) []types.Order {
	records := getOrderBookRecords(t, testApp, sdkCtx, orderBookID, side)
	orders := make([]types.Order, 0, len(records))
	for _, record := range records {
		addr := sdk.MustAccAddressFromBech32(testApp.AccountKeeper.GetAccountAddressByID(sdkCtx, record.AccountNumber))
		order, err := testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, addr, record.OrderID)
		require.NoError(t, err)
		orders = append(orders, order)
	}

	return orders
}

func getOrderBookRecords(
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
