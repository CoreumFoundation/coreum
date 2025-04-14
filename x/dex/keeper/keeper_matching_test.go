package keeper_test

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

type TestSet struct {
	acc1 sdk.AccAddress
	acc2 sdk.AccAddress
	acc3 sdk.AccAddress

	issuer                                     sdk.AccAddress
	denom1, denom2, denom3                     string
	ftDenomWhitelisting1, ftDenomWhitelisting2 string

	orderReserve sdk.Coin
}

func (t TestSet) orderReserveTimes(times int64) sdk.Coin {
	return sdk.NewCoin(t.orderReserve.Denom, t.orderReserve.Amount.MulRaw(times))
}

type tst struct {
	name                          string
	balances                      func(testSet TestSet) map[string]sdk.Coins
	whitelistedBalances           func(testSet TestSet) map[string]sdk.Coins
	orders                        func(testSet TestSet) []types.Order
	wantOrders                    func(testSet TestSet) []types.Order
	wantAvailableBalances         func(testSet TestSet) map[string]sdk.Coins
	wantExpectedToReceiveBalances func(testSet TestSet) map[string]sdk.Coins
	wantErrorContains             func(testSet TestSet) string
}

func (tt tst) run(t *testing.T) {
	logger := log.NewTestLogger(t)
	testApp := simapp.New(simapp.WithCustomLogger(logger))
	sdkCtx := testApp.BaseApp.NewContext(false)

	testSet := genTestSet(t, sdkCtx, testApp)
	t.Logf(
		"Test set: acc1: %s, acc2: %s, acc3: %s",
		testSet.acc1, testSet.acc2, testSet.acc3,
	)

	if tt.whitelistedBalances != nil {
		for addr, coins := range tt.whitelistedBalances(testSet) {
			testApp.AssetFTKeeper.SetWhitelistedBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
		}
	}

	for addr, coins := range tt.balances(testSet) {
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
	}

	orderBooksIDs := make(map[uint32]struct{})
	initialOrders := tt.orders(testSet)

	ordersDenoms := make(map[string]struct{}, 0)
	for i, order := range initialOrders {
		ordersDenoms[order.BaseDenom] = struct{}{}
		ordersDenoms[order.QuoteDenom] = struct{}{}
		availableBalancesBefore, err := getAvailableBalances(sdkCtx, testApp, sdk.MustAccAddressFromBech32(order.Creator))
		require.NoError(t, err)

		// use new event manager for each order
		sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
		gasBefore := sdkCtx.GasMeter().GasConsumed()
		err = testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
		if err != nil && tt.wantErrorContains != nil {
			require.True(t, sdkerrors.IsOf(
				err,
				assetfttypes.ErrDEXInsufficientSpendableBalance, assetfttypes.ErrWhitelistedLimitExceeded,
			))
			expectedErr := tt.wantErrorContains(testSet)
			require.ErrorContains(t, err, expectedErr)
			return
		}
		gasAfter := sdkCtx.GasMeter().GasConsumed()
		t.Logf("Used gas for order %d placement: %d", i, gasAfter-gasBefore)
		require.NoError(t, err)
		assertOrderPlacementResult(t, sdkCtx, testApp, availableBalancesBefore, order)
		orderBooksID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, order.BaseDenom, order.QuoteDenom)
		require.NoError(t, err)
		orderBooksIDs[orderBooksID] = struct{}{}
	}
	if tt.wantErrorContains != nil {
		expectedErr := tt.wantErrorContains(testSet)
		require.Failf(t, "expected error not found", expectedErr)
	}

	orders := make([]types.Order, 0)
	for orderBookID := range orderBooksIDs {
		orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_BUY)...)
		orders = append(orders, getSorterOrderBookOrders(t, testApp, sdkCtx, orderBookID, types.SIDE_SELL)...)
	}
	wantOrders := tt.wantOrders(testSet)
	// set order reserve and order sequence for all orders
	wantOrders = fillReserveAndOrderSequence(t, sdkCtx, testApp, wantOrders)
	require.ElementsMatch(t, wantOrders, orders, "orders do not match: \n%s", cmp.Diff(wantOrders, orders))

	availableBalances := make(map[string]sdk.Coins)
	lockedBalances := make(map[string]sdk.Coins)
	expectedToReceiveBalances := make(map[string]sdk.Coins)
	for addr := range tt.balances(testSet) {
		addrBalances := testApp.BankKeeper.GetAllBalances(sdkCtx, sdk.MustAccAddressFromBech32(addr))
		addrFTLockedBalances := sdk.NewCoins()
		for _, balance := range addrBalances {
			lockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(addr), balance.Denom,
			)
			addrFTLockedBalances = addrFTLockedBalances.Add(lockedBalance)
			addrBalances = addrBalances.Sub(lockedBalance)
		}

		addrFTExpectedToReceiveBalances := sdk.NewCoins()
		for denom := range ordersDenoms {
			addrFTExpectedToReceiveBalance := testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(addr), denom,
			)
			addrFTExpectedToReceiveBalances = addrFTExpectedToReceiveBalances.Add(addrFTExpectedToReceiveBalance)
		}

		availableBalances[addr] = addrBalances
		lockedBalances[addr] = addrFTLockedBalances
		expectedToReceiveBalances[addr] = addrFTExpectedToReceiveBalances
	}
	availableBalances = removeEmptyBalances(availableBalances)
	lockedBalances = removeEmptyBalances(lockedBalances)
	expectedToReceiveBalances = removeEmptyBalances(expectedToReceiveBalances)

	wantAvailableBalances := tt.wantAvailableBalances(testSet)
	require.True(
		t,
		reflect.DeepEqual(wantAvailableBalances, availableBalances),
		"available balances do not match: %v", cmp.Diff(wantAvailableBalances, availableBalances),
	)

	// by default must be empty
	wantExpectedToReceiveBalances := make(map[string]sdk.Coins)
	if tt.wantExpectedToReceiveBalances != nil {
		wantExpectedToReceiveBalances = tt.wantExpectedToReceiveBalances(testSet)
	}

	require.True(
		t,
		reflect.DeepEqual(wantExpectedToReceiveBalances, expectedToReceiveBalances),
		"expected to receive balances do not match: %v", cmp.Diff(wantAvailableBalances, availableBalances),
	)

	// check that balance locked in the orders correspond the balance locked in the asset ft
	orderLockedBalances := make(map[string]sdk.Coins)
	for _, order := range orders {
		coins, ok := orderLockedBalances[order.Creator]
		if !ok {
			coins = sdk.NewCoins()
		}
		coins = coins.Add(sdk.NewCoin(order.GetSpendDenom(), order.RemainingSpendableBalance))
		params, err := testApp.DEXKeeper.GetParams(sdkCtx)
		require.NoError(t, err)
		// add reserve for each order
		coins = coins.Add(params.OrderReserve)
		orderLockedBalances[order.Creator] = coins
	}
	orderLockedBalances = removeEmptyBalances(orderLockedBalances)
	require.True(
		t,
		reflect.DeepEqual(lockedBalances, orderLockedBalances),
		"want: %v, got: %v", lockedBalances, orderLockedBalances,
	)

	cancelAllOrdersAndAssertState(t, sdkCtx, testApp)
}

