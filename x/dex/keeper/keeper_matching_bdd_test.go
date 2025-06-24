package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cucumber/godog"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/bdd"
	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var ts testScenario

	sc.Given(`^there are users with balances:$`, ts.thereAreUsersWithBalances)
	sc.Given(`^there are orders:$`, ts.thereAreOrders)
	sc.Then(`^no orders are matched$`, ts.noOrdersAreMatched)
	sc.Then(`^there will be no available balances$`, ts.thereWillBeNoAvailableBalances)
	sc.Then(`^there will be remaining orders:$`, ts.thereWillBeRemainingOrders)
	sc.Then(`^there will be no remaining orders$`, ts.thereWillBeNoRemainingOrders)
	sc.Then(`^expecting error that contains:$`, ts.expectingErrorThatContains)
	sc.Then(`^there will be users with balances:$`, ts.thereWillBeUsersWithBalances)

	setupScenarioHooks(sc, &ts)
}

func setupScenarioHooks(sc *godog.ScenarioContext, ts *testScenario) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		t := godog.T(ctx)

		ts.logger = log.NewTestLoggerError(bdd.NewLogWrapper(t))
		ts.testApp = simapp.New(simapp.WithCustomLogger(ts.logger))
		ts.sdkCtx = ts.testApp.NewContext(false)

		param, err := ts.testApp.DEXKeeper.GetParams(ts.sdkCtx)
		require.NoError(t, err)
		ts.orderReserve = param.OrderReserve

		return ctx, nil
	})
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err != nil {
			return ctx, err
		}
		ts.run(ctx)
		return ctx, nil
	})
	// TODO: We can use step hook to fill scenario or run the test before "then" steps (it should be idempotent)
	//
	//sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
	//	if st.Type == messages.PickleStepType_OUTCOME {
	//		t := godog.T(ctx)
	//		fillTestScenario(t, ts.sdkCtx, ts.testApp, ts)
	//	}
	//	return ctx, nil
	//})
}

func (ts *testScenario) thereAreUsersWithBalances(ctx context.Context, users *godog.Table) (err error) {
	ts.balances, err = parseUserBalanceTable(users)
	return err
}

func (ts *testScenario) thereWillBeUsersWithBalances(ctx context.Context, users *godog.Table) (err error) {
	ts.wantAvailableBalances, err = parseUserBalanceTable(users)
	return err
}

func (ts *testScenario) thereAreOrders(ctx context.Context, orders *godog.Table) (err error) {
	ts.orders, err = parseOrdersTable(orders)
	return err
}

func (ts *testScenario) noOrdersAreMatched(ctx context.Context) (err error) {
	// TODO
	return nil
}

func (ts *testScenario) thereWillBeRemainingOrders(ctx context.Context, orders *godog.Table) (err error) {
	ts.wantOrders, err = parseOrdersTable(orders)
	return err
}

func (ts *testScenario) thereWillBeNoRemainingOrders(ctx context.Context) (err error) {
	ts.wantOrders = make([]types.Order, 0)
	return nil
}

func (ts *testScenario) thereWillBeNoAvailableBalances(ctx context.Context) (err error) {
	ts.wantAvailableBalances = make(map[string]sdk.Coins)
	return nil
}

func (ts *testScenario) expectingErrorThatContains(ctx context.Context, expectedErr string) (err error) {
	ts.wantErrorContains = expectedErr
	return nil
}

func parseUserBalanceTable(usersTable *godog.Table) (map[string]sdk.Coins, error) {
	users := bdd.ParseTable(usersTable)
	balances := make(map[string]sdk.Coins)
	for i := range users.RowsCount {
		coins, err := sdk.ParseCoinsNormalized(users.Rows["balances"][i])
		if err != nil {
			return nil, errors.Errorf("error parsing coins %s", users.Rows["balances"][i])
		}
		balances[users.Rows["account"][i]] = coins
	}
	return balances, nil
}

func parseOrdersTable(ordersTable *godog.Table) ([]types.Order, error) {
	result := make([]types.Order, 0)
	orders := bdd.ParseTable(ordersTable)

	for i := range orders.RowsCount {
		orderType := parseOrderType(orders.Rows["type"][i])
		var price *types.Price
		if orderType == types.ORDER_TYPE_LIMIT {
			parsedPrice, err := types.NewPriceFromString(orders.Rows["price"][i])
			if err != nil {
				return nil, err
			}
			price = &parsedPrice
		}
		quantity, ok := sdkmath.NewIntFromString(orders.Rows["quantity"][i])
		if !ok {
			return nil, fmt.Errorf("error parsing quantity %s", orders.Rows["quantity"][i])
		}
		order := types.Order{
			ID:          orders.Rows["id"][i],
			Creator:     orders.Rows["creator"][i],
			BaseDenom:   orders.Rows["base"][i],
			QuoteDenom:  orders.Rows["quote"][i],
			Type:        orderType,
			Price:       price,
			Quantity:    quantity,
			Side:        parseSide(orders.Rows["side"][i]),
			TimeInForce: parseTimeInForce(orders.Rows["tif"][i]),
		}
		if remainingBaseQuantity, exists := orders.Rows["remaining quantity"]; exists {
			order.RemainingBaseQuantity, ok = sdkmath.NewIntFromString(remainingBaseQuantity[i])
			if !ok {
				return nil, fmt.Errorf("error parsing remaining base quantity %s", remainingBaseQuantity[i])
			}
		}
		if remainingSpendableBalance, exists := orders.Rows["remaining balance"]; exists {
			order.RemainingSpendableBalance, ok = sdkmath.NewIntFromString(remainingSpendableBalance[i])
			if !ok {
				return nil, fmt.Errorf("error parsing remaining spendable balance %s", remainingSpendableBalance[i])
			}
		}
		result = append(result, order)
	}
	return result, nil
}

func parseOrderType(input string) types.OrderType {
	switch input {
	case "limit":
		return types.ORDER_TYPE_LIMIT
	case "market":
		return types.ORDER_TYPE_MARKET
	default:
		return types.ORDER_TYPE_UNSPECIFIED
	}
}

func parseSide(input string) types.Side {
	switch input {
	case "sell":
		return types.SIDE_SELL
	case "buy":
		return types.SIDE_BUY
	default:
		return types.SIDE_UNSPECIFIED
	}
}

func parseTimeInForce(input string) types.TimeInForce {
	switch input {
	case "gtc":
		return types.TIME_IN_FORCE_GTC
	case "ioc":
		return types.TIME_IN_FORCE_IOC
	case "fok":
		return types.TIME_IN_FORCE_FOK
	default:
		return types.TIME_IN_FORCE_UNSPECIFIED
	}
}
