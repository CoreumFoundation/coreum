//go:build integrationtests

package modules

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v5/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v5/x/customparams/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// TestLimitOrdersMatching tests the dex modules ability to place get and match limit orders.
func TestLimitOrdersMatching(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	sellOrderRes, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dextypes.Order{
		Creator:           acc1.String(),
		Type:              dextypes.ORDER_TYPE_LIMIT,
		ID:                "id1",
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:          sdkmath.NewInt(100),
		Side:              dextypes.SIDE_SELL,
		TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
		Reserve:           dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
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
		Type:              dextypes.ORDER_TYPE_LIMIT,
		ID:                "id1", // same ID allowed for different users
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:          sdkmath.NewInt(300),
		Side:              dextypes.SIDE_BUY,
		TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
		RemainingQuantity: sdkmath.NewInt(200),
		RemainingBalance:  sdkmath.NewInt(22),
		Reserve:           dexParamsRes.Params.OrderReserve,
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

// TestMarketOrdersMatching tests the dex modules ability to place match market orders.
func TestMarketOrdersMatching(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	// place buy market order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc2.String(),
		Type:       dextypes.ORDER_TYPE_MARKET,
		ID:         "id2",
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Quantity:   sdkmath.NewInt(300),
		Side:       dextypes.SIDE_BUY,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	acc1BalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: acc1.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdkmath.NewInt(999900).String(),
		acc1BalancesRes.Balances.AmountOf(denom1).String(),
	)
	requireT.Equal(
		sdkmath.NewInt(10).String(),
		acc1BalancesRes.Balances.AmountOf(denom2).String(),
	)

	acc2BalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: acc2.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdkmath.NewInt(100).String(),
		acc2BalancesRes.Balances.AmountOf(denom1).String(),
	)
	requireT.Equal(
		sdkmath.NewInt(999990).String(),
		acc2BalancesRes.Balances.AmountOf(denom2).String(),
	)
}

// TestOrderCancellation tests the dex modules ability to place cancel placed order.
func TestOrderCancellation(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgCancelOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_whitelisting)

	// fund acc1
	bankSendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom1, 100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(bankSendMsg)),
		bankSendMsg,
	)
	requireT.NoError(err)

	// whitelisting acc1 to receive denom2
	setWhitelistedLimitMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewInt64Coin(denom2, 10),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(setWhitelistedLimitMsg)),
		setWhitelistedLimitMsg,
	)
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceDenom1Res, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceDenom1Res.LockedInDEX.String())
	requireT.True(balanceDenom1Res.ExpectedToReceiveInDEX.IsZero())

	balanceDenom2Res, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.True(balanceDenom2Res.LockedInDEX.IsZero())
	expectedToReceiveInDex, err := dextypes.ComputeLimitOrderExpectedToReceiveBalance(
		placeSellOrderMsg.Side,
		placeSellOrderMsg.BaseDenom,
		placeSellOrderMsg.QuoteDenom,
		placeSellOrderMsg.Quantity,
		*placeSellOrderMsg.Price,
	)
	requireT.NoError(err)
	requireT.Equal(expectedToReceiveInDex.Amount.String(), balanceDenom2Res.ExpectedToReceiveInDEX.String())

	countRes, err := dexClient.AccountDenomOrdersCount(ctx, &dextypes.QueryAccountDenomOrdersCountRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(uint64(1), countRes.Count)

	countRes, err = dexClient.AccountDenomOrdersCount(ctx, &dextypes.QueryAccountDenomOrdersCountRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.Equal(uint64(1), countRes.Count)

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

	balanceDenom1Res, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	// check that nothing is locked
	requireT.True(balanceDenom1Res.LockedInDEX.IsZero())
	requireT.True(balanceDenom1Res.ExpectedToReceiveInDEX.IsZero())

	balanceDenom2Res, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	// check that nothing is locked
	requireT.True(balanceDenom2Res.LockedInDEX.IsZero())
	requireT.True(balanceDenom2Res.ExpectedToReceiveInDEX.IsZero())

	countRes, err = dexClient.AccountDenomOrdersCount(ctx, &dextypes.QueryAccountDenomOrdersCountRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(uint64(0), countRes.Count)

	countRes, err = dexClient.AccountDenomOrdersCount(ctx, &dextypes.QueryAccountDenomOrdersCountRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.Equal(uint64(0), countRes.Count)
}

// TestOrderTilBlockHeight tests the dex modules ability to place cancel placed order with good til block height.
func TestOrderTilBlockHeight(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))

	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: "denom2",
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:   sdkmath.NewInt(100),
		Side:       dextypes.SIDE_SELL,
		GoodTil: lo.ToPtr(dextypes.GoodTil{
			GoodTilBlockHeight: uint64(blockRes.SdkBlock.Header.Height + 20),
		}),
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// await for cancellation
	requireT.NoError(chain.AwaitState(ctx, func(ctx context.Context) error {
		// check that order is cancelled and balance is unlocked
		balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
			Account: acc1.String(),
			Denom:   denom1,
		})
		requireT.NoError(err)
		// check that nothing is locked
		if balanceRes.LockedInDEX.IsZero() {
			return nil
		}
		// we are waiting for the cosmos state, not block, because the block might be updated but cosmos state not
		return retry.Retryable(errors.Errorf("waiting to balance to be unlocked"))
	}))
}

