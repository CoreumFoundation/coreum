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
	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v6/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/testutil/event"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v6/x/customparams/types"
	testcontracts "github.com/CoreumFoundation/coreum/v6/x/dex/keeper/test-contracts"
	dextypes "github.com/CoreumFoundation/coreum/v6/x/dex/types"
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

	acc1, denom1 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(1_000_000),
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
		Creator:                   acc1.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  sellOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:                  sdkmath.NewInt(1_000_000),
		Side:                      dextypes.SIDE_SELL,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
		RemainingSpendableBalance: sdkmath.NewInt(1_000_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(3_000_000),
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
		Creator:                   acc2.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1", // same ID allowed for different users
		Sequence:                  buyOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:                  sdkmath.NewInt(3_000_000),
		Side:                      dextypes.SIDE_BUY,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(2_000_000),
		RemainingSpendableBalance: sdkmath.NewInt(220_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, buyOrderRes.Order)

	assertBalance(ctx, t, bankClient, acc1, denom2, 100_000)
	assertBalance(ctx, t, bankClient, acc2, denom1, 1_000_000)
}

// TestLimitOrdersMatchingFast tests the dex modules ability to place get and match limit orders.
func TestLimitOrdersMatchingFast(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1, denom1 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(1_000_000),
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
		Creator:                   acc1.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  sellOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:                  sdkmath.NewInt(1_000_000),
		Side:                      dextypes.SIDE_SELL,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(1_000_000),
		RemainingSpendableBalance: sdkmath.NewInt(1_000_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(3_000_000),
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
		Creator:                   acc2.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1", // same ID allowed for different users
		Sequence:                  buyOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:                  sdkmath.NewInt(3_000_000),
		Side:                      dextypes.SIDE_BUY,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(2_000_000),
		RemainingSpendableBalance: sdkmath.NewInt(220_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, buyOrderRes.Order)

	assertBalance(ctx, t, bankClient, acc1, denom2, 100_000)
	assertBalance(ctx, t, bankClient, acc2, denom1, 1_000_000)
}

// TestMarketOrdersMatching tests the dex modules ability to place match market orders.
func TestMarketOrdersMatching(t *testing.T) {
	t.Parallel()
	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	acc1, denom1 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
		Side:        dextypes.SIDE_SELL,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactoryAuto(),
		placeSellOrderMsg,
	)
	requireT.NoError(err)

	// place buy market order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_MARKET,
		ID:          "id2",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Quantity:    sdkmath.NewInt(300_000),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_IOC,
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
		sdkmath.NewInt(900_000).String(),
		acc1BalancesRes.Balances.AmountOf(denom1).String(),
	)
	requireT.Equal(
		sdkmath.NewInt(10_000).String(),
		acc1BalancesRes.Balances.AmountOf(denom2).String(),
	)

	acc2BalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: acc2.String(),
	})
	requireT.NoError(err)
	requireT.Equal(
		sdkmath.NewInt(100_000).String(),
		acc2BalancesRes.Balances.AmountOf(denom1).String(),
	)
	requireT.Equal(
		sdkmath.NewInt(990_000).String(),
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
	acc1 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgSetWhitelistedLimit{},
				},
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&dextypes.MsgCancelOrder{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
			},
		},
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6))
	denom2Whtlst := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_whitelisting)

	// fund acc1
	bankSendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(denom1, 1_000_000)),
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
		Coin:    sdk.NewInt64Coin(denom2Whtlst, 100_000),
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
		QuoteDenom:  denom2Whtlst,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
		Denom:   denom2Whtlst,
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
		Denom:   denom2Whtlst,
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
		Denom:   denom2Whtlst,
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
		Denom:   denom2Whtlst,
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
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))

	latestBlock, err := chain.LatestBlockHeader(ctx)
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:   sdkmath.NewInt(100_000),
		Side:       dextypes.SIDE_SELL,
		GoodTil: lo.ToPtr(dextypes.GoodTil{
			GoodTilBlockHeight: uint64(latestBlock.Height + 20),
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
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	acc1 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc1, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
	})

	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))
	denom2 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 6))

	latestBlock, err := chain.LatestBlockHeader(ctx)
	requireT.NoError(err)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc1.String(),
		Type:       dextypes.ORDER_TYPE_LIMIT,
		ID:         "id1",
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:   sdkmath.NewInt(100_000),
		Side:       dextypes.SIDE_SELL,
		GoodTil: lo.ToPtr(dextypes.GoodTil{
			GoodTilBlockTime: lo.ToPtr(latestBlock.Time.Add(10 * time.Second)),
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

	// issue assetft

	acc1 := chain.GenAccount()
	denom1 := issueFT(ctx, t, chain, acc1, sdkmath.NewIntWithDecimal(1, 10))
	acc2 := chain.GenAccount()
	denom2 := issueFT(ctx, t, chain, acc2, sdkmath.NewIntWithDecimal(1, 10))

	// create acc1 orders
	latestBlock, err := chain.LatestBlockHeader(ctx)
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
			Quantity:   sdkmath.NewInt(100_000),
			Side:       dextypes.SIDE_SELL,
			GoodTil: &dextypes.GoodTil{
				GoodTilBlockHeight: uint64(latestBlock.Height + 500),
			},
			TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
			RemainingBaseQuantity:     sdkmath.NewInt(100_000),
			RemainingSpendableBalance: sdkmath.NewInt(100_000),
			Reserve:                   dexParamsRes.Params.OrderReserve,
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
	acc1Orders = fillOrderSequences(ctx, t, chain.ClientContext, acc1Orders)

	// create acc2 orders
	acc2Orders := []dextypes.Order{
		{
			Creator:    acc2.String(),
			Type:       dextypes.ORDER_TYPE_LIMIT,
			ID:         "id1",
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(dextypes.MustNewPriceFromString("996")),
			Quantity:   sdkmath.NewInt(10_000),
			Side:       dextypes.SIDE_BUY,
			GoodTil: &dextypes.GoodTil{
				GoodTilBlockHeight: uint64(latestBlock.Height + 1000),
			},
			TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
			RemainingBaseQuantity:     sdkmath.NewInt(10_000),
			RemainingSpendableBalance: sdkmath.NewInt(9_960_000),
			Reserve:                   dexParamsRes.Params.OrderReserve,
		},
		{
			Creator:                   acc2.String(),
			Type:                      dextypes.ORDER_TYPE_LIMIT,
			ID:                        "id2",
			BaseDenom:                 denom1,
			QuoteDenom:                denom2,
			Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("997")),
			Quantity:                  sdkmath.NewInt(10_000),
			Side:                      dextypes.SIDE_BUY,
			TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
			RemainingBaseQuantity:     sdkmath.NewInt(10_000),
			RemainingSpendableBalance: sdkmath.NewInt(9_970_000),
			Reserve:                   dexParamsRes.Params.OrderReserve,
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
	acc2Orders = fillOrderSequences(ctx, t, chain.ClientContext, acc2Orders)

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

	orderBookParamsRes, err := dexClient.OrderBookParams(ctx, &dextypes.QueryOrderBookParamsRequest{
		BaseDenom:  denom1,
		QuoteDenom: denom2,
	})
	requireT.NoError(err)
	require.Equal(t,
		fmt.Sprintf("1e%d", dexParamsRes.Params.PriceTickExponent), // 1e-6
		orderBookParamsRes.PriceTick.String(),
	)
	require.Equal(t,
		"10000", orderBookParamsRes.QuantityStep.String(),
	)
	require.Equal(
		t, dexParamsRes.Params.DefaultUnifiedRefAmount.String(), orderBookParamsRes.BaseDenomUnifiedRefAmount.String(),
	)
	require.Equal(
		t, dexParamsRes.Params.DefaultUnifiedRefAmount.String(), orderBookParamsRes.QuoteDenomUnifiedRefAmount.String(),
	)

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
	newParams.QuantityStepExponent = -15
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

	acc1 := chain.GenAccount()

	issuer, denom1 := genAccountAndIssueFT(
		ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_freezing,
	)
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgFreeze{},
					&assetfttypes.MsgUnfreeze{},
					&assetfttypes.MsgFreeze{},
				},
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
			},
		},
	})

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150_000)),
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
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(150_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())

	// place order should fail because all the funds are frozen
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())

	// change the frozen amount to less than the order quantity
	unfreezeMsg := &assetfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(50_000).String(), balanceRes.Frozen.String())

	// now placing order should succeed because the needed funds are more than frozen amount
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(50_000).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// freeze remaining tokens
	freezeMsg = &assetfttypes.MsgFreeze{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300_000),
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

	assertBalance(ctx, t, bankClient, acc1, denom2, 10_000)
	assertBalance(ctx, t, bankClient, acc2, denom1, 100_000)
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

	acc1 := chain.GenAccount()

	issuer, denom1 := genAccountAndIssueFT(
		ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_freezing,
	)
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgGloballyFreeze{},
					&assetfttypes.MsgGloballyUnfreeze{},
					&assetfttypes.MsgGloballyFreeze{},
					&assetfttypes.MsgGloballyUnfreeze{},
				},
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
			},
		},
	})

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())

	// place order should fail because all the funds are globally frozen
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())
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
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Frozen.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300_000),
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

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	requireT.NoError(err)

	assertBalance(ctx, t, bankClient, acc1, denom2, 10_000)
	assertBalance(ctx, t, bankClient, acc2, denom1, 100_000)
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
	acc1 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
					&assetfttypes.MsgClawback{},
					&banktypes.MsgSend{},
					&assetfttypes.MsgClawback{},
					&assetfttypes.MsgClawback{},
				},
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&dextypes.MsgCancelOrder{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(2).Add(sdkmath.NewInt(100_000).MulRaw(2)),
			},
		},
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_clawback)
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6))

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(150_000)),
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
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(50_000).String(), balanceRes.Balance.String())

	// place order should fail because of insufficient funds
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(50_000).String(), balanceRes.Balance.String())

	// send enough amounts for the order
	msgSend = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Balance.String())

	// now placing order should succeed because the needed funds are available
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Balance.String())
	requireT.Equal(placeSellOrderMsg.Quantity.String(), balanceRes.LockedInDEX.String())

	// try to clawback after placing the order should fail
	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Balance.String())
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.ZeroInt().String(), balanceRes.LockedInDEX.String())

	// now clawback should succeed
	clawbackMsg = &assetfttypes.MsgClawback{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(100_000)),
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
	requireT.Equal(sdkmath.NewInt(50_000).String(), balanceRes.Balance.String())
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
	acc2 := chain.GenAccount()

	issueFeeAmount := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: acc1,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgIssue{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)).Add(issueFeeAmount),
			},
		}, {
			Acc: acc2,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
			},
		},
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
		Quantity:    sdkmath.NewInt(100_000),
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
		Creator:                   acc1.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  sellOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:                  sdkmath.NewInt(100_000),
		Side:                      dextypes.SIDE_SELL,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(100_000),
		RemainingSpendableBalance: sdkmath.NewInt(100_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300_000),
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

	// burn rate is not applied to receiver account and full amount is received
	assertBalance(ctx, t, bankClient, acc2, denom1, 100_000)
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
	acc2 := chain.GenAccount()

	issueFeeAmount := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: acc1,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgIssue{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)).Add(issueFeeAmount),
			},
		}, {
			Acc: acc2,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
			},
		},
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
		Quantity:    sdkmath.NewInt(100_000),
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
		Creator:                   acc1.String(),
		Type:                      dextypes.ORDER_TYPE_LIMIT,
		ID:                        "id1",
		Sequence:                  sellOrderRes.Order.Sequence,
		BaseDenom:                 denom1,
		QuoteDenom:                denom2,
		Price:                     lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:                  sdkmath.NewInt(100_000),
		Side:                      dextypes.SIDE_SELL,
		TimeInForce:               dextypes.TIME_IN_FORCE_GTC,
		RemainingBaseQuantity:     sdkmath.NewInt(100_000),
		RemainingSpendableBalance: sdkmath.NewInt(100_000),
		Reserve:                   dexParamsRes.Params.OrderReserve,
	}, sellOrderRes.Order)

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300_000),
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

	// commission rate is not applied to receiver account and full amount is received
	assertBalance(ctx, t, bankClient, acc2, denom1, 100_000)
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

	acc1 := chain.GenAccount()

	issuer, denom1 := genAccountAndIssueFT(
		ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_whitelisting,
	)
	acc2, denom2 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgSetWhitelistedLimit{},
					&assetfttypes.MsgSetWhitelistedLimit{},
					&banktypes.MsgSend{},
					&assetfttypes.MsgSetWhitelistedLimit{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(100_000)),
			},
		},
	})

	// whitelist denom1 for acc1
	msgSetWhitelistedLimit := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: acc1.String(),
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(1_000_000)),
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
			sdk.NewCoin(denom1, sdkmath.NewInt(150_000)),
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
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.Balance.String())

	// place order should fail because acc2 is out of whitelisted coins
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc2.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("11e-2")),
		Quantity:    sdkmath.NewInt(300_000),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
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
		Coin:    sdk.NewCoin(denom1, sdkmath.NewInt(300_000)),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgSetWhitelistedLimit)),
		msgSetWhitelistedLimit,
	)
	requireT.NoError(err)

	// now placing of the same order should succeed because the receiving amount is within the whitelist limit
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
	requireT.Equal(sdkmath.NewInt(33_000).String(), balanceRes.LockedInDEX.String())

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
	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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

	assertBalance(ctx, t, bankClient, acc1, denom2, 11_000)
	assertBalance(ctx, t, bankClient, acc2, denom1, 100_000)

	// place order should succeed as issuer because admin (issuer) doesn't have whitelist limit
	placeSellOrderMsg = &dextypes.MsgPlaceOrder{
		Sender:      issuer.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1e-1")),
		Quantity:    sdkmath.NewInt(100_000),
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

	issuer := chain.GenAccount()
	acc1 := chain.GenAccount()
	acc2 := chain.GenAccount()

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	ordersCount := int(dexParamsRes.Params.MaxOrdersPerDenom)

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
				},
				Amount: sdkmath.NewIntWithDecimal(1, 6), // amount to cover cancellation
			},
		}, {
			Acc: acc2,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewIntWithDecimal(1, 6), // amount to cover cancellation
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(int64(ordersCount)).AddRaw(100_000 * int64(ordersCount)),
			},
		},
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 10), assetfttypes.Feature_dex_order_cancellation)
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 10), assetfttypes.Feature_dex_order_cancellation)

	latestBlock, err := chain.LatestBlockHeader(ctx)
	requireT.NoError(err)

	amtPerOrder := sdkmath.NewInt(100_000)
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
				GoodTilBlockHeight: uint64(latestBlock.Height + 20_000),
				GoodTilBlockTime:   lo.ToPtr(latestBlock.Time.Add(time.Hour)),
			}),
			TimeInForce: dextypes.TIME_IN_FORCE_GTC,
		}
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

	// Cancelation from non-issuer and order creator accounts should fail.
	for _, sender := range []sdk.AccAddress{acc1, acc2} {
		_, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(sender),
			chain.TxFactoryAuto(),
			&dextypes.MsgCancelOrdersByDenom{
				Sender:  sender.String(),
				Account: acc1.String(),
				Denom:   denom2,
			})
		requireT.Error(err)
		requireT.ErrorContains(err, "only admin is able to cancel orders by denom")
	}

	// Cancellation from issuer account succeeds.
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
	acc := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgIssue{},
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
					&banktypes.MsgSend{},
				},
				Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2),
			},
		}, {
			Acc: acc,
			Options: integration.BalancesOptions{
				// 2 to place directly and 1 through the smart contract
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(3).
					Add(sdkmath.NewInt(100_000).MulRaw(3)).
					Add(chain.QueryDEXParams(ctx, t).OrderReserve.Amount),
			},
		},
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom1WithBlockSmartContract, sdkmath.NewInt(100_000))),
	}
	sendMsg2 := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom2, sdkmath.NewInt(100_000))),
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
		Quantity:    sdkmath.NewInt(100_000),
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
		Quantity:    sdkmath.NewInt(100_000),
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
		InitialAmount: sdkmath.NewInt(100_000).String(),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom1WithBlockSmartContract, sdkmath.NewInt(100_000))),
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
		Amount:      sdk.NewCoins(sdk.NewCoin(denom2, sdkmath.NewInt(100_000))),
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
			Order: dextypes.MsgPlaceOrder{
				Sender:      contractAddr,
				Type:        dextypes.ORDER_TYPE_LIMIT,
				ID:          "id1",
				BaseDenom:   denom1WithBlockSmartContract,
				QuoteDenom:  denom2,
				Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
				Quantity:    sdkmath.NewInt(100_000),
				Side:        dextypes.SIDE_BUY,
				TimeInForce: dextypes.TIME_IN_FORCE_GTC,
			},
		},
	})
	requireT.NoError(err)

	// Even though, the contract has enough balance to place such and order, the placement is failed because
	// the order expects to receive the asset ft with block_smart_contracts feature
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
	acc1 := chain.GenAccount()

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: issuer,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&banktypes.MsgSend{},
				},
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				Messages: []sdk.Msg{
					&assetfttypes.MsgBurn{},
					&dextypes.MsgCancelOrder{},
					&assetfttypes.MsgBurn{},
					&assetfttypes.MsgBurn{},
				},
				Amount: dexParamsRes.Params.OrderReserve.Amount.MulRaw(1).Add(sdkmath.NewInt(100_000)),
			},
		},
	})

	denom1 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_burning)
	denom2 := issueFT(ctx, t, chain, issuer, sdkmath.NewIntWithDecimal(1, 6))

	msgSend := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   acc1.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom1, sdkmath.NewInt(200_000)),
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
		Quantity:    sdkmath.NewInt(150_000),
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
	requireT.Equal(sdkmath.NewInt(200_000).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.NewInt(150_000).String(), balanceRes.LockedInDEX.String())

	// try to burn burnable token, locked in dex
	burnMsg := &assetfttypes.MsgBurn{
		Sender: acc1.String(),
		Coin: sdk.Coin{
			Denom: denom1,
			// it's allowed to burn only 50_000 not locked in dex
			Amount: sdkmath.NewInt(100_000),
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.ErrorContains(err, "is not available, available")

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
	// 100k is burnt 100k remains
	requireT.Equal(sdkmath.NewInt(100_000).String(), balanceRes.Balance.String())
	requireT.Equal(sdkmath.NewInt(0).String(), balanceRes.LockedInDEX.String())
}

// TestDEXReentrancyViaExtension tests to make sure that reentrancy bug does not exist in DEX. It might happen
// if the control is given to smart contract in middle of an order placement.
func TestDEXReentrancyViaExtension(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	assetFTClint := assetfttypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	dexReserver := dexParamsRes.Params.OrderReserve

	admin := chain.GenAccount()
	acc1 := chain.GenAccount()
	acc2 := chain.GenAccount()

	issueFeeAmount := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	chain.FundAccountsWithOptions(ctx, t, []integration.AccWithBalancesOptions{
		{
			Acc: admin,
			Options: integration.BalancesOptions{
				Amount: issueFeeAmount.
					AddRaw(10_000_000).
					Add(dexReserver.Amount),
			},
		}, {
			Acc: acc1,
			Options: integration.BalancesOptions{
				// message + order reserve
				Amount: sdkmath.NewInt(5_000_000).
					Add(dexReserver.Amount.MulRaw(3)),
			},
		}, {
			Acc: acc2,
			Options: integration.BalancesOptions{
				Amount: sdkmath.NewInt(5_000_000).
					AddRaw(1_000_000).
					Add(dexReserver.Amount), // message + balance to place an order + order reserve
			},
		},
	})

	codeID, err1 := chain.Wasm.DeployWASMContract(
		ctx, chain.TxFactory().WithSimulateAndExecute(true), admin, testcontracts.DexReentrancyPocWasm,
	)
	requireT.NoError(err1)

	// issue tokenA
	denomA := issueFT(ctx, t, chain, admin, sdkmath.NewIntWithDecimal(1, 8))

	//nolint:tagliatelle // these will be exposed to rust and must be snake case.
	issuanceMsg := struct {
		ExtraData string `json:"extra_data"`
	}{
		ExtraData: denomA,
	}

	issuanceMsgBytes, err := json.Marshal(issuanceMsg)
	requireT.NoError(err)

	attachedFund := chain.NewCoin(sdkmath.NewInt(1_000_000))
	issueMsgB := &assetfttypes.MsgIssue{
		Issuer:        admin.String(),
		Symbol:        "TKNB",
		Subunit:       "ub",
		Precision:     6,
		Description:   "TKNB Description",
		InitialAmount: sdkmath.NewIntWithDecimal(1, 8),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_extension,
		},
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		ExtensionSettings: &assetfttypes.ExtensionIssueSettings{
			CodeId:      codeID,
			Funds:       sdk.NewCoins(attachedFund),
			Label:       "testing-reentrancy-bug",
			IssuanceMsg: issuanceMsgBytes,
		},
	}

	denomB := assetfttypes.BuildDenom(issueMsgB.Subunit, admin)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		issueMsgB,
	)
	requireT.NoError(err)
	// get extension contract addr
	denomBTokenRes, err := assetFTClint.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denomB,
	})
	requireT.NoError(err)
	tokenBExtensionAddress := denomBTokenRes.Token.ExtensionCWAddress

	sendMsgs := make([]sdk.Msg, 0)
	// send from admin to cw to place an order
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   tokenBExtensionAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin("udevcore", dexReserver.Amount)),
	})
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   tokenBExtensionAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(denomA, sdkmath.NewInt(5_000_000))),
	})
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   tokenBExtensionAddress,
		Amount:      sdk.NewCoins(sdk.NewCoin(denomB, sdkmath.NewInt(5_000_000))),
	})
	// send from admin to acc1 to place an order
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denomA, sdkmath.NewInt(5_000_000))),
	})
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denomB, sdkmath.NewInt(5_000_000))),
	})
	// send from admin to acc2 to place an order
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   acc2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denomA, sdkmath.NewInt(5_000_000))),
	})
	sendMsgs = append(sendMsgs, &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   acc2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denomB, sdkmath.NewInt(5_000_000))),
	})
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		sendMsgs...,
	)
	requireT.NoError(err)

	// place 3 SELL order from acc1(SELL A(without extension) BUY B(with extension))
	_, err = placeOrder(ctx, chain, acc1, dextypes.SIDE_SELL, "id1", denomA, denomB, "1", sdkmath.NewInt(1_200_000))
	requireT.NoError(err)
	_, err = placeOrder(ctx, chain, acc1, dextypes.SIDE_SELL, "id2", denomA, denomB, "1", sdkmath.NewInt(1_200_000))
	requireT.NoError(err)
	_, err = placeOrder(ctx, chain, acc1, dextypes.SIDE_SELL, "id3", denomA, denomB, "1e2", sdkmath.NewInt(990_000))
	requireT.NoError(err)

	// place BUY order from acc2(BUY A(without extension) SELL B(with extension))
	_, err = placeOrder(ctx, chain, acc2, dextypes.SIDE_BUY, "hackid0", denomA, denomB, "1e1", sdkmath.NewInt(1_000_000))
	requireT.NoError(err)

	acc1ABalanceRes, err := assetFTClint.Balance(ctx, &assetfttypes.QueryBalanceRequest{
		Account: acc1.String(),
		Denom:   denomA,
	})
	requireT.NoError(err)
	// if reentrancy bug exists, orders might match multiple times, and LockedInDEX might be 89000000
	requireT.Equal("990000", acc1ABalanceRes.LockedInDEX.String())
}

