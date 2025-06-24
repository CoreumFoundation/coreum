package keeper_test

import (
	"context"
	"reflect"
	"sort"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

type testScenario struct {
	testApp *simapp.App
	sdkCtx  sdk.Context
	logger  log.Logger

	orderReserve sdk.Coin

	name string
	// given
	balances            map[string]sdk.Coins
	whitelistedBalances map[string]sdk.Coins
	orders              []types.Order
	// expected
	wantOrders                    []types.Order
	wantAvailableBalances         map[string]sdk.Coins
	wantExpectedToReceiveBalances map[string]sdk.Coins
	wantErrorContains             string
}

func fillTestScenario(t require.TestingT, sdkCtx sdk.Context, testApp *simapp.App, ts *testScenario) {
	issuer, _ := testApp.GenAccount(sdkCtx)

	// create map
	denoms := map[string]string{}
	accounts := map[string]string{}
	for acc, coins := range ts.balances {
		genAcc, _ := testApp.GenAccount(sdkCtx)
		accounts[acc] = genAcc.String()
		for _, coin := range coins {
			if coin.Denom != "orderReserve" {
				denoms[coin.Denom] = ""
			}
		}
	}
	for _, order := range ts.orders {
		// TODO: Should we also add order.Creator to accounts?
		denoms[order.BaseDenom] = ""
		denoms[order.QuoteDenom] = ""
	}
	for _, order := range ts.wantOrders {
		// TODO: Should we also add order.Creator to accounts?
		denoms[order.BaseDenom] = ""
		denoms[order.QuoteDenom] = ""
	}

	// issue denoms
	for symbolicName := range denoms {
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, assetfttypes.IssueSettings{
			Issuer:        issuer,
			Subunit:       symbolicName,
			Symbol:        strings.ToUpper(symbolicName),
			Precision:     6,
			InitialAmount: sdkmath.NewIntWithDecimal(1, 20),
		})
		require.NoError(t, err)
		denoms[symbolicName] = denom
	}

	// fill balances with real accounts and denoms
	filledBalances := map[string]sdk.Coins{}
	for acc, coins := range ts.balances {
		filledCoins := sdk.Coins{}
		for _, coin := range coins {
			if coin.Denom == "orderReserve" {
				filledCoins = filledCoins.Add(sdk.NewCoin(ts.orderReserve.Denom, ts.orderReserve.Amount.Mul(coin.Amount)))
				continue
			}
			filledCoins = filledCoins.Add(sdk.NewCoin(denoms[coin.Denom], coin.Amount))
		}
		filledBalances[accounts[acc]] = filledCoins
	}

	ts.balances = filledBalances

	// fill orders with real accounts and denoms
	var filledOrders []types.Order
	for _, order := range ts.orders {
		order.Creator = accounts[order.Creator]
		order.BaseDenom = denoms[order.BaseDenom]
		order.QuoteDenom = denoms[order.QuoteDenom]
		filledOrders = append(filledOrders, order)

		// add order reserve coins to the creator
		//ts.orders = filledOrders
		//filledCoins, exists := filledBalances[order.Creator]
		//if !exists {
		//	filledCoins = sdk.Coins{}
		//}
		//filledCoins.Add(sdk.NewCoin(ts.orderReserve.Denom, ts.orderReserve.Amount))
		//filledBalances[order.Creator] = filledCoins
	}

	//ts.balances = filledBalances
	ts.orders = filledOrders

	var filledWantedOrders []types.Order
	for _, order := range ts.wantOrders {
		order.Creator = accounts[order.Creator]
		order.BaseDenom = denoms[order.BaseDenom]
		order.QuoteDenom = denoms[order.QuoteDenom]
		filledWantedOrders = append(filledWantedOrders, order)
		ts.wantOrders = filledWantedOrders
	}

	filledBalances = map[string]sdk.Coins{}
	for acc, coins := range ts.wantAvailableBalances {
		filledCoins := sdk.Coins{}
		for _, coin := range coins {
			if coin.Denom == "orderReserve" {
				filledCoins = filledCoins.Add(sdk.NewCoin(ts.orderReserve.Denom, ts.orderReserve.Amount.Mul(coin.Amount)))
				continue
			}
			filledCoins = filledCoins.Add(sdk.NewCoin(denoms[coin.Denom], coin.Amount))
		}
		filledBalances[accounts[acc]] = filledCoins
	}

	ts.wantAvailableBalances = filledBalances

	if ts.wantErrorContains != "" {
		accs := lo.Keys(accounts)
		sort.SliceStable(accs, func(i, j int) bool {
			return len(accs[i]) > len(accs[j])
		})
		for _, acc := range accs {
			ts.wantErrorContains = strings.ReplaceAll(ts.wantErrorContains, acc, accounts[acc])
		}

		dens := lo.Keys(denoms)
		sort.SliceStable(dens, func(i, j int) bool {
			return len(dens[i]) > len(dens[j])
		})
		for _, denom := range dens {
			ts.wantErrorContains = strings.ReplaceAll(ts.wantErrorContains, denom, denoms[denom])
		}
	}
}