// TestOrderTilBlockTime tests the dex modules ability to place cancel placed order with good til block time.
func TestOrderTilBlockTime(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))

	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: "denom2",
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:   sdkmath.NewInt(100),
		Side:       dextypes.SIDE_SELL,
		GoodTil: lo.ToPtr(dextypes.GoodTil{
			GoodTilBlockTime: lo.ToPtr(blockRes.SdkBlock.Header.Time.Add(10 * time.Second)),
		}),
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// await for cancellation
	requireT.NoError(chain.AwaitState(ctx, func(ctx context.Context) error {
		// check that order is cancelled and balance is unlocked
		balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
			Account: acc1.String(),
			Denom:   denom1,
		})
		requireT.NoError(err)
		// check that nothing is locked
		if balanceRes.LockedInDEX.IsZero() {
			return nil
		}
		// we are waiting for the cosmos state, not block, because the block might be updated but cosmos state not
		return retry.Retryable(errors.Errorf("waiting to balance to be unlocked"))
	}))
}

// TestOrderBooksAndOrdersQueries tests the dex modules order queries.
func TestOrderBooksAndOrdersQueries(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)

	// issue assetft
	acc1 := chain.GenAccount()
	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	acc2 := chain.GenAccount()
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	// create acc1 orders
	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1Orders := []dextypes.Order{
		{
			Creator:    acc1.String(),
			Type:       dextypes.ORDER_TYPE_LIMIT,
			ID:         "id1",
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(dextypes.MustNewPriceFromString("999")),
			Quantity:   sdkmath.NewInt(100),
			Side:       dextypes.SIDE_SELL,
			GoodTil: &dextypes.GoodTil{
				GoodTilBlockHeight: uint64(blockRes.SdkBlock.Header.Height + 500),
			},
			TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
			RemainingQuantity: sdkmath.NewInt(100),
			RemainingBalance:  sdkmath.NewInt(100),
			Reserve:           dexParamsRes.Params.OrderReserve,
		},
	}
	acc1OrderPlaceMsgs := ordersToPlaceMsgs(acc1Orders)
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		acc1OrderPlaceMsgs...,
	)
	requireT.NoError(err)

	// create acc2 orders
	acc2Orders := []dextypes.Order{
		{
			Creator:    acc2.String(),
			Type:       dextypes.ORDER_TYPE_LIMIT,
			ID:         "id1",
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(dextypes.MustNewPriceFromString("996")),
			Quantity:   sdkmath.NewInt(10),
			Side:       dextypes.SIDE_BUY,
			GoodTil: &dextypes.GoodTil{
				GoodTilBlockHeight: uint64(blockRes.SdkBlock.Header.Height + 1000),
			},
			TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
			RemainingQuantity: sdkmath.NewInt(10),
			RemainingBalance:  sdkmath.NewInt(9960),
			Reserve:           dexParamsRes.Params.OrderReserve,
		},
		{
			Creator:           acc2.String(),
			Type:              dextypes.ORDER_TYPE_LIMIT,
			ID:                "id2",
			BaseDenom:         denom1,
			QuoteDenom:        denom2,
			Price:             lo.ToPtr(dextypes.MustNewPriceFromString("997")),
			Quantity:          sdkmath.NewInt(10),
			Side:              dextypes.SIDE_BUY,
			TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
			RemainingQuantity: sdkmath.NewInt(10),
			RemainingBalance:  sdkmath.NewInt(9970),
			Reserve:           dexParamsRes.Params.OrderReserve,
		},
	}
	acc2OrderPlaceMsgs := ordersToPlaceMsgs(acc2Orders)
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(int64(len(acc2OrderPlaceMsgs))).AddRaw(100_000),
	})
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
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
	orderBookOrdersRes, err := dexClient.OrderBookOrders(ctx, &dextypes.QueryOrderBookOrdersRequest{
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Side:       dextypes.SIDE_SELL,
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

// TestDEXProposalParamChange checks that dex param change proposal works correctly.
func TestDEXProposalParamChange(t *testing.T) {
	// Since this test changes global we can't run it in parallel with other tests.
	// That's why t.Parallel() is not here.

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx, false)
	// For the test we need to create the proposal twice.
	requireT.NoError(err)
	proposerBalance = proposerBalance.Add(proposerBalance)
	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(proposer, proposerBalance))

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	initialParams := dexParamsRes.Params

	// restore params at the end of test
	defer func() {
		t.Logf("Restoring initial dex params.")
		chain.Governance.ProposalFromMsgAndVote(
			ctx, t, nil,
			"-", "-", "-", govtypesv1.OptionYes,
			&dextypes.MsgUpdateParams{
				Params:    initialParams,
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			},
		)
		dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
		requireT.NoError(err)
		requireT.Equal(initialParams, dexParamsRes.Params)
	}()

	newParams := initialParams
	newParams.DefaultUnifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("33.01")
	newParams.PriceTickExponent = -33
	newParams.OrderReserve = sdk.NewInt64Coin(initialParams.OrderReserve.Denom, 1)

	chain.Governance.ProposalFromMsgAndVote(
		ctx, t, nil,
		"-", "-", "-", govtypesv1.OptionYes,
		&dextypes.MsgUpdateParams{
			Params:    newParams,
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		},
	)
	requireT.NoError(err)

	dexParamsRes, err = dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(newParams, dexParamsRes.Params)
}

