//go:build integrationtests

package modules

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// TestOrdersMatching tests the dex modules ability to place get and match orders.
func TestOrdersMatching(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      dextypes.MustNewPriceFromString("1e-1"),
		Quantity:   sdkmath.NewInt(100),
		Side:       dextypes.Side_sell,
	}

	txResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(placeSellOrderMsg), uint64(txResult.GasUsed))

	sellOrderRes, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dextypes.Order{
		Creator:           acc1.String(),
		ID:                "id1",
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             dextypes.MustNewPriceFromString("1e-1"),
		Quantity:          sdkmath.NewInt(100),
		Side:              dextypes.Side_sell,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc2.String(),
		ID:         "id1", // same ID allowed for different user
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      dextypes.MustNewPriceFromString("11e-2"),
		Quantity:   sdkmath.NewInt(300),
		Side:       dextypes.Side_buy,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeBuyOrderMsg)),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	// now query the sell order
	_, err = dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.ErrorContains(err, dextypes.ErrRecordNotFound.Error())

	// check remaining buy order
	buyOrderRes, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeBuyOrderMsg.Sender,
		Id:      placeBuyOrderMsg.ID,
	})
	requireT.NoError(err)
	requireT.NotNil(buyOrderRes.Order)

	requireT.Equal(dextypes.Order{
		Creator:           acc2.String(),
		ID:                "id1", // same ID allowed for different users
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             dextypes.MustNewPriceFromString("11e-2"),
		Quantity:          sdkmath.NewInt(300),
		Side:              dextypes.Side_buy,
		RemainingQuantity: sdkmath.NewInt(200),
		RemainingBalance:  sdkmath.NewInt(23),
	}, buyOrderRes.Order)

	acc1Denom2BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(10).String(), acc1Denom2BalanceRes.Balance.Amount.String())

	acc2Denom1BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc2.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(100).String(), acc2Denom1BalanceRes.Balance.Amount.String())
}

// TestOrderCancellation tests the dex modules ability to place cancel placed order.
func TestOrderCancellation(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgCancelOrder{},
		},
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: "denom2",
		Price:      dextypes.MustNewPriceFromString("1e-1"),
		Quantity:   sdkmath.NewInt(100),
		Side:       dextypes.Side_sell,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	cancelOrderMsg := &dextypes.MsgCancelOrder{
		Sender: placeSellOrderMsg.Sender,
		ID:     placeSellOrderMsg.ID,
	}

	txResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(cancelOrderMsg)),
		cancelOrderMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(cancelOrderMsg), uint64(txResult.GasUsed))

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	// check that nothing is locked
	requireT.Equal(sdkmath.ZeroInt().String(), balanceRes.LockedInDEX.String())
}

// TestOrderBooksAndOrdersQueries tests the dex modules order queries.
func TestOrderBooksAndOrdersQueries(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	// issue assetft
	acc1 := chain.GenAccount()
	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	acc2 := chain.GenAccount()
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	// create acc1 orders
	acc1Orders := []dextypes.Order{
		{
			Creator:           acc1.String(),
			ID:                "id1",
			BaseDenom:         denom1,
			QuoteDenom:        denom2,
			Price:             dextypes.MustNewPriceFromString("999"),
			Quantity:          sdkmath.NewInt(100),
			Side:              dextypes.Side_sell,
			RemainingQuantity: sdkmath.NewInt(100),
			RemainingBalance:  sdkmath.NewInt(100),
		},
	}
	acc1OrderPlaceMsgs := ordersToPlaceMsgs(acc1Orders)
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: acc1OrderPlaceMsgs,
	})
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(acc1OrderPlaceMsgs...)),
		acc1OrderPlaceMsgs...,
	)
	requireT.NoError(err)

	// create acc2 orders
	acc2Orders := []dextypes.Order{
		{
			Creator:           acc2.String(),
			ID:                "id1",
			BaseDenom:         denom1,
			QuoteDenom:        denom2,
			Price:             dextypes.MustNewPriceFromString("996"),
			Quantity:          sdkmath.NewInt(10),
			Side:              dextypes.Side_buy,
			RemainingQuantity: sdkmath.NewInt(10),
			RemainingBalance:  sdkmath.NewInt(9960),
		},
		{
			Creator:           acc2.String(),
			ID:                "id2",
			BaseDenom:         denom1,
			QuoteDenom:        denom2,
			Price:             dextypes.MustNewPriceFromString("997"),
			Quantity:          sdkmath.NewInt(10),
			Side:              dextypes.Side_buy,
			RemainingQuantity: sdkmath.NewInt(10),
			RemainingBalance:  sdkmath.NewInt(9970),
		},
	}
	acc2OrderPlaceMsgs := ordersToPlaceMsgs(acc2Orders)
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: acc2OrderPlaceMsgs,
	})
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(acc2OrderPlaceMsgs...)),
		acc2OrderPlaceMsgs...,
	)
	requireT.NoError(err)

	// check order books query
	orderBooksRes, err := dexClient.OrderBooks(ctx, &dextypes.QueryOrderBooksRequest{})
	requireT.NoError(err)
	requireT.Contains(orderBooksRes.OrderBooks, dextypes.OrderBookData{
		BaseDenom:  denom1,
		QuoteDenom: denom2,
	})
	requireT.Contains(orderBooksRes.OrderBooks, dextypes.OrderBookData{
		BaseDenom:  denom2,
		QuoteDenom: denom1,
	})

	// check order book orders query
	orderBookOrdersRes, err := dexClient.OrdersBookOrders(ctx, &dextypes.QueryOrderBookOrdersRequest{
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Side:       dextypes.Side_sell,
	})
	requireT.NoError(err)
	// acc1 orders because all of them sell
	requireT.Equal(acc1Orders, orderBookOrdersRes.Orders)

	// check account orders query
	ordersRes, err := dexClient.Orders(ctx, &dextypes.QueryOrdersRequest{
		Creator: acc2.String(),
	})
	requireT.NoError(err)
	requireT.Equal(acc2Orders, ordersRes.Orders)
}

func issueFT(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	issuer sdk.AccAddress,
	initialAmount sdkmath.Int,
) string {
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "TKN" + uuid.NewString()[:4],
		Subunit:       "tkn" + uuid.NewString()[:4],
		Precision:     5,
		InitialAmount: initialAmount,
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	return assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
}

func ordersToPlaceMsgs(orders []dextypes.Order) []sdk.Msg {
	return lo.Map(orders, func(order dextypes.Order, _ int) sdk.Msg {
		return &dextypes.MsgPlaceOrder{
			Sender:     order.Creator,
			ID:         order.ID,
			BaseDenom:  order.BaseDenom,
			QuoteDenom: order.QuoteDenom,
			Price:      order.Price,
			Quantity:   order.Quantity,
			Side:       order.Side,
		}
	})
}