func TestKeeper_MatchOrders_Other(t *testing.T) {
	t.SkipNow()
	tests := []tst{
		// ******************** Inverted OB limit matching ********************

		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9625),
						RemainingSpendableBalance: sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_invertedOB_maker_sell_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 9999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return fmt.Sprintf("10000%s is not available, available 9999%s", testSet.denom2, testSet.denom2)
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_maker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1001)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,

						RemainingBaseQuantity:     sdkmath.NewInt(9625),
						RemainingSpendableBalance: sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1),
						sdk.NewInt64Coin(testSet.denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(999),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(7336),
						RemainingSpendableBalance: sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 2664),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 1001),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(7336),
						RemainingSpendableBalance: sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 2664), // 2664 = 999 / 0.375, so 999 matched for price 0.375.
						sdk.NewInt64Coin(testSet.denom2, 2),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 381)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 26506),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10002),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:                  sdkmath.NewInt(10002),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9621),
						RemainingSpendableBalance: sdkmath.NewInt(25496),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 10),
						sdk.NewInt64Coin(testSet.denom2, 381),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_invertedOB_maker_buy_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 381),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 26490),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return fmt.Sprintf("26491%s is not available, available 26490%s", testSet.denom1, testSet.denom1)
			},
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_buy_close_taker_with_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 4234),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 2650),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(11111),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:                  sdkmath.NewInt(11111),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9111),
						RemainingSpendableBalance: sdkmath.NewInt(3472),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 2000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 650),
						sdk.NewInt64Coin(testSet.denom2, 762),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9500),
						RemainingSpendableBalance: sdkmath.NewInt(9500),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 500),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(500),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 1000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 500),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_close_two_makers_buy_and_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 25),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 25),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 105),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(50),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 50)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 50)),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 5),
						sdk.NewInt64Coin(testSet.denom2, 50),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 4995),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too low
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingBaseQuantity:     sdkmath.NewInt(1125),
						RemainingSpendableBalance: sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 120),
						sdk.NewInt64Coin(testSet.denom2, 1835),
					),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_sell_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom1, 2000+2000+1000+2000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 1880),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too low
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("26e-1")),
						Quantity:    sdkmath.NewInt(1880),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(1000),
					},
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(2000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 1878),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 5000),
						sdk.NewInt64Coin(testSet.denom2, 2),
					),
				}
			},
		},

		// ******************** Inverted OB market matching ********************

		{
			name: "match_market_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(testSet.denom2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1000),
						sdk.NewInt64Coin(testSet.denom2, 9625),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_sell_taker_sell_partial_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 9999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1000),
						sdk.NewInt64Coin(testSet.denom2, 9624),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_buy_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 381),
					),
					// ceil(10101*(1/381e-3))
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 26512),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("381e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Quantity:    sdkmath.NewInt(10101),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 25512),
						sdk.NewInt64Coin(testSet.denom2, 381),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_buy_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 380),
					),
					// not enough balance to cover both orders so id2 will be matched partially.
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 100+900-1)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:    sdkmath.NewInt(900),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("38e-2")),
						Quantity:                  sdkmath.NewInt(900),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(50),
						RemainingSpendableBalance: sdkmath.NewInt(19),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100+850),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 49),
						sdk.NewInt64Coin(testSet.denom2, 38+323),
					),
				}
			},
		},
		{
			name: "match_market_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Quantity:    sdkmath.NewInt(999),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(7336),
						RemainingSpendableBalance: sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 999)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 2664)),
				}
			},
		},

		// ******************** Combined matching ********************

		{
			name: "match_limit_directOB_and_invertedOB_buy_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
						// order has fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id3",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
						// will remain with the partial filling
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(500),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(4600),
						RemainingSpendableBalance: sdkmath.NewInt(4600),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 9955),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 45),
						sdk.NewInt64Coin(testSet.denom2, 5500),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_buy_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 500+5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 220),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// order "id1", "id2" and "id3" matches, but we fill partially only "id2"
						//	with the best price and fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("22e-1")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(1000),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(800),
						RemainingSpendableBalance: sdkmath.NewInt(400),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10000),
						RemainingSpendableBalance: sdkmath.NewInt(5000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 200),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 100),
						sdk.NewInt64Coin(testSet.denom2, 20),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 825),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(1500),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(500),
						RemainingSpendableBalance: sdkmath.NewInt(500),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10000),
						RemainingSpendableBalance: sdkmath.NewInt(19000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 250),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1500),
						sdk.NewInt64Coin(testSet.denom2, 75),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 2100),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 2100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						// order "id1", "id2" and "id3" matches, but we fill partially only "id1"
						//	with the best price and fifo priority
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(990),
						RemainingSpendableBalance: sdkmath.NewInt(2079),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(2100),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10000),
						RemainingSpendableBalance: sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 10),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 21),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_same_price_respect_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 25000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// partially matches. Inverted price is the same as id1 but id1 has higher priority because of FIFO
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")), // 0.5 = 1/2 which is inversion of 2
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(25000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(5000),
						RemainingSpendableBalance: sdkmath.NewInt(2500),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 20000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 12500),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_buy_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:    sdkmath.NewInt(100000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc3.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("49e-2")),
						Quantity:                  sdkmath.NewInt(100000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(80719),
						RemainingSpendableBalance: sdkmath.NewInt(80719),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 18281),
					),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 10600)),
				}
			},
		},
		{
			name: "match_limit_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 82500)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(150000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc3.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:                  sdkmath.NewInt(150000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(129000),
						RemainingSpendableBalance: sdkmath.NewInt(70950),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 10500),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 21000),
						sdk.NewInt64Coin(testSet.denom2, 550),
					),
				}
			},
		},
		{
			name: "match_market_directOB_and_invertedOB_buy_close_invertedOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 100+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
						// order has fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id3",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// better price 181e-2 sell ~= 0.55 Inverted OB buy, greater is better price
						// will remain with the partial filling
						Price:       lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(500),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("181e-2")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(4600),
						RemainingSpendableBalance: sdkmath.NewInt(4600),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 9955),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 45),
						sdk.NewInt64Coin(testSet.denom2, 5500),
					),
				}
			},
		},
		{
			name: "match_market_directOB_and_invertedOB_buy_close_directOB_taker_with_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 500+5000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 200),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom2,
						QuoteDenom: testSet.denom1,
						// order "id1", "id2" and "id3" matches, but we fill partially only "id2"
						//	with the best price and fifo priority
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(100),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("21e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(1000),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(800),
						RemainingSpendableBalance: sdkmath.NewInt(400),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10000),
						RemainingSpendableBalance: sdkmath.NewInt(5000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 200)),
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 100)),
				}
			},
		},
		{
			name: "match_market_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 75000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(150000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 10500),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 21000),
						sdk.NewInt64Coin(testSet.denom2, 64000),
					),
				}
			},
		},

		// ******************** IOC matching ********************

		{
			name: "no_match_limit_sell_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1005),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 3760),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 5),
						sdk.NewInt64Coin(testSet.denom2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1000),
						// 3385testSet.denom2 refunded
						sdk.NewInt64Coin(testSet.denom2, 3385),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 1005),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled, but the order will be executed fully
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingBaseQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingSpendableBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 375)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000), sdk.NewInt64Coin(testSet.denom2, 630)),
				}
			},
		},

		// ******************** FOK matching ********************

		{
			name: "no_match_limit_sell_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000)),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000)),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1005+7),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 3760),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(1005),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1005),
						RemainingSpendableBalance: sdkmath.NewInt(1005),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 3760)),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_partial_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000+3)),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 1005),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled, but the order will be executed fully, that's why we cancel the FOK
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10000),
						RemainingSpendableBalance: sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 3)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 1005)),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker_with_full_filling_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10000+3),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 1005),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("1")),
						// only 1000 will be filled fully with the better price
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9000),
						RemainingSpendableBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 3), sdk.NewInt64Coin(testSet.denom2, 375)),
					// expected result + not used amount
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 1000), sdk.NewInt64Coin(testSet.denom2, 630)),
				}
			},
		},

		// ******************** Whitelisting ********************

		{
			name: "no_match_whitelisting_limit_directOB_buy_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 42),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001), // initial balance
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 42), // initial balance
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(111),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(1001),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1001),
						RemainingSpendableBalance: sdkmath.NewInt(1001),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(111),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(111),
						RemainingSpendableBalance: sdkmath.NewInt(42),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					), // floor(376e-3 * 1001) + 1
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(
						testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
		},
		{
			name: "try_to_place_no_match_whitelisting_limit_directOB_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),  // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377-1), // expected to receive - 1
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return "is not enough to receive 377"
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 417),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1001), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),  // expected to receive
					),
					testSet.acc2.String(): sdk.NewCoins(
						// 1000 to receive when match the "id1" order, and 101 when place order to the order book
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+101), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 417),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("39e-2")),
						Quantity:    sdkmath.NewInt(1101),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("39e-2")),
						Quantity:                  sdkmath.NewInt(1101),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(101),
						RemainingSpendableBalance: sdkmath.NewInt(40),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 438),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1101), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 438),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 390),  // expected to receive
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("39e-2")),
						Quantity:    sdkmath.NewInt(1101),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("39e-2")),
						Quantity:                  sdkmath.NewInt(1101),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(101),
						RemainingSpendableBalance: sdkmath.NewInt(40),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						// total balance is 438 where:
						// - 390 = 0.39*1000 - executed
						// - 40 = ceil(0.39*101) - locked
						// - 8 = 438-390-40 - available
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 8),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 390),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 101)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingBaseQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingSpendableBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "try_to_match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000-1),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1001),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return "is not enough to receive 1000"
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376+3375),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingBaseQuantity: sdkmath.NewInt(9000),
						// 10000 - 1000
						RemainingSpendableBalance: sdkmath.NewInt(9000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 376),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3375)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+25507),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.ftDenomWhitelisting2,
						QuoteDenom:                testSet.ftDenomWhitelisting1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9625),
						RemainingSpendableBalance: sdkmath.NewInt(9625),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 25507)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 10000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2664),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("265e-2")),
						Quantity:    sdkmath.NewInt(999),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(10000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(7336),
						RemainingSpendableBalance: sdkmath.NewInt(7336),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 999),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2664),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 2751)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_maker_sell_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2")),
						Quantity:    sdkmath.NewInt(500),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 500),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_and_invertedOB_sell_close_all_makers",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+19000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 82500),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+19000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000+149000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 82500),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("19e-1")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:    sdkmath.NewInt(150000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc3.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("55e-2")),
						Quantity:                  sdkmath.NewInt(150000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(129000),
						RemainingSpendableBalance: sdkmath.NewInt(70950),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 500+10000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 21000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 550),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc3.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 129000)),
				}
			},
		},
		{
			name: "match_whitelisting_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4995),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2000+2000+1000+2000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4995),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1835),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too low
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.ftDenomWhitelisting2,
						QuoteDenom:  testSet.ftDenomWhitelisting1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingBaseQuantity:     sdkmath.NewInt(1125),
						RemainingSpendableBalance: sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 120),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1835),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 2125)),
				}
			},
		},
		{
			name: "match_whitelisting_market_directOB_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4*1000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 4*1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777+777),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 3000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("555e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id5",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Quantity:    sdkmath.NewInt(3000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(1000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(3),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375+555+777),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 3000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 777),
					),
				}
			},
		},
		{
			name: "no_match_whitelisting_limit_sell_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // just initial amount
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling_time_in_force_ioc",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.ftDenomWhitelisting1,
						QuoteDenom: testSet.ftDenomWhitelisting2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						// only 1000 will be filled
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 5),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 375),
					),
					// 3385testSet.ftDenomWhitelisting2 refunded
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3385),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_whitelisting_limit_sell_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// lock required balance for the full order
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_taker_not_enough_market_time_in_force_fok",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005+7),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(
						testSet.ftDenomWhitelisting2, 3760),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1005+7),
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760),
						// we don't need ft2 since we don't apply the matching changes
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1005),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_FOK,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(1005),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1005),
						RemainingSpendableBalance: sdkmath.NewInt(1005),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 7)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 3760)),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 377)),
				}
			},
		},

		{
			name: "match_whitelisting_limit_directOB_maker_sell_taker_buy_close_maker_with_zero_filled_quantity",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			whitelistedBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111), // initial
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 1),   // expected to receive
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000), // expected to receive
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting2, 10000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.ftDenomWhitelisting1,
						QuoteDenom: testSet.ftDenomWhitelisting2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// can't fill since 111 * 376e-5 ~= 0.41736
						Quantity:    sdkmath.NewInt(111),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.ftDenomWhitelisting1,
						QuoteDenom:  testSet.ftDenomWhitelisting2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("1e1")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.ftDenomWhitelisting1,
						QuoteDenom:                testSet.ftDenomWhitelisting2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("1e1")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(10000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 111),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.ftDenomWhitelisting1, 1000)),
				}
			},
		},
		{
			name: "match_limit_invertedOB_multiple_maker_buy_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom2, 754+752+4+752),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 4995),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too low
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:    sdkmath.NewInt(1000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("27e-1")),
						Quantity:    sdkmath.NewInt(1850),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("4e-3")),
						Quantity:                  sdkmath.NewInt(1000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1000),
						RemainingSpendableBalance: sdkmath.NewInt(4),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(2000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// part was used
						RemainingBaseQuantity:     sdkmath.NewInt(1125),
						RemainingSpendableBalance: sdkmath.NewInt(423),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 4875),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 120),
						sdk.NewInt64Coin(testSet.denom2, 1835),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_ten_orders",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(10),
						sdk.NewInt64Coin(testSet.denom1, 100000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(1),
						sdk.NewInt64Coin(testSet.denom2, 98991560000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				orders := make([]types.Order, 0)
				for i := range 10 {
					orders = append(orders, types.Order{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          fmt.Sprintf("id%d", i),
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString(fmt.Sprintf("1%d1e-1", i))),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					})
				}
				orders = append(orders, types.Order{
					Creator:     testSet.acc2.String(),
					Type:        types.ORDER_TYPE_LIMIT,
					ID:          "id101",
					BaseDenom:   testSet.denom1,
					QuoteDenom:  testSet.denom2,
					Price:       lo.ToPtr(types.MustNewPriceFromString("9999")),
					Quantity:    sdkmath.NewInt(10_000_000),
					Side:        types.SIDE_BUY,
					TimeInForce: types.TIME_IN_FORCE_GTC,
				})
				return orders
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id101",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("9999")),
						Quantity:                  sdkmath.NewInt(10_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9900000),
						RemainingSpendableBalance: sdkmath.NewInt(98990100000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(10),
						sdk.NewInt64Coin(testSet.denom2, 1460000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 100000),
					),
				}
			},
		},
	}
	for _, tt := range tests {
		if tt.name == "match_limit_directOB_multiple_maker_sell_taker_buy_close_taker_with_same_price_fifo_priority" {
			break
		}
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestKeeper_MatchOrders_NoMatching(t *testing.T) {
	tests := []tst{
		{
			name: "no_match_limit_directOB_and_invertedOB_buy_and_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
						sdk.NewInt64Coin(testSet.denom2, 1_000_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 2_659_000),
						sdk.NewInt64Coin(testSet.denom2, 375_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("266e-2")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom2,
						QuoteDenom:  testSet.denom1,
						Price:       lo.ToPtr(types.MustNewPriceFromString("2659e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(1_000_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(1_000_000),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(1_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(375_000),
					},
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("266e-2")),
						Quantity:                  sdkmath.NewInt(1_000_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(1_000_000),
					},
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom2,
						QuoteDenom:                testSet.denom1,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("2659e-3")),
						Quantity:                  sdkmath.NewInt(1_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(2_659_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_market_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// no orders in the order book so nothing to lock
				return map[string]sdk.Coins{}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "no_match_market_buy",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				// no orders in the order book so nothing to lock
				return map[string]sdk.Coins{}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
		{
			name: "try_to_match_limit_directOB_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 999_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return fmt.Sprintf("1000000%s is not available, available 999000%s", testSet.denom1, testSet.denom1)
			},
		},
		{
			name: "not_fillable_orders_cancelled_right_after_creation",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 10_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// not fillable since 10_000 * 376e-5 ~= 37.6
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id2",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("333e-5")),
						// not fillable since 10_000*333e-5 = 33.3
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 10_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
					),
				}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},

		// TODO: Not sure if this is correct behavior but that is how it works now.
		{
			name: "partially_fillable_orders_accepted_for_creation",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 20_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:    testSet.acc1.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// not fully fillable since 20_000 * 376e-5 = 75.2, but 12_500 * 376e-5 = 47
						Quantity:    sdkmath.NewInt(20_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						Quantity:                  sdkmath.NewInt(20_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(20_000),
						RemainingSpendableBalance: sdkmath.NewInt(20_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
			wantExpectedToReceiveBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestKeeper_MatchOrders_DirectOBLimitMatching(t *testing.T) {
	tests := []tst{
		// ******************** Direct OB limit matching ********************
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3_761_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc2.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(10_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(3_384_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 375_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
						sdk.NewInt64Coin(testSet.denom2, 2_000), // 1000 - unused balance & 1000 - because executed for better price
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_maker_same_account",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
						sdk.NewInt64Coin(testSet.denom2, 3_761_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(10_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(9_000_000),
						RemainingSpendableBalance: sdkmath.NewInt(3_384_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
						sdk.NewInt64Coin(testSet.denom2, 377_000),
					),
				}
			},
		},
		{
			name: "try_to_match_limit_directOB_maker_sell_taker_buy_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 3_758_000),
						testSet.orderReserve,
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			// we fill the id1 first, so the used balance from id2 is 1_000_000 * 375e-3 = 375_000
			// to fill remaining part we need (10_000_000 - 1_000_000) * 376e-3 = 3_384_000,
			// so total expected to send 3_384_00 + 375_000 = 3_759_000
			wantErrorContains: func(testSet TestSet) string {
				return fmt.Sprintf("3759000%s is not available, available 3758000%s", testSet.denom2, testSet.denom2)
			},
		},

		// TODO: Figure out behavior in this case. Because we can possibly create an order,
		// but it seems to violate quote quantity step rule.
		// {
		// 	name: "match_limit_directOB_maker_sell_taker_buy_close_maker_with_partial_filling",
		// 	balances: func(testSet TestSet) map[string]sdk.Coins {
		// 		return map[string]sdk.Coins{
		// 			testSet.acc1.String(): sdk.NewCoins(
		// 				testSet.orderReserve,
		// 				sdk.NewInt64Coin(testSet.denom1, 1005),
		// 			),
		// 			testSet.acc2.String(): sdk.NewCoins(
		// 				testSet.orderReserve,
		// 				sdk.NewInt64Coin(testSet.denom2, 3760),
		// 			),
		// 		}
		// 	},
		// 	orders: func(testSet TestSet) []types.Order {
		// 		return []types.Order{
		// 			{
		// 				Creator:    testSet.acc1.String(),
		// 				Type:       types.ORDER_TYPE_LIMIT,
		// 				ID:         "id1",
		// 				BaseDenom:  testSet.denom1,
		// 				QuoteDenom: testSet.denom2,
		// 				Price:      lo.ToPtr(types.MustNewPriceFromString("375e-3")),
		// 				// only 1000 will be filled
		// 				Quantity:    sdkmath.NewInt(1005),
		// 				Side:        types.SIDE_SELL,
		// 				TimeInForce: types.TIME_IN_FORCE_GTC,
		// 			},
		// 			{
		// 				Creator:     testSet.acc2.String(),
		// 				Type:        types.ORDER_TYPE_LIMIT,
		// 				ID:          "id2",
		// 				BaseDenom:   testSet.denom1,
		// 				QuoteDenom:  testSet.denom2,
		// 				Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
		// 				Quantity:    sdkmath.NewInt(10000),
		// 				Side:        types.SIDE_BUY,
		// 				TimeInForce: types.TIME_IN_FORCE_GTC,
		// 			},
		// 		}
		// 	},
		// 	wantOrders: func(testSet TestSet) []types.Order {
		// 		return []types.Order{
		// 			{
		// 				Creator:     testSet.acc2.String(),
		// 				Type:        types.ORDER_TYPE_LIMIT,
		// 				ID:          "id2",
		// 				BaseDenom:   testSet.denom1,
		// 				QuoteDenom:  testSet.denom2,
		// 				Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
		// 				Quantity:    sdkmath.NewInt(10000),
		// 				Side:        types.SIDE_BUY,
		// 				TimeInForce: types.TIME_IN_FORCE_GTC,
		// 				// 10000 - 1000
		// 				RemainingBaseQuantity: sdkmath.NewInt(9000),
		// 				// (10000 - 1000) * 376e-3 = 3384
		// 				RemainingSpendableBalance: sdkmath.NewInt(3384),
		// 			},
		// 		}
		// 	},
		// 	wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
		// 		return map[string]sdk.Coins{
		// 			testSet.acc1.String(): sdk.NewCoins(
		// 				testSet.orderReserve,
		// 				sdk.NewInt64Coin(testSet.denom1, 5),
		// 				sdk.NewInt64Coin(testSet.denom2, 375),
		// 			),
		// 			testSet.acc2.String(): sdk.NewCoins(
		// 				sdk.NewInt64Coin(testSet.denom1, 1000),
		// 				sdk.NewInt64Coin(testSet.denom2, 1),
		// 			),
		// 		}
		// 	},
		// },
		{
			name: "match_limit_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3760),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 10000 - 1000
						RemainingBaseQuantity: sdkmath.NewInt(90_000),
						// 10000 - 1000
						RemainingSpendableBalance: sdkmath.NewInt(90_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 3750),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
						sdk.NewInt64Coin(testSet.denom2, 10),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_partially_fillable_taker_fully_cancelled",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 100),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("374e-5")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// not fully fillable since 20_000 * 376e-5 = 75.2, but 12_500 * 376e-5 = 47.
						// However with maker price it is not fillable at all
						Quantity:    sdkmath.NewInt(20_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("374e-5")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(100_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 100),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_sell_taker_buy_partially_matchable_taker_filled_fully",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 100),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-5")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:    testSet.acc2.String(),
						Type:       types.ORDER_TYPE_LIMIT,
						ID:         "id1",
						BaseDenom:  testSet.denom1,
						QuoteDenom: testSet.denom2,
						Price:      lo.ToPtr(types.MustNewPriceFromString("376e-5")),
						// not fully fillable since 20_000 * 376e-5 = 75.2, but 12_500 * 376e-5 = 47.
						// However, using maker price it is fillable 20_00 * 375e-5 = 75
						Quantity:    sdkmath.NewInt(20_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-5")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 100k - 20k
						RemainingBaseQuantity: sdkmath.NewInt(80_000),
						// 100k - 20k
						RemainingSpendableBalance: sdkmath.NewInt(80_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 75)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 20_000),
						sdk.NewInt64Coin(testSet.denom2, 100-75),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_maker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3760),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 100k - 10k
						RemainingBaseQuantity: sdkmath.NewInt(90_000),
						// 100k - 10k
						RemainingSpendableBalance: sdkmath.NewInt(90_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
					),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 3760)),
				}
			},
		},
		{
			name: "try_to_match_limit_directOB_maker_buy_taker_sell_insufficient_funds",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3_760),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 99_999),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantErrorContains: func(testSet TestSet) string {
				return fmt.Sprintf("100000%s is not available, available 99999%s", testSet.denom1, testSet.denom1)
			},
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 37_600)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 100k - 10k
						RemainingBaseQuantity: sdkmath.NewInt(90_000),
						// 376e-3 * 90_000 = 33840
						RemainingSpendableBalance: sdkmath.NewInt(33_840),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 10_000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3_760),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_maker_buy_taker_sell_close_taker_with_same_price",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 37_500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 10_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// 100k - 10k
						RemainingBaseQuantity: sdkmath.NewInt(90_000),
						// 375e-3 * 10000 - 375e-3 * 1000 = 3375
						RemainingSpendableBalance: sdkmath.NewInt(33_750),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 10_000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 3750),
					),
				}
			},
		},
		{
			// STOPPED HERE.
			name: "match_limit_directOB_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000)),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 50_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 50_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_close_two_makers_sell_and_taker_buy_with_remainder",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 50_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 50_000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 60_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(50_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" will match the "id1" and "id2" cover them fully and the remainder will be returned
					//	to the creator's balance
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("6e-1")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 25_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 25_000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
						sdk.NewInt64Coin(testSet.denom2, 10_000),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_close_two_makers_buy_and_and_taker_sell",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 50_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 50_000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 200_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					// "id1" and "id2" orders don't match
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("5e-1")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// "id3" closes "id1" and "id2", with better price for the "id3", expected to receive 80, but receive 100
					{
						Creator:     testSet.acc3.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("4e-1")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc3.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 100_000),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_multiple_maker_buy_taker_sell_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom2, 75_400+75_200+71_000+75_200),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 500_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("377e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too low
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("355e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain. Order sequence respected.
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("37e-2")),
						Quantity:    sdkmath.NewInt(500_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("355e-3")),
						Quantity:                  sdkmath.NewInt(200_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(200_000),
						RemainingSpendableBalance: sdkmath.NewInt(71_000),
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
						// partially executed
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(37_600),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 500_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						// executed id1: 200k*0.377 id2: 200k*0.376 and id4(partially): 100k*0.376
						sdk.NewInt64Coin(testSet.denom2, 188_200),
					),
				}
			},
		},
		{
			name: "match_limit_directOB_multiple_maker_sell_taker_buy_close_taker_with_same_price_fifo_priority",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom1, 200_000+200_000+100_000+200_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 189_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// remains unmatched price is too high
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("399e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the part of the order should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id5",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("378e-3")),
						Quantity:    sdkmath.NewInt(500_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id3",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("399e-3")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(100_000),
					},

					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(200_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(100_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom2, 187_800),
					),
					testSet.acc2.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 500_000),
						sdk.NewInt64Coin(testSet.denom2, 1_200),
					),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestKeeper_MatchOrders_DirectOBMarketMatching(t *testing.T) {
	tests := []tst{
		{
			name: "match_market_directOB_maker_sell_taker_buy_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(testSet.denom2, 375_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 37_500),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 100_000),
						// Locked full balance and returned remainer: 3750 - 375 = 3375
						sdk.NewInt64Coin(testSet.denom2, 337_500),
					),
				}
			},
		},
		{
			name: "match_market_directOB_multiple_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(4),
						sdk.NewInt64Coin(testSet.denom1, 4*100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						// no reserve is needed for market
						sdk.NewInt64Coin(testSet.denom2, 37_500+55_500+77_700),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("555e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// should remain
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id4",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id5",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(300_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id4",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("777e-3")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(100_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(3),
						sdk.NewInt64Coin(testSet.denom2, 37_500+55_500+77_700),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 300_000),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_sell_taker_buy_close_with_no_change_zero_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 110_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					// the order will be placed but, since it cannot be matched, it will be executed with no state change
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(100_000),
						RemainingSpendableBalance: sdkmath.NewInt(100_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 10_000)), // 10k locked by the order
				}
			},
		},
		{
			// TODO(v6): Revise this behavior.
			name: "match_market_directOB_maker_sell_taker_buy_with_partially_filling",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserveTimes(2),
						sdk.NewInt64Coin(testSet.denom1, 200_000),
					),
					// the account has coins to cover one order fully and 100 is not enough for 2nd one,
					// also not reserve is needed for the market order
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 37_500+37_500),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id3",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(200_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id2",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(375),
						RemainingSpendableBalance: sdkmath.NewInt(375),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 37_500+37_459),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 199_625),
						sdk.NewInt64Coin(testSet.denom2, 41),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_buy_taker_sell_close_both",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 376_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 1_000_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
						sdk.NewInt64Coin(testSet.denom2, 338_400),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 900_000),
						sdk.NewInt64Coin(testSet.denom2, 37_600),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_buy_taker_sell_close_both_with_taker_partial_filling_lack_of_balance",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom2, 376_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 990_000),
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(1_000_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("376e-3")),
						Quantity:                  sdkmath.NewInt(1_000_000),
						Side:                      types.SIDE_BUY,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(10_000),
						RemainingSpendableBalance: sdkmath.NewInt(3_760),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom1, 990_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 372_240),
					),
				}
			},
		},
		{
			name: "match_market_directOB_maker_sell_taker_buy_close_taker",
			balances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(
						testSet.orderReserve,
						sdk.NewInt64Coin(testSet.denom1, 100_000),
					),
					testSet.acc2.String(): sdk.NewCoins(
						sdk.NewInt64Coin(testSet.denom2, 3_750+9_000), // 3_750 should be spent
					),
				}
			},
			orders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:     testSet.acc1.String(),
						Type:        types.ORDER_TYPE_LIMIT,
						ID:          "id1",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Price:       lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:    sdkmath.NewInt(100_000),
						Side:        types.SIDE_SELL,
						TimeInForce: types.TIME_IN_FORCE_GTC,
					},
					{
						Creator:     testSet.acc2.String(),
						Type:        types.ORDER_TYPE_MARKET,
						ID:          "id2",
						BaseDenom:   testSet.denom1,
						QuoteDenom:  testSet.denom2,
						Quantity:    sdkmath.NewInt(10_000),
						Side:        types.SIDE_BUY,
						TimeInForce: types.TIME_IN_FORCE_IOC,
					},
				}
			},
			wantOrders: func(testSet TestSet) []types.Order {
				return []types.Order{
					{
						Creator:                   testSet.acc1.String(),
						Type:                      types.ORDER_TYPE_LIMIT,
						ID:                        "id1",
						BaseDenom:                 testSet.denom1,
						QuoteDenom:                testSet.denom2,
						Price:                     lo.ToPtr(types.MustNewPriceFromString("375e-3")),
						Quantity:                  sdkmath.NewInt(100_000),
						Side:                      types.SIDE_SELL,
						TimeInForce:               types.TIME_IN_FORCE_GTC,
						RemainingBaseQuantity:     sdkmath.NewInt(90_000),
						RemainingSpendableBalance: sdkmath.NewInt(90_000),
					},
				}
			},
			wantAvailableBalances: func(testSet TestSet) map[string]sdk.Coins {
				return map[string]sdk.Coins{
					testSet.acc1.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom2, 3_750)),
					testSet.acc2.String(): sdk.NewCoins(sdk.NewInt64Coin(testSet.denom1, 10_000), sdk.NewInt64Coin(testSet.denom2, 9_000)),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func genTestSet(t require.TestingT, sdkCtx sdk.Context, testApp *simapp.App) TestSet {
	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)
	acc3, _ := testApp.GenAccount(sdkCtx)

	issuer, _ := testApp.GenAccount(sdkCtx)

	denoms := make([]string, 0)
	for _, subunits := range []string{"denom1", "denom2", "denom3"} {
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
			Issuer:        issuer,
			Subunit:       subunits,
			Symbol:        strings.ToUpper(subunits),
			Precision:     6,
			InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		})
		require.NoError(t, err)
		denoms = append(denoms, denom)
	}

	ftDenomWhitelisting1, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
		Issuer:        issuer,
		Subunit:       "ftwhitelisting1",
		Symbol:        "FTWHITELISTING1",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	})
	require.NoError(t, err)

	ftDenomWhitelisting2, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
		Issuer:        issuer,
		Subunit:       "ftwhitelisting2",
		Symbol:        "FTWHITELISTING2",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	})
	require.NoError(t, err)

	param, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)

	testSet := TestSet{
		acc1: acc1,
		acc2: acc2,
		acc3: acc3,

		issuer:               issuer,
		denom1:               denoms[0],
		denom2:               denoms[1],
		denom3:               denoms[2],
		ftDenomWhitelisting1: ftDenomWhitelisting1,
		ftDenomWhitelisting2: ftDenomWhitelisting2,

		orderReserve: param.OrderReserve,
	}

	return testSet
}