// TestLimitOrdersMatchingWithAssetFTFreeze tests the dex modules ability to place get and match limit orders
// with asset ft with freezing.
func TestLimitOrdersMatchingWithAssetFTFreeze(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgFreeze{},
			&assetfttypes.MsgUnfreeze{},
			&assetfttypes.MsgFreeze{},
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_freezing)
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// freeze all tokens
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(150)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())

	// place order should fail because all the funds are frozen
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrDEXInsufficientSpendableBalance.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())

	// change the frozen amount to less than the order quantity
	unfreezeMsg := &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(50).String(), balanceRes.Frozen.String())

	// now placing order should succeed because the needed funds are more than frozen amount
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(50).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// freeze remaining tokens
	freezeMsg = &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100)),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

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

// TestLimitOrdersMatchingWithAssetFTGloballyFreeze tests the dex modules ability to place get and match limit orders
// with asset ft with globally freezing.
func TestLimitOrdersMatchingWithAssetFTGloballyFreeze(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgGloballyFreeze{},
			&assetfttypes.MsgGloballyUnfreeze{},
			&assetfttypes.MsgGloballyFreeze{},
			&assetfttypes.MsgGloballyUnfreeze{},
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_freezing)
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// globally freeze the denom
	freezeMsg := &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom1,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())

	// place order should fail because all the funds are globally frozen
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(
		err,
		fmt.Sprintf("usage of %s for DEX is blocked because the token is globally frozen", denom1),
	)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())

	// unfreeze the denom globally
	unfreezeMsg := &assetfttypes.MsgGloballyUnfreeze{
		Sender: issuer.String(),
		Denom:  denom1,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.Frozen.String())

	// now placing order should succeed because the needed funds are not frozen
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// globally freeze the denom
	freezeMsg = &assetfttypes.MsgGloballyFreeze{
		Sender: issuer.String(),
		Denom:  denom1,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	// it's because the token the acc receives is globally frozen
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("usage of %s for DEX is blocked because the token is globally frozen", denom1))

	// globally unfreeze the denom and place order one more time
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)

	// it's because the token the acc receives is globally frozen
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

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