func (ts *testScenario) run(ctx context.Context) {
	t := godog.T(ctx)

	fillTestScenario(t, ts.sdkCtx, ts.testApp, ts)

	if ts.whitelistedBalances != nil {
		for addr, coins := range ts.whitelistedBalances {
			ts.testApp.AssetFTKeeper.SetWhitelistedBalances(ts.sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
		}
	}

	for addr, coins := range ts.balances {
		ts.testApp.MintAndSendCoin(t, ts.sdkCtx, sdk.MustAccAddressFromBech32(addr), coins)
	}

	orderBooksIDs := make(map[uint32]struct{})
	initialOrders := ts.orders

	ordersDenoms := make(map[string]struct{})
	for i, order := range initialOrders {
		ordersDenoms[order.BaseDenom] = struct{}{}
		ordersDenoms[order.QuoteDenom] = struct{}{}
		availableBalancesBefore, err := getAvailableBalances(ts.sdkCtx, ts.testApp, sdk.MustAccAddressFromBech32(order.Creator))
		require.NoError(t, err)

		// use a new event manager for each order
		ts.sdkCtx = ts.sdkCtx.WithEventManager(sdk.NewEventManager())
		gasBefore := ts.sdkCtx.GasMeter().GasConsumed()
		err = ts.testApp.DEXKeeper.PlaceOrder(ts.sdkCtx, order)
		if err != nil && ts.wantErrorContains != "" {
			require.True(t, sdkerrors.IsOf(
				err,
				assetfttypes.ErrDEXInsufficientSpendableBalance, assetfttypes.ErrWhitelistedLimitExceeded,
			))
			expectedErr := ts.wantErrorContains
			require.ErrorContains(t, err, expectedErr)
			return
		}
		gasAfter := ts.sdkCtx.GasMeter().GasConsumed()
		t.Logf("Used gas for order %d placement: %d", i, gasAfter-gasBefore)
		require.NoError(t, err)
		assertOrderPlacementResult(t, ts.sdkCtx, ts.testApp, availableBalancesBefore, order)
		orderBooksID, err := ts.testApp.DEXKeeper.GetOrderBookIDByDenoms(ts.sdkCtx, order.BaseDenom, order.QuoteDenom)
		require.NoError(t, err)
		orderBooksIDs[orderBooksID] = struct{}{}
	}
	if ts.wantErrorContains != "" {
		expectedErr := ts.wantErrorContains
		require.Failf(t, "expected error not found", expectedErr)
	}

	orders := make([]types.Order, 0)
	for orderBookID := range orderBooksIDs {
		orders = append(orders, getSorterOrderBookOrders(t, ts.testApp, ts.sdkCtx, orderBookID, types.SIDE_BUY)...)
		orders = append(orders, getSorterOrderBookOrders(t, ts.testApp, ts.sdkCtx, orderBookID, types.SIDE_SELL)...)
	}
	wantOrders := ts.wantOrders
	// set order reserve and order sequence for all orders
	wantOrders = fillReserveAndOrderSequence(t, ts.sdkCtx, ts.testApp, wantOrders)
	require.ElementsMatch(t, wantOrders, orders, "orders do not match: \n%s", cmp.Diff(wantOrders, orders))

	availableBalances := make(map[string]sdk.Coins)
	lockedBalances := make(map[string]sdk.Coins)
	expectedToReceiveBalances := make(map[string]sdk.Coins)
	for addr := range ts.balances {
		addrBalances := ts.testApp.BankKeeper.GetAllBalances(ts.sdkCtx, sdk.MustAccAddressFromBech32(addr))
		addrFTLockedBalances := sdk.NewCoins()
		for _, balance := range addrBalances {
			lockedBalance := ts.testApp.AssetFTKeeper.GetDEXLockedBalance(
				ts.sdkCtx, sdk.MustAccAddressFromBech32(addr), balance.Denom,
			)
			addrFTLockedBalances = addrFTLockedBalances.Add(lockedBalance)
			addrBalances = addrBalances.Sub(lockedBalance)
		}

		addrFTExpectedToReceiveBalances := sdk.NewCoins()
		for denom := range ordersDenoms {
			addrFTExpectedToReceiveBalance := ts.testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				ts.sdkCtx, sdk.MustAccAddressFromBech32(addr), denom,
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

	wantAvailableBalances := ts.wantAvailableBalances
	require.True(
		t,
		reflect.DeepEqual(wantAvailableBalances, availableBalances),
		"available balances do not match: %v", cmp.Diff(wantAvailableBalances, availableBalances),
	)

	// by default, must be empty
	wantExpectedToReceiveBalances := make(map[string]sdk.Coins)
	if ts.wantExpectedToReceiveBalances != nil {
		wantExpectedToReceiveBalances = ts.wantExpectedToReceiveBalances
	}

	require.True(
		t,
		reflect.DeepEqual(wantExpectedToReceiveBalances, expectedToReceiveBalances),
		"expected to receive balances do not match: %v", cmp.Diff(wantAvailableBalances, availableBalances),
	)

	// check that the balance locked in the orders correspond the balance locked in the asset ft
	orderLockedBalances := make(map[string]sdk.Coins)
	for _, order := range orders {
		coins, ok := orderLockedBalances[order.Creator]
		if !ok {
			coins = sdk.NewCoins()
		}
		coins = coins.Add(sdk.NewCoin(order.GetSpendDenom(), order.RemainingSpendableBalance))
		params, err := ts.testApp.DEXKeeper.GetParams(ts.sdkCtx)
		require.NoError(t, err)
		// add order reserve for each order
		coins = coins.Add(params.OrderReserve)
		orderLockedBalances[order.Creator] = coins
	}
	orderLockedBalances = removeEmptyBalances(orderLockedBalances)
	require.True(
		t,
		reflect.DeepEqual(lockedBalances, orderLockedBalances),
		"want: %v, got: %v", lockedBalances, orderLockedBalances,
	)

	cancelAllOrdersAndAssertState(t, ts.sdkCtx, ts.testApp)
}

// maybe a better name
type ContextProcessor struct {
	simApp     simapp.App
	accountMap map[string]sdk.Address
	denomMap   map[string]string
}

func (cp ContextProcessor) EnsureDenom(string) string           { return "" }
func (cp ContextProcessor) EnsureAccount(string) sdk.AccAddress { return nil }

func (cp ContextProcessor) issueDenom()       {}
func (cp ContextProcessor) mintAndSendCoins() {}

func (cp ContextProcessor) whitelistDenom() {} // does this belong here?