func removeEmptyBalances(balances map[string]sdk.Coins) map[string]sdk.Coins {
	for addr, balance := range balances {
		if balance.IsZero() {
			delete(balances, addr)
		}
	}

	return balances
}

func fillReserveAndOrderSequence(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	orders []types.Order,
) []types.Order {
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	orderReserve := params.OrderReserve
	for i, order := range orders {
		storedOrder, err := testApp.DEXKeeper.GetOrderByAddressAndID(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID,
		)
		require.NoError(t, err)
		require.Positive(t, storedOrder.Sequence)
		orders[i].Sequence = storedOrder.Sequence
		orders[i].Reserve = orderReserve
	}

	return orders
}

func assertOrderPlacementResult(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	availableBalancesBefore map[string]sdkmath.Int,
	order types.Order,
) {
	events := readOrderEvents(t, sdkCtx)
	assertPlacementEvents(t, order, events)

	sentAmt, receivedAmt := assetOrderSentReceivedAmounts(
		t, sdkCtx, testApp, availableBalancesBefore, order, events,
	)
	storedOrder, err := testApp.DEXKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID,
	)
	if err != nil {
		require.ErrorIs(t, err, types.ErrRecordNotFound)
		t.Logf("Order not found in the order book.")
		assertFilledQuantity(t, order, sentAmt, receivedAmt)
		return
	}

	t.Logf("Order found in the order book.")
	require.NotNil(t, events.OrderCreated)
	require.Equal(t, types.EventOrderCreated{
		Creator:                   storedOrder.Creator,
		ID:                        storedOrder.ID,
		Sequence:                  events.OrderPlaced.Sequence,
		RemainingBaseQuantity:     storedOrder.RemainingBaseQuantity,
		RemainingSpendableBalance: storedOrder.RemainingSpendableBalance,
	}, *events.OrderCreated)

	if order.Type != types.ORDER_TYPE_LIMIT {
		t.Fatalf("Saved not market order, type: %s", order.Type.String())
	}
	if order.TimeInForce != types.TIME_IN_FORCE_GTC {
		t.Fatalf("Saved not GTC order, time in force: %s", order.TimeInForce.String())
	}
}