// TestLimitOrdersMatchingWithAssetClawback tests the dex modules ability to place get and match limit orders
// with asset ft with clawback feature.
func TestLimitOrdersMatchingWithAssetClawback(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgClawback{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgClawback{},
			&assetfttypes.MsgClawback{},
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgCancelOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_clawback)
	denom2 := "denom2"

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	// clawback some of the amount
	clawbackMsg := &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(50).String(), balanceRes.Balance.String())

	// place order should fail because of insufficient funds
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, cosmoserrors.ErrInsufficientFunds.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(50).String(), balanceRes.Balance.String())

	// send enough amounts for the order
	msgSend = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(100)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Balance.String())

	// now placing order should succeed because the needed funds are available
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Balance.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// try to clawback after placing the order should fail
	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.ErrorContains(err, cosmoserrors.ErrInsufficientFunds.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Balance.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// the order should be cancelled, in order to do the clawback
	cancelOrderMsg := &dextypes.MsgCancelOrder{
		Sender: acc1.String(),
		ID:     placeSellOrderMsg.ID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(cancelOrderMsg)),
		cancelOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.ZeroInt().String(), balanceRes.LockedInDEX.String())

	// now clawback should succeed
	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100)),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(clawbackMsg)),
		clawbackMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(50).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.ZeroInt().String(), balanceRes.LockedInDEX.String())
}

// TestLimitOrdersMatchingWithStaking tests the dex modules ability to place get and match limit orders with staking.
func TestLimitOrdersMatchingWithStaking(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	requireT.NoError(err)

	delegateAmount := sdkmath.NewInt(1_000_000)

	acc := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
		},
		Amount: delegateAmount.Add(dexParamsRes.Params.OrderReserve.Amount).Add(sdkmath.NewInt(100_000)),
	})

	denomToStake := chain.ChainSettings.Denom
	denom2 := issueFT(ctx, t, chain, acc, sdkmath.NewIntWithDecimal(1, 6))

	// setup validator
	_, validator1Address, deactivateValidator, err := chain.CreateValidator(
		ctx, t, customStakingParams.Params.MinSelfDelegation, customStakingParams.Params.MinSelfDelegation,
	)
	requireT.NoError(err)
	defer deactivateValidator()

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc.String(),
		Denom:   denomToStake,
	})
	requireT.NoError(err)
	requireT.True(balanceRes.Balance.GTE(delegateAmount))

	delegateMsg := &stakingtypes.MsgDelegate{
		DelegatorAddress: acc.String(),
		ValidatorAddress: validator1Address.String(),
		Amount:           sdk.NewCoin(denomToStake, delegateAmount),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	requireT.NoError(err)

	// place order should fail because all the funds are staked
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denomToStake,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    delegateAmount,
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrDEXInsufficientSpendableBalance.Error())

	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		Amount: delegateAmount.Add(sdkmath.NewInt(100_000)),
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc.String(),
		Denom:   denomToStake,
	})
	requireT.NoError(err)
	requireT.True(
		delegateAmount.Add(dexParamsRes.Params.OrderReserve.Amount).LTE(balanceRes.Balance),
	)
	requireT.Equal(
		delegateAmount.Add(dexParamsRes.Params.OrderReserve.Amount).String(),
		balanceRes.LockedInDEX.String(),
	)
}

