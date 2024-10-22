package stress

import (
	sdkmath "cosmossdk.io/math"
	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/event"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	deterministicgastypes "github.com/CoreumFoundation/coreum/v5/x/deterministicgas/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestLimitOrdersStressMatching tests the dex modules ability to match a lot of orders.
func TestLimitOrdersStressMatching(t *testing.T) {
	// t.Parallel() is disabled intentionally
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	preCreatedOrdersCreator := "devcore1fnrehr95flfgnzjcatv7a8hpernwufpd5zjm2v"
	baseDenom := "dexsu-" + preCreatedOrdersCreator // created VIA the genesis
	quoteDenom := chain.ChainSettings.Denom

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// check initial state
	creatorOrdersRes, err := dexClient.Orders(ctx, &dextypes.QueryOrdersRequest{
		Creator: preCreatedOrdersCreator,
		Pagination: &query.PageRequest{
			Limit:      query.PaginationMaxLimit,
			CountTotal: true,
		},
	})
	requireT.NoError(err)
	requireT.Len(creatorOrdersRes.Orders, 10_000)

	sellTotal := sdkmath.ZeroInt()
	for _, order := range creatorOrdersRes.Orders {
		// check at lest one order
		requireT.Equal(baseDenom, order.BaseDenom)
		requireT.Equal(quoteDenom, order.QuoteDenom)
		requireT.Equal(dextypes.SIDE_SELL, order.Side)
		requireT.Equal(dextypes.MustNewPriceFromString("1").String(), order.Price.String())
		sellTotal = sellTotal.Add(order.Quantity)
	}

	// place order to match all orders
	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	buyTotal := sellTotal
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(buyTotal),
	})

	baseDenomBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc1.String(),
		Denom:   baseDenom,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.ZeroInt().String(), baseDenomBalanceRes.Balance.Amount.String())

	placeOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   baseDenom,
		QuoteDenom:  quoteDenom,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    buyTotal,
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	t.Logf("Placing order to match %d orders", len(creatorOrdersRes.Orders))
	now := time.Now()
	txRes, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeOrderMsg)),
		placeOrderMsg,
	)
	requireT.NoError(err)
	t.Logf("Placement passed, spent %s", time.Since(now))

	// check final state
	creatorOrdersRes, err = dexClient.Orders(ctx, &dextypes.QueryOrdersRequest{
		Creator: preCreatedOrdersCreator,
		Pagination: &query.PageRequest{
			Limit:      query.PaginationMaxLimit,
			CountTotal: true,
		},
	})
	requireT.NoError(err)
	requireT.Empty(creatorOrdersRes.Orders)

	baseDenomBalanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc1.String(),
		Denom:   baseDenom,
	})
	requireT.NoError(err)
	requireT.Equal(buyTotal.String(), baseDenomBalanceRes.Balance.Amount.String())

	gasEvts, err := event.FindTypedEvents[*deterministicgastypes.EventGas](txRes.Events)
	requireT.NoError(err)
	requireT.Len(gasEvts, 1)

	t.Logf("RealGas: %d, deterministicGas: %d", gasEvts[0].RealGas, gasEvts[0].DeterministicGas)
}
