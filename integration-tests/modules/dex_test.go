//go:build integrationtests

package modules

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeBuyOrderMsg)),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	acc1BalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: acc1.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(999900)),
			sdk.NewCoin(denom2, sdkmath.NewInt(10)),
			dexParamsRes.Params.OrderReserve,
		).String(),
		acc1BalancesRes.Balances.String(),
	)

	acc2BalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: acc2.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(100)),
			sdk.NewCoin(denom2, sdkmath.NewInt(999990)),
			dexParamsRes.Params.OrderReserve,
		).String(),
		acc2BalancesRes.Balances.String(),
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
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgCancelOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceDenom1Res, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom1,
	})
	requireT.NoError(err)
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceDenom1Res.LockedInDEX.String())
	requireT.True(balanceDenom1Res.WhitelistingReservedInDex.IsZero())

	balanceDenom2Res, err := assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	requireT.True(balanceDenom2Res.LockedInDEX.IsZero())
	whitelistingReservedInDex, err := dextypes.ComputeLimitOrderWhitelistingReservedBalance(
		placeSellOrderMsg.Side,
		placeSellOrderMsg.BaseDenom,
		placeSellOrderMsg.QuoteDenom,
		placeSellOrderMsg.Quantity,
		*placeSellOrderMsg.Price,
	)
	requireT.NoError(err)
	requireT.Equal(whitelistingReservedInDex.Amount.String(), balanceDenom2Res.WhitelistingReservedInDex.String())

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
	requireT.True(balanceDenom1Res.WhitelistingReservedInDex.IsZero())

	balanceDenom2Res, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denom2,
	})
	requireT.NoError(err)
	// check that nothing is locked
	requireT.True(balanceDenom2Res.LockedInDEX.IsZero())
	requireT.True(balanceDenom2Res.WhitelistingReservedInDex.IsZero())

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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		Messages: acc1OrderPlaceMsgs,
		Amount:   dexParamsRes.Params.OrderReserve.Amount,
	})
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(acc1OrderPlaceMsgs...)),
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
		Messages: acc2OrderPlaceMsgs,
		Amount:   dexParamsRes.Params.OrderReserve.Amount.MulRaw(int64(len(acc2OrderPlaceMsgs))),
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrDEXLockFailed.Error())

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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeBuyOrderMsg)),
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
		},
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2),
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrDEXLockFailed.Error())

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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeBuyOrderMsg)),
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
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgCancelOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
	require.NoError(t, err)

	delegateAmount := sdkmath.NewInt(1000)

	acc := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&dextypes.MsgPlaceOrder{},
		},
		Amount: delegateAmount.Add(dexParamsRes.Params.OrderReserve.Amount),
	})

	denomToStake := chain.ChainSettings.Denom
	denom2 := issueFT(ctx, t, chain, acc, sdkmath.NewIntWithDecimal(1, 6))

	// setup validator
	_, validator1Address, deactivateValidator, err := chain.CreateValidator(
		ctx, t, customStakingParams.Params.MinSelfDelegation, customStakingParams.Params.MinSelfDelegation,
	)
	require.NoError(t, err)
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.ErrorContains(err, assetfttypes.ErrDEXLockFailed.Error())

	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{}, // fund one more time since we paid for the failed message
		},
		Amount: delegateAmount,
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	balanceRes, err = assetFTClient.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc.String(),
		Denom:   denomToStake,
	})
	requireT.NoError(err)
	requireT.Equal(
		delegateAmount.Add(dexParamsRes.Params.OrderReserve.Amount).String(),
		balanceRes.Balance.String(),
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
	require.NoError(t, err)
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
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
	require.NoError(t, err)
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount,
	})

	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&dextypes.MsgPlaceOrder{},
			&dextypes.MsgPlaceOrder{},
		},
		Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeBuyOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeSellOrderMsg)),
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
		Messages: placeMsgs,
		Amount:   dexParamsRes.Params.OrderReserve.Amount.MulRaw(int64(len(placeMsgs))),
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