// TestLimitOrdersMatchingWithBurnRate tests the dex modules ability to place get and match limit orders with burn rate.
func TestLimitOrdersMatchingWithBurnRate(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        acc1.String(),
		Symbol:        "BURNTKN" + uuid.NewString()[:4],
		Subunit:       "burntkn" + uuid.NewString()[:4],
		Precision:     5,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 6),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_burning},
		BurnRate:      sdkmath.LegacyMustNewDecFromStr("0.5"),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	denom1 := assetfttypes.BuildDenom(issueMsg.Subunit, acc1)
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceBeforePlaceOrder := balanceRes.Balance

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceAfterPlaceOrder := balanceRes.Balance
	requireT.Equal(balanceBeforePlaceOrder, balanceAfterPlaceOrder)

	sellOrderRes, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dextypes.Order{
		Creator:           acc1.String(),
		Type:              dextypes.ORDER_TYPE_LIMIT,
		ID:                "id1",
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:          sdkmath.NewInt(100),
		Side:              dextypes.SIDE_SELL,
		TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
		Reserve:           dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	// now query the sell order
	_, err = dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.ErrorContains(err, dextypes.ErrRecordNotFound.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceAfterMatchingOrder := balanceRes.Balance
	requireT.Equal(balanceAfterPlaceOrder.Sub(placeSellOrderMsg.Quantity).String(), balanceAfterMatchingOrder.String())

	acc2Denom1BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc2.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(100).String(), acc2Denom1BalanceRes.Balance.Amount.String())
}

// TestLimitOrdersMatchingWithCommissionRate tests the dex modules ability to place get and match limit orders with
// send commission rate.
func TestLimitOrdersMatchingWithCommissionRate(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:             acc1.String(),
		Symbol:             "COMMISSIONTKN" + uuid.NewString()[:4],
		Subunit:            "commissiontkn" + uuid.NewString()[:4],
		Precision:          5,
		InitialAmount:      sdkmath.NewIntWithDecimal(1, 6),
		Features:           []assetfttypes.Feature{assetfttypes.Feature_burning},
		SendCommissionRate: sdkmath.LegacyMustNewDecFromStr("0.5"),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	denom1 := assetfttypes.BuildDenom(issueMsg.Subunit, acc1)
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceBeforePlaceOrder := balanceRes.Balance

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceAfterPlaceOrder := balanceRes.Balance
	requireT.Equal(balanceBeforePlaceOrder, balanceAfterPlaceOrder)

	sellOrderRes, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.NoError(err)

	requireT.Equal(dextypes.Order{
		Creator:           acc1.String(),
		Type:              dextypes.ORDER_TYPE_LIMIT,
		ID:                "id1",
		BaseDenom:         denom1,
		QuoteDenom:        denom2,
		Price:             lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:          sdkmath.NewInt(100),
		Side:              dextypes.SIDE_SELL,
		TimeInForce:       dextypes.TIME_IN_FORCE_GTC,
		RemainingQuantity: sdkmath.NewInt(100),
		RemainingBalance:  sdkmath.NewInt(100),
		Reserve:           dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	// now query the sell order
	_, err = dexClient.Order(ctx, &dextypes.QueryOrderRequest{
		Creator: placeSellOrderMsg.Sender,
		Id:      placeSellOrderMsg.ID,
	})
	requireT.ErrorContains(err, dextypes.ErrRecordNotFound.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	balanceAfterMatchingOrder := balanceRes.Balance
	requireT.Equal(balanceAfterPlaceOrder.Sub(placeSellOrderMsg.Quantity).String(), balanceAfterMatchingOrder.String())

	acc2Denom1BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc2.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(100).String(), acc2Denom1BalanceRes.Balance.Amount.String())
}

// TestLimitOrdersMatchingWithAssetFTWhitelist tests the dex modules ability to place get and match limit orders
// with asset ft with whitelisting.
func TestLimitOrdersMatchingWithAssetFTWhitelist(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgSetWhitelistedLimit{},
			&assetfttypes.MsgSetWhitelistedLimit{},
			&banktypes.MsgSend{},
			&assetfttypes.MsgSetWhitelistedLimit{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_whitelisting)
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 6))

	// whitelist denom1 for acc1
	msgSetWhitelistedLimit := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(1000000)),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.Balance.String())

	// place order should fail because acc2 is out of whitelisted coins
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrWhitelistedLimitExceeded.Error())

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())

	msgSetWhitelistedLimit = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: acc2.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(300)),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	// now placing order should succeed because the receiving amount is within the whitelist limit
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc2.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.Equal("33", balanceRes.LockedInDEX.String())

	// Reducing the whitelist limit will not interfere with DEX order after order placement
	msgSetWhitelistedLimit = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: acc2.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(0)),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	// place sell order to match the buy
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	acc1Denom2BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(11).String(), acc1Denom2BalanceRes.Balance.Amount.String())

	acc2Denom1BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc2.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(100).String(), acc2Denom1BalanceRes.Balance.Amount.String())

	// place order should succeed as issuer because admin (issuer) doesn't have whitelist limit
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      issuer.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)
}

