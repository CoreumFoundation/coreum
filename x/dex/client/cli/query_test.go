package cli_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v6/testutil/cli"
	"github.com/CoreumFoundation/coreum/v6/testutil/network"
	"github.com/CoreumFoundation/coreum/v6/x/dex/client/cli"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

var defaultQuantity = sdkmath.NewInt(1_000_000)

func TestQueryParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx

	var resp types.QueryParamsResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryParams(), []string{}, &resp)
	requireT.Equal(types.DefaultParams(), resp.Params)
}

func TestQueryOrderBookParams(t *testing.T) {
	requireT := require.New(t)

	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, sdkmath.NewInt(1_000_000))
	denom2 := issueFT(ctx, requireT, testNetwork, sdkmath.NewInt(1_000_000))

	var resp types.QueryOrderBookParamsResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryOrderBookParams(), []string{denom1, denom2}, &resp)
	requireT.Equal("1e-6", resp.PriceTick.String())
	requireT.Equal("10000", resp.QuantityStep.String())
	requireT.Equal("1000000", resp.BaseDenomUnifiedRefAmount.TruncateInt().String())
	requireT.Equal("1000000", resp.QuoteDenomUnifiedRefAmount.TruncateInt().String())
}

func TestCmdQueryOrderBooksAndOrders(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, defaultQuantity.MulRaw(10))
	denom2 := issueFT(ctx, requireT, testNetwork, defaultQuantity)
	denom3 := issueFT(ctx, requireT, testNetwork, defaultQuantity)

	var resp types.QueryParamsResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryParams(), []string{}, &resp)

	creator := validator1Address(testNetwork)
	order1 := types.Order{
		Creator:                   creator.String(),
		Type:                      types.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  1,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(types.MustNewPriceFromString("123e-2")),
		Quantity:                  defaultQuantity,
		Side:                      types.SIDE_SELL,
		TimeInForce:               types.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     defaultQuantity,
		RemainingSpendableBalance: defaultQuantity,
		Reserve:                   resp.Params.OrderReserve,
	}
	placeOrder(ctx, requireT, testNetwork, order1)

	// check single order
	var orderRes types.QueryOrderResponse
	coreumclitestutil.ExecQueryCmd(
		t, ctx, cli.CmdQueryOrder(), []string{creator.String(), order1.ID}, &orderRes,
	)
	requireT.Equal(order1, orderRes.Order)

	// check order books
	var orderBooksRes types.QueryOrderBooksResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryOrderBooks(), []string{}, &orderBooksRes)
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
		Creator:                   creator.String(),
		Type:                      types.ORDER_TYPE_LIMIT,
		ID:                        "id2",
		Sequence:                  2,
		BaseDenom:                 denom1,
		QuoteDenom:                denom3,
		Price:                     lo.ToPtr(types.MustNewPriceFromString("124e-2")),
		Quantity:                  defaultQuantity,
		Side:                      types.SIDE_SELL,
		TimeInForce:               types.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     defaultQuantity,
		RemainingSpendableBalance: defaultQuantity,
		Reserve:                   resp.Params.OrderReserve,
	}
	placeOrder(ctx, requireT, testNetwork, order2)

	// check orders
	var ordersRes types.QueryOrdersResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryOrders(), []string{creator.String()}, &ordersRes)
	requireT.ElementsMatch([]types.Order{
		order1,
		order2,
	}, ordersRes.Orders)

	// check order book orders
	var orderBookOrdersRes types.QueryOrderBookOrdersResponse
	coreumclitestutil.ExecQueryCmd(
		t, ctx, cli.CmdQueryOrderBookOrders(), []string{denom1, denom2, types.SIDE_SELL.String()}, &orderBookOrdersRes,
	)
	requireT.ElementsMatch([]types.Order{
		order1,
	}, orderBookOrdersRes.Orders)
}

func TestCmdQueryAccountDenomOrdersCount(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, defaultQuantity.MulRaw(10))
	denom2 := issueFT(ctx, requireT, testNetwork, defaultQuantity)
	denom3 := issueFT(ctx, requireT, testNetwork, defaultQuantity)

	var resp types.QueryParamsResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryParams(), []string{}, &resp)

	creator := validator1Address(testNetwork)
	order1 := types.Order{
		Creator:                   creator.String(),
		Type:                      types.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  1,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(types.MustNewPriceFromString("123e-2")),
		Quantity:                  defaultQuantity,
		Side:                      types.SIDE_SELL,
		TimeInForce:               types.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     defaultQuantity,
		RemainingSpendableBalance: defaultQuantity,
		Reserve:                   resp.Params.OrderReserve,
	}
	placeOrder(ctx, requireT, testNetwork, order1)

	// check single order
	var orderRes types.QueryOrderResponse
	coreumclitestutil.ExecQueryCmd(
		t, ctx, cli.CmdQueryOrder(), []string{creator.String(), order1.ID}, &orderRes,
	)
	requireT.Equal(order1, orderRes.Order)

	// check order books
	var orderBooksRes types.QueryOrderBooksResponse
	coreumclitestutil.ExecQueryCmd(t, ctx, cli.CmdQueryOrderBooks(), []string{}, &orderBooksRes)
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
		Creator:                   creator.String(),
		Type:                      types.ORDER_TYPE_LIMIT,
		ID:                        "id2",
		Sequence:                  2,
		BaseDenom:                 denom1,
		QuoteDenom:                denom3,
		Price:                     lo.ToPtr(types.MustNewPriceFromString("124e-2")),
		Quantity:                  defaultQuantity,
		Side:                      types.SIDE_SELL,
		TimeInForce:               types.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     defaultQuantity,
		RemainingSpendableBalance: defaultQuantity,
	}
	placeOrder(ctx, requireT, testNetwork, order2)

	// check orders
	var ordersRes types.QueryAccountDenomOrdersCountResponse
	coreumclitestutil.ExecQueryCmd(
		t, ctx, cli.CmdQueryAccountDenomOrdersCount(), []string{creator.String(), denom1}, &ordersRes,
	)
	requireT.Equal(uint64(2), ordersRes.Count)
}
