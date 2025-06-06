package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cucumber/godog"
	cucumbermessage "github.com/cucumber/messages/go/v21"
	"github.com/pkg/errors"
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
	var tc testScenario
	sc.Given(`^there are users with balances:$`, tc.thereAreUsersWithBalances)
	sc.Given(`^there are orders:`, tc.thereAreOrders)
	sc.Then(`^no orders are matched$`, tc.noOrdersAreMatched)

	// prepareScenario()
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// tc.testApp = init()
		return context.WithValue(ctx, "key", "myval"), nil
	})
}

func (a *testScenario) thereAreUsersWithBalances(ctx context.Context, users *godog.Table) error {
	fmt.Println("thereAreUsersWithBalances")
	var err error
	a.balances, err = parseUserbalanceTalbe(users)
	return err
}

func (a *testScenario) thereAreOrders(orders *godog.Table) error {
	var err error
	a.orders, err = parseOrdersTalbe(orders)
	return err
}

func (a *testScenario) noOrdersAreMatched() error {
	return nil
}

func parseUserbalanceTalbe(users *godog.Table) (map[string]sdk.Coins, error) {
	assertTableHeader(users.Rows[0], []string{"account", "balances"})
	balances := make(map[string]sdk.Coins)
	for _, cells := range users.Rows[1:] {
		coins, err := sdk.ParseCoinsNormalized(cells.Cells[1].Value)
		if err != nil {
			return nil, errors.Errorf("error parsing coins %s", cells.Cells[1].Value)
		}
		balances[cells.Cells[0].Value] = coins
	}
	return balances, nil
}

func parseOrdersTalbe(users *godog.Table) ([]types.Order, error) {
	orders := make([]types.Order, 0)
	if err := assertTableHeader(users.Rows[0], []string{
		"id", "creator", "base", "quote", "type", "price", "quantity", "side", "tif",
	}); err != nil {
		return nil, err
	}

	for _, cells := range users.Rows[1:] {
		price, err := types.NewPriceFromString(cells.Cells[6].Value)
		if err != nil {
			return nil, err
		}
		order := types.Order{
			ID:         cells.Cells[0].Value,
			Creator:    cells.Cells[1].Value,
			BaseDenom:  cells.Cells[2].Value,
			QuoteDenom: cells.Cells[3].Value,
			Type:       parseOrderType(cells.Cells[5].Value),
			Price:      &price,
		}
		orders = append(orders, order)
	}
	return orders, nil
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

func assertTableHeader(row *cucumbermessage.PickleTableRow, headers []string) error {
	for i, header := range headers {
		if row.Cells[i].Value != header {
			return errors.Errorf("row %d should be %s but it was %s", i, header, row.Cells[i].Value)
		}
	}
	return nil
}