// TestCancelOrdersByDenom tests the dex modules ability to cancel all orders of the account and by denom.
func TestCancelOrdersByDenom(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	const ordersPerChunk = 10

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
		Amount: sdkmath.NewIntWithDecimal(1, 6), // amount to cover cancellation
	})

	acc1 := chain.GenAccount()
	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 10), assetfttypes.Feature_dex_order_cancellation)
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 10), assetfttypes.Feature_dex_order_cancellation)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	blockRes, err := tmQueryClient.GetLatestBlock(ctx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)

	ordersCount := int(dexParamsRes.Params.MaxOrdersPerDenom)

	amtPerOrder := sdkmath.NewInt(100)
	placeMsgs := lo.RepeatBy(ordersCount, func(_ int) sdk.Msg {
		return &dextypes.MsgPlaceOrder{
			Sender:     acc1.String(),
			Type:       dextypes.ORDER_TYPE_LIMIT,
			ID:         uuid.NewString(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
			Quantity:   amtPerOrder,
			Side:       dextypes.SIDE_SELL,
			GoodTil: lo.ToPtr(dextypes.GoodTil{
				GoodTilBlockHeight: uint64(blockRes.SdkBlock.Header.Height + 20_000),
				GoodTilBlockTime:   lo.ToPtr(blockRes.SdkBlock.Header.Time.Add(time.Hour)),
			}),
			TimeInForce: dextypes.TIME_IN_FORCE_GTC,
		}
	})
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(int64(len(placeMsgs))).AddRaw(100_000 * int64(ordersCount)),
	})

	// send required tokens to acc1
	coinToFundAcc := sdk.NewCoin(denom1, amtPerOrder.MulRaw(int64(ordersCount)))
	msgBankSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(coinToFundAcc),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgBankSend)),
		msgBankSend,
	)
	requireT.NoError(err)

	// place all orders
	for i, chunk := range lo.Chunk(placeMsgs, ordersPerChunk) {
		t.Logf("Placing orders chunk %d", i)
		_, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(acc1),
			chain.TxFactoryAuto(),
			chunk...,
		)
		requireT.NoError(err)
	}

	orderRes, err := dexClient.Orders(ctx, &dextypes.QueryOrdersRequest{
		Creator: acc1.String(),
		Pagination: &query.PageRequest{
			Limit: uint64(ordersCount),
		},
	})
	requireT.NoError(err)
	requireT.Len(orderRes.Orders, ordersCount)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   coinToFundAcc.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(coinToFundAcc.Amount.String(), balanceRes.LockedInDEX.String())

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		&dextypes.MsgCancelOrdersByDenom{
			Sender:  issuer.String(),
			Account: acc1.String(),
			Denom:   denom2,
		})
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   coinToFundAcc.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(coinToFundAcc.Amount.String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.ZeroInt().String(), balanceRes.LockedInDEX.String())

	orderRes, err = dexClient.Orders(ctx, &dextypes.QueryOrdersRequest{
		Creator: acc1.String(),
		Pagination: &query.PageRequest{
			Limit: uint64(ordersCount),
		},
	})
	requireT.NoError(err)
	requireT.Empty(orderRes.Orders)
}

