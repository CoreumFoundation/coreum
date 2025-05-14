package dex_test

import (
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v6/x/dex"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()

	sdkCtx := testApp.NewContextLegacy(false, tmproto.Header{})
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	denoms := lo.Times(3, func(i int) string {
		subunit := fmt.Sprintf("denom%d", i)
		settings := assetfttypes.IssueSettings{
			Issuer:        issuer,
			Symbol:        strings.ToUpper(subunit),
			Subunit:       subunit,
			Precision:     1,
			InitialAmount: sdkmath.NewInt(100),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_dex_order_cancellation,
			},
		}
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		requireT.NoError(err)
		return denom
	})

	prams := types.DefaultParams()
	genState := types.GenesisState{
		Params: prams,
		OrderBooks: []types.OrderBookDataWithID{
			{
				ID: 0,
				Data: types.OrderBookData{
					BaseDenom:  denoms[0],
					QuoteDenom: denoms[1],
				},
			},
			{
				ID: 1,
				Data: types.OrderBookData{
					BaseDenom:  denoms[1],
					QuoteDenom: denoms[0],
				},
			},
			{
				ID: 2,
				Data: types.OrderBookData{
					BaseDenom:  denoms[1],
					QuoteDenom: denoms[2],
				},
			},
			{
				ID: 3,
				Data: types.OrderBookData{
					BaseDenom:  denoms[2],
					QuoteDenom: denoms[1],
				},
			},
		},
		Orders: []types.Order{
			{
				Creator:    acc1.String(),
				Type:       types.ORDER_TYPE_LIMIT,
				ID:         "id1",
				Sequence:   1,
				BaseDenom:  denoms[0],
				QuoteDenom: denoms[1],
				Price:      lo.ToPtr(types.MustNewPriceFromString("1e-2")),
				Quantity:   sdkmath.NewInt(100),
				Side:       types.SIDE_BUY,
				GoodTil: &types.GoodTil{
					GoodTilBlockHeight: 1000,
				},
				TimeInForce:               types.TIME_IN_FORCE_GTC,
				RemainingBaseQuantity:     sdkmath.NewInt(90),
				RemainingSpendableBalance: sdkmath.NewInt(90),
				Reserve:                   prams.OrderReserve,
			},
			{
				Creator:                   acc2.String(),
				Type:                      types.ORDER_TYPE_LIMIT,
				ID:                        "id2",
				Sequence:                  2,
				BaseDenom:                 denoms[1],
				QuoteDenom:                denoms[0],
				Price:                     lo.ToPtr(types.MustNewPriceFromString("3e3")),
				Quantity:                  sdkmath.NewInt(100),
				Side:                      types.SIDE_SELL,
				TimeInForce:               types.TIME_IN_FORCE_GTC,
				RemainingBaseQuantity:     sdkmath.NewInt(90),
				RemainingSpendableBalance: sdkmath.NewInt(90),
				Reserve:                   prams.OrderReserve,
			},
			{
				Creator:    acc2.String(),
				Type:       types.ORDER_TYPE_LIMIT,
				ID:         "id3",
				Sequence:   3,
				BaseDenom:  denoms[1],
				QuoteDenom: denoms[2],
				Price:      lo.ToPtr(types.MustNewPriceFromString("11111e12")),
				Quantity:   sdkmath.NewInt(10000000),
				Side:       types.SIDE_BUY,
				GoodTil: &types.GoodTil{
					GoodTilBlockHeight: 323,
				},
				TimeInForce:               types.TIME_IN_FORCE_GTC,
				RemainingBaseQuantity:     sdkmath.NewInt(70000),
				RemainingSpendableBalance: sdkmath.NewInt(50),
				Reserve:                   prams.OrderReserve,
			},
		},
	}

	accountDenomToAccountDenomOrdersCount := make(map[string]types.AccountDenomOrdersCount, 0)
	for _, order := range genState.Orders {
		creator := sdk.MustAccAddressFromBech32(order.Creator)
		accNum := testApp.AccountKeeper.GetAccount(sdkCtx, creator).GetAccountNumber()
		genState.ReservedOrderIds = append(genState.ReservedOrderIds, types.CreateReserveOrderIDKey(accNum, order.ID))
		for _, denom := range order.Denoms() {
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
		lockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(lockedBalance))
		require.NoError(t, testApp.AssetFTKeeper.DEXIncreaseLocked(
			sdkCtx, creator, lockedBalance,
		))
		testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(prams.OrderReserve))
		require.NoError(t, testApp.AssetFTKeeper.DEXIncreaseLocked(
			sdkCtx, creator, order.Reserve,
		))
	}
	// set the correct state
	genState.AccountsDenomsOrdersCounts = lo.Values(accountDenomToAccountDenomOrdersCount)

	// the order sequence is last order sequence
	genState.OrderSequence = 3

	// init the keeper
	dex.InitGenesis(sdkCtx, dexKeeper, testApp.AccountKeeper, genState)

	// check imported state
	params, err := dexKeeper.GetParams(sdkCtx)
	requireT.NoError(err)
	requireT.Equal(prams, params)

	// check that export is equal import
	exportedGenState := dex.ExportGenesis(sdkCtx, dexKeeper)
	require.NoError(t, exportedGenState.Validate())

	requireT.Equal(genState.Params, exportedGenState.Params)
	requireT.Equal(genState.OrderBooks, exportedGenState.OrderBooks)
	requireT.Equal(genState.Orders, exportedGenState.Orders)

	// check that imported state is valid

	denom2Count, err := dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denoms[1])
	require.NoError(t, err)
	require.Equal(t, uint64(2), denom2Count)
	denom3Count, err := dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denoms[2])
	require.NoError(t, err)
	require.Equal(t, uint64(1), denom3Count)

	// place an order with the existing order book
	orderWithExisingOrderBook := types.Order{
		Creator:     acc2.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          "id4",
		Sequence:    0,
		BaseDenom:   denoms[1],
		QuoteDenom:  denoms[2],
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
	orderWithExisingOrderBook.Sequence = 4
	orderWithExisingOrderBook.RemainingBaseQuantity = sdkmath.NewInt(10000000)
	orderWithExisingOrderBook.RemainingSpendableBalance = sdkmath.NewInt(40000000000)
	orderWithExisingOrderBook.Reserve = params.OrderReserve

	// check that denom orders counters are incremented
	denom2Count, err = dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denoms[1])
	require.NoError(t, err)
	require.Equal(t, uint64(3), denom2Count)
	denom3Count, err = dexKeeper.GetAccountDenomOrdersCount(sdkCtx, acc2, denoms[2])
	require.NoError(t, err)
	require.Equal(t, uint64(2), denom3Count)

	// check that this order sequence is next
	orders, _, err := dexKeeper.GetAccountsOrders(
		sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit},
	)
	require.NoError(t, err)

	orderFound := false
	for _, order := range orders {
		if order.Creator == acc2.String() && order.ID == orderWithExisingOrderBook.ID {
			orderFound = true
			// the `orderWithExisingOrderBook` has the sequence eq to 4 to check that next sequence from imported is used
			requireT.Equal(orderWithExisingOrderBook, order)
		}
	}
	require.True(t, orderFound)

	// place an order in the new order book
	orderWithNewOrderBook := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          "id5",
		BaseDenom:   denoms[0],
		QuoteDenom:  denoms[2],
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
	denom1ToDenom3OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, denoms[0], denoms[2])
	require.NoError(t, err)
	require.Equal(t, uint32(4), denom1ToDenom3OrderBookID)
	denom3ToDenom1OrderBookID, err := dexKeeper.GetOrderBookIDByDenoms(sdkCtx, denoms[2], denoms[0])
	require.NoError(t, err)
	require.Equal(t, uint32(5), denom3ToDenom1OrderBookID)

	// cancel orders by denom to be sure that the acc-denom-order mapping is saved
	acc1Orders, _, err := dexKeeper.GetOrders(sdkCtx, acc2, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)
	require.Len(t, acc1Orders, 3)

	require.NoError(t, dexKeeper.CancelOrdersByDenom(sdkCtx, issuer, acc2, denoms[2]))

	acc1Orders, _, err = dexKeeper.GetOrders(sdkCtx, acc2, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)
	require.Len(t, acc1Orders, 1)
}