func assertPlacementEvents(t *testing.T, order types.Order, events OrderPlacementEvents) {
	require.Positive(t, events.OrderPlaced.Sequence)
	require.Equal(t, types.EventOrderPlaced{
		Creator:  order.Creator,
		ID:       order.ID,
		Sequence: events.OrderPlaced.Sequence,
	}, events.OrderPlaced)

	// set initial quantity
	expectedRemainingSpendQuantity := order.Quantity

	takerSentAmt := sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt())
	takerReceivedAmt := sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt())
	makerSentAmt := sdk.NewCoin(order.GetReceiveDenom(), sdkmath.ZeroInt())
	makerReceivedAmt := sdk.NewCoin(order.GetSpendDenom(), sdkmath.ZeroInt())
	for _, reducedEvt := range events.OrdersReduced {
		// is taker
		if events.OrderPlaced.Sequence == reducedEvt.Sequence {
			require.True(t, takerSentAmt.IsZero())
			require.True(t, takerReceivedAmt.IsZero())
			takerSentAmt = reducedEvt.SentCoin
			takerReceivedAmt = reducedEvt.ReceivedCoin

			if order.Side == types.SIDE_BUY {
				expectedRemainingSpendQuantity = expectedRemainingSpendQuantity.Sub(reducedEvt.ReceivedCoin.Amount)
			} else {
				expectedRemainingSpendQuantity = expectedRemainingSpendQuantity.Sub(reducedEvt.SentCoin.Amount)
			}
			continue
		}
		makerSentAmt = makerSentAmt.Add(reducedEvt.SentCoin)
		makerReceivedAmt = makerReceivedAmt.Add(reducedEvt.ReceivedCoin)
	}
	require.Equal(t, takerSentAmt.String(), makerReceivedAmt.String())
	require.Equal(t, makerSentAmt.String(), takerReceivedAmt.String())
	if events.OrderCreated != nil {
		expectedRemainingBalance, err := types.ComputeLimitOrderLockedBalance(
			order.Side, order.BaseDenom, order.QuoteDenom, expectedRemainingSpendQuantity, *order.Price,
		)
		require.NoError(t, err)
		require.Equal(t, types.EventOrderCreated{
			Creator:                   order.Creator,
			ID:                        order.ID,
			Sequence:                  events.OrderPlaced.Sequence,
			RemainingBaseQuantity:     expectedRemainingSpendQuantity,
			RemainingSpendableBalance: expectedRemainingBalance.Amount,
		}, *events.OrderCreated)
	}
}