func issueFT(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	issuer sdk.AccAddress,
	initialAmount sdkmath.Int,
	features ...assetfttypes.Feature,
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
		Features:      features,
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

// TestAssetFTBlockSmartContractsFeatureWithDEX tests the dex module integration with the asset ft
// block_smart_contracts features.
func TestAssetFTBlockSmartContractsFeatureWithDEX(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
			&banktypes.MsgSend{},
		},
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2),
	})

	acc := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		// 2 to place directly and 1 through the smart contract
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(3).
			Add(sdkmath.NewInt(100_000).MulRaw(3)).
			Add(chain.QueryDEXParams(ctx, t).OrderReserve.Amount),
	})

	issue1Msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "BLK" + uuid.NewString()[:4],
		Subunit:       "blk" + uuid.NewString()[:4],
		Precision:     5,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features:      []assetfttypes.Feature{assetfttypes.Feature_block_smart_contracts},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue1Msg)),
		issue1Msg,
	)
	requireT.NoError(err)
	denom1WithBlockSmartContract := assetfttypes.BuildDenom(issue1Msg.Subunit, issuer)

	// issue 2nd denom without block_smart_contracts
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6))

	sendMsg1 := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom1WithBlockSmartContract, sdkmath.NewInt(100))),
	}
	sendMsg2 := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom2, sdkmath.NewInt(100))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactoryAuto(),
		sendMsg1, sendMsg2,
	)
	requireT.NoError(err)

	// we expect to receive denom with block_smart_contracts feature
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1WithBlockSmartContract,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	// send it to chain directly
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	// we expect to spend denom with block_smart_contracts feature
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id2",
		BaseDenom:   denom2,
		QuoteDenom:  denom1WithBlockSmartContract,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(100),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	// send it to chain directly
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	// now same tokens but for the DEX smart contract

	// fund the DEX smart contract
	issuanceReq := issueFTRequest{
		Symbol:        "CTR",
		Subunit:       "ctr",
		Precision:     6,
		InitialAmount: sdkmath.NewInt(100).String(),
	}
	issuerFTInstantiatePayload, err := json.Marshal(issuanceReq)
	requireT.NoError(err)

	// instantiate new contract from the acc (the contract issues a token, but we don't use it for the test)
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		chain.TxFactoryAuto(),
		acc,
		moduleswasm.DEXWASM,
		integration.InstantiateConfig{
			Amount:     chain.QueryAssetFTParams(ctx, t).IssueFee,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    issuerFTInstantiatePayload,
			Label:      "dex",
		},
	)
	requireT.NoError(err)

	// it's prohibited to send tokens to the DEX smart contract with the denom with block_smart_contracts feature,
	// that's why we can't place and order with it
	sendMsg1 = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom1WithBlockSmartContract, sdkmath.NewInt(100))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg1)),
		sendMsg1,
	)
	requireT.Error(err)
	requireT.True(cosmoserrors.ErrUnauthorized.Is(err))
	requireT.ErrorContains(err, "transfers to smart contracts are disabled")

	// send tokens to place and order from the smart contract
	sendMsg2 = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   contractAddr,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom2, sdkmath.NewInt(100))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg2)),
		sendMsg2,
	)
	requireT.NoError(err)

	placeOrderPayload, err := json.Marshal(map[dexMethod]placeOrderBodyDEXRequest{
		dexMethodPlaceOrder: {
			Order: dextypes.Order{
				Creator:     contractAddr,
				Type:        dextypes.ORDER_TYPE_LIMIT,
				ID:          "id1",
				BaseDenom:   denom1WithBlockSmartContract,
				QuoteDenom:  denom2,
				Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
				Quantity:    sdkmath.NewInt(100),
				Side:        dextypes.SIDE_BUY,
				TimeInForce: dextypes.TIME_IN_FORCE_GTC,
				// next attributes are required by smart contract, but not used
				RemainingQuantity: sdkmath.ZeroInt(),
				RemainingBalance:  sdkmath.ZeroInt(),
				GoodTil:           nil,
				Reserve:           sdk.NewCoin("denom1", sdkmath.ZeroInt()),
			},
		},
	})
	requireT.NoError(err)

	// however the contract has the coins to place such and order, the placement is failed because the order expects
	// to receive the asset ft with block_smart_contracts feature
	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactoryAuto(),
		acc,
		contractAddr,
		placeOrderPayload,
		chain.NewCoin(chain.QueryDEXParams(ctx, t).OrderReserve.Amount),
	)
	requireT.Error(err)
	requireT.ErrorContains(
		err,
		fmt.Sprintf("usage of %s is not supported for DEX in smart contract", denom1WithBlockSmartContract),
	)
}