func placeOrder(
	ctx context.Context,
	chain integration.CoreumChain,
	acc sdk.AccAddress,
	side dextypes.Side,
	id, denomA, denomB, price string,
	amount sdkmath.Int,
) (*sdk.TxResponse, error) {
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          id,
		BaseDenom:   denomA,
		QuoteDenom:  denomB,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString(price)),
		Quantity:    amount,
		Side:        side,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}
	return client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
}

// TestTradeByAdmin tests trades by admin without limits like frozen amount.
func TestTradeByAdmin(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	dexParamsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)

	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Amount: dexParamsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(10_000_000)),
	})

	acc1, denom1 := genAccountAndIssueFT(ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6))
	acc2, denom2 := genAccountAndIssueFT(
		ctx, t, chain, 10_000_000, sdkmath.NewIntWithDecimal(1, 6), assetfttypes.Feature_freezing,
	)

	msgSend := &banktypes.MsgSend{
		FromAddress: acc2.String(),
		ToAddress:   admin.String(),
		Amount: sdk.NewCoins(
			sdk.NewCoin(denom2, sdkmath.NewInt(1_000_000)),
		),
	}

	// freeze whole tokens before becoming admin
	freezeMsg := &assetfttypes.MsgFreeze{
		Sender:  acc2.String(),
		Account: admin.String(),
		Coin:    sdk.NewCoin(denom2, sdkmath.NewInt(1_000_000)),
	}

	// Transfer administration of fungible token
	transferAdminMsg := &assetfttypes.MsgTransferAdmin{
		Sender:  acc2.String(),
		Account: admin.String(),
		Denom:   denom2,
	}

	msgList := []sdk.Msg{
		msgSend, freezeMsg, transferAdminMsg,
	}

	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)

	requireT.NoError(err)

	adminTransferredEvts, err := event.FindTypedEvents[*assetfttypes.EventAdminTransferred](res.Events)
	requireT.NoError(err)
	requireT.Equal(acc2.String(), adminTransferredEvts[0].PreviousAdmin)
	requireT.Equal(admin.String(), adminTransferredEvts[0].CurrentAdmin)

	placeSellOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      acc1.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(1_000_000),
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

	// place buy order to match the sell
	placeBuyOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:      admin.String(),
		Type:        dextypes.ORDER_TYPE_LIMIT,
		ID:          "id1", // same ID allowed for different user
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
		Quantity:    sdkmath.NewInt(1_000_000),
		Side:        dextypes.SIDE_BUY,
		TimeInForce: dextypes.TIME_IN_FORCE_GTC,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactoryAuto(),
		placeBuyOrderMsg,
	)
	// regardless of frozen amount, admin can trade
	requireT.NoError(err)
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