func assetOrderSentReceivedAmounts(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
	availableBalancesBefore map[string]sdkmath.Int,
	order types.Order,
	events OrderPlacementEvents,
) (sdkmath.Int, sdkmath.Int) {
	var (
		orderSentAmt     = sdkmath.ZeroInt()
		orderReceivedAmt = sdkmath.ZeroInt()
	)
	if len(events.OrdersReduced) > 0 {
		require.GreaterOrEqual(t, len(events.OrdersReduced), 2)
		currentOrderReducedEvent, ok := events.getOrderReduced(order.Creator, order.ID)
		require.True(t, ok)
		matchedSent := sdk.NewCoin(currentOrderReducedEvent.ReceivedCoin.Denom, sdkmath.ZeroInt())
		matchedReceived := sdk.NewCoin(currentOrderReducedEvent.SentCoin.Denom, sdkmath.ZeroInt())
		for _, evt := range events.OrdersReduced {
			if evt.Creator == order.Creator && evt.ID == order.ID {
				continue
			}
			matchedSent = matchedSent.Add(evt.SentCoin)
			matchedReceived = matchedReceived.Add(evt.ReceivedCoin)
		}
		require.Equal(t, currentOrderReducedEvent.SentCoin.String(), matchedReceived.String())
		require.Equal(t, currentOrderReducedEvent.ReceivedCoin.String(), matchedSent.String())

		orderSentAmt = currentOrderReducedEvent.SentCoin.Amount
		orderReceivedAmt = currentOrderReducedEvent.ReceivedCoin.Amount

		if order.Type == types.ORDER_TYPE_LIMIT {
			assertExecutionPrice(
				t,
				order,
				orderSentAmt,
				orderReceivedAmt,
			)
		}
	}

	// check that balance used amount either sent or locked in the order
	orderUsedAmt := orderSentAmt
	if events.OrderCreated != nil {
		// locked balance
		orderUsedAmt = orderUsedAmt.Add(events.OrderCreated.RemainingSpendableBalance)
	}

	creator := sdk.MustAccAddressFromBech32(order.Creator)
	// check the balances updated
	availableAmtBefore, ok := availableBalancesBefore[order.GetSpendDenom()]
	if !ok {
		availableAmtBefore = sdkmath.ZeroInt()
	}
	availableBalancesAfter, err := getAvailableBalances(sdkCtx, testApp, creator)
	require.NoError(t, err)
	availableBalanceAmtAfter, ok := availableBalancesAfter[order.GetSpendDenom()]
	if !ok {
		availableBalanceAmtAfter = sdkmath.ZeroInt()
	}
	// balanceUsedAmt includes locked and frozen balances
	balanceUsedAmt := availableAmtBefore.Sub(availableBalanceAmtAfter)
	require.False(t, balanceUsedAmt.IsNegative())

	// adjust with the direct OB matched orders
	for _, evt := range events.OrdersReduced {
		if evt.Creator != order.Creator {
			continue
		}
		if evt.ID == order.ID {
			continue
		}
		if evt.SentCoin.Denom != order.GetSpendDenom() {
			balanceUsedAmt = balanceUsedAmt.Add(evt.ReceivedCoin.Amount)
		} else {
			// we can't same coin as `order.GetSpendDenom()` with the direct OB match
			t.Fatalf("Unexpected to sent coin: %s", evt.SentCoin)
		}
	}

	require.Equal(t, balanceUsedAmt.String(), orderUsedAmt.String())

	return orderSentAmt, orderReceivedAmt
}