// TestLimitOrdersMatchingWithAssetBurning tests the dex modules ability to place get and match limit orders
// with asset ft with burning feature.
func TestLimitOrdersMatchingWithAssetBurning(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	assetFTClient := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgBurn{},
			&dextypes.MsgCancelOrder{},
			&assetfttypes.MsgBurn{},
			&assetfttypes.MsgBurn{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(1).Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_burning)
	denom2 := "denom2"

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(200)),
		),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSend)),
		msgSend,
	)
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(150),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(sdkmath.NewInt(200).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.NewInt(150).String(), balanceRes.LockedInDEX.String())

	// try to burn unburnable token, locked in dex
	burnMsg := &assetfttypes.MsgBurn{
		Sender: acc1.String(),
		Coin: sdk.Coin{
			Denom: denom1,
			// it's allowed to burn only 50 not locked in dex
			Amount: sdkmath.NewInt(100),
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.ErrorContains(err, fmt.Sprintf("100%s is not available, available 50%s", denom1, denom1))

	// cancel to burn
	cancelOrderMsg := &dextypes.MsgCancelOrder{
		Sender: acc1.String(),
		ID:     placeSellOrderMsg.ID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(cancelOrderMsg)),
		cancelOrderMsg)
	requireT.NoError(err)

	// now it's allowed to burn
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	// 100 is burnt 100 remains
	requireT.Equal(sdkmath.NewInt(100).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())
}

func ordersToPlaceMsgs(orders []dextypes.Order) []sdk.Msg {
	return lo.Map(orders, func(order dextypes.Order, _ int) sdk.Msg {
		return &dextypes.MsgPlaceOrder{
			Sender:      order.Creator,
			Type:        dextypes.ORDER_TYPE_LIMIT,
			ID:          order.ID,
			BaseDenom:   order.BaseDenom,
			QuoteDenom:  order.QuoteDenom,
			Price:       order.Price,
			Quantity:    order.Quantity,
			Side:        order.Side,
			GoodTil:     order.GoodTil,
			TimeInForce: order.TimeInForce,
		}
	})
}
