package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	"github.com/CoreumFoundation/coreum/v4/x/dex/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func TestCmdQueryOrderBooksAndOrders(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, sdkmath.NewInt(1000))

	creator := validator1Address(testNetwork)
	order1 := types.Order{
		Creator:           creator.String(),
		ID:                "id1",
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             types.MustNewPriceFromString("123e-2"),
		Quantity:          sdkmath.NewInt(100),
		Side:              types.Side_sell,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
	}
	placeOrder(ctx, requireT, testNetwork, order1)

	// check single order
	var orderRes types.QueryOrderResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx, cli.CmdQueryOrder(), []string{creator.String(), order1.ID}, &orderRes,
	))
	requireT.Equal(order1, orderRes.Order)

	// check order books
	var orderBooksRes types.QueryOrderBooksResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryOrderBooks(), []string{}, &orderBooksRes))
	requireT.ElementsMatch([]types.OrderBookData{
		{
			BaseDenom:  denom1,
			QuoteDenom: denom2,
		},
		{
			BaseDenom:  denom2,
			QuoteDenom: denom1,
		},
	}, orderBooksRes.OrderBooks)

	order2 := types.Order{
		Creator:           creator.String(),
		ID:                "id2",
		BaseDenom:         denom1,
		QuoteDenom:        denom3,
		Price:             types.MustNewPriceFromString("124e-2"),
		Quantity:          sdkmath.NewInt(100),
		Side:              types.Side_sell,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
	}
	placeOrder(ctx, requireT, testNetwork, order2)

	// check orders
	var ordersRes types.QueryOrdersResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(ctx, cli.CmdQueryOrders(), []string{creator.String()}, &ordersRes))
	requireT.ElementsMatch([]types.Order{
		order1,
		order2,
	}, ordersRes.Orders)

	// check order book orders
	var orderBookOrdersRes types.QueryOrderBookOrdersResponse
	requireT.NoError(coreumclitestutil.ExecQueryCmd(
		ctx, cli.CmdQueryOrderBookOrders(), []string{denom1, denom2, types.Side_sell.String()}, &orderBookOrdersRes),
	)
	requireT.ElementsMatch([]types.Order{
		order1,
	}, orderBookOrdersRes.Orders)
}