//nolint:unparam // using constant values here will make this function less flexible.
func genAccountAndIssueFT(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	balance int64,
	initialIssueAmount sdkmath.Int,
	features ...assetfttypes.Feature,
) (sdk.AccAddress, string) {
	issuer := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
		},
		Amount: chain.QueryDEXParams(ctx, t).OrderReserve.Amount.
			Add(chain.QueryAssetFTParams(ctx, t).IssueFee.Amount).
			Add(sdkmath.NewInt(balance)),
	})
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "TKN" + uuid.NewString()[:4],
		Subunit:       "tkn" + uuid.NewString()[:4],
		Precision:     5,
		InitialAmount: initialIssueAmount,
		Features:      features,
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)
	return issuer, assetfttypes.BuildDenom(issueMsg.Subunit, issuer)
}

func fillOrderSequences(
	ctx context.Context,
	t *testing.T,
	clientCtx client.Context,
	orders []dextypes.Order,
) []dextypes.Order {
	dexClient := dextypes.NewQueryClient(clientCtx)
	for i, order := range orders {
		res, err := dexClient.Order(ctx, &dextypes.QueryOrderRequest{
			Creator: order.Creator,
			Id:      order.ID,
		})
		require.NoError(t, err)
		orders[i].Sequence = res.Order.Sequence
	}

	return orders
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

func assertBalance(
	ctx context.Context,
	t *testing.T,
	bankClient banktypes.QueryClient,
	acc sdk.AccAddress,
	denom string,
	expectedBalance int64,
) {
	acc1Denom2BalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc.String(),
		Denom:   denom,
	})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(expectedBalance).String(), acc1Denom2BalanceRes.Balance.Amount.String())
}