func assertExecutionPrice(t *testing.T, order types.Order, spendAmt, receiveAmt sdkmath.Int) {
	orderPriceRat := order.Price.Rat()
	var executionPriceRat *big.Rat
	if order.Side == types.SIDE_BUY {
		if receiveAmt.IsZero() {
			return
		}
		executionPriceRat = cbig.NewRatFromBigInts(spendAmt.BigInt(), receiveAmt.BigInt())
		require.True(
			t,
			cbig.RatLTE(executionPriceRat, orderPriceRat),
			"orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String(),
		)
	} else {
		if spendAmt.IsZero() {
			return
		}
		executionPriceRat = cbig.NewRatFromBigInts(receiveAmt.BigInt(), spendAmt.BigInt())
		require.True(
			t,
			cbig.RatGTE(executionPriceRat, orderPriceRat),
			"orderPrice: %s, executionPrice: %s", orderPriceRat.String(), executionPriceRat.String(),
		)
	}
	t.Logf(
		"Execution prices, side:%s, order: %s, execution: %s, ",
		order.Side.String(), orderPriceRat.String(), executionPriceRat.String(),
	)
}

func assertFilledQuantity(t *testing.T, order types.Order, sent, receiveAmt sdkmath.Int) {
	var filledQuantity sdkmath.Int
	if order.Side == types.SIDE_BUY {
		filledQuantity = receiveAmt
	} else {
		filledQuantity = sent
	}
	t.Logf(
		"Filled qunitities, side:%s, orderTimeInForce: %s, orderQuantity: %s, filledQuantity: %s",
		order.Side.String(), order.TimeInForce.String(), order.Quantity.String(), filledQuantity.String(),
	)
	// check that we never exceed the order's quantity
	require.True(t, order.Quantity.GTE(filledQuantity))
}

func getAvailableBalances(sdkCtx sdk.Context, testApp *simapp.App, acc sdk.AccAddress) (map[string]sdkmath.Int, error) {
	balances := testApp.BankKeeper.GetAllBalances(sdkCtx, acc)
	spendableBalances := make(map[string]sdkmath.Int)
	for _, balance := range balances {
		frozenBalance, err := testApp.AssetFTKeeper.GetFrozenBalance(sdkCtx, acc, balance.Denom)
		if err != nil {
			return nil, err
		}
		frozenAmt := frozenBalance.Amount
		dexLockedAmt := testApp.AssetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, balance.Denom).Amount
		// can be negative
		spendableBalances[balance.Denom] = balance.Amount.Sub(frozenAmt).Sub(dexLockedAmt)
	}

	return spendableBalances, nil
}

func cancelAllOrdersAndAssertState(
	t *testing.T,
	sdkCtx sdk.Context,
	testApp *simapp.App,
) {
	t.Helper()

	orders, _, err := testApp.DEXKeeper.GetAccountsOrders(sdkCtx, &query.PageRequest{Limit: query.PaginationMaxLimit})
	require.NoError(t, err)

	t.Logf("Cancelling all orders, count %d", len(orders))

	accounts := make(map[string]struct{})
	denoms := make(map[string]struct{})
	for _, order := range orders {
		require.NoError(t, testApp.DEXKeeper.CancelOrder(
			sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID),
		)
		accounts[order.Creator] = struct{}{}
		denoms[order.BaseDenom] = struct{}{}
		denoms[order.QuoteDenom] = struct{}{}
	}
	for acc := range accounts {
		for denom := range denoms {
			dexLockedBalance := testApp.AssetFTKeeper.GetDEXLockedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexLockedBalance.IsZero(),
				"denom: %s, acc: %s, dexLockedBalance: %s", denom, acc, dexLockedBalance.String(),
			)

			dexExpectedToReceiveBalance := testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.True(
				t, dexExpectedToReceiveBalance.IsZero(),
				"denom: %s, acc: %s, dexExpectedToReceiveBalance: %s",
				denom, acc, dexExpectedToReceiveBalance.String(),
			)

			accountDenomOrdersCount, err := testApp.DEXKeeper.GetAccountDenomOrdersCount(
				sdkCtx, sdk.MustAccAddressFromBech32(acc), denom,
			)
			require.NoError(t, err)
			require.Zero(t, accountDenomOrdersCount)
		}
	}
}
