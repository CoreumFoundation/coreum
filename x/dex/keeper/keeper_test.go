package keeper_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

func TestKeeper_UpdateParams(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	gotParams := dexKeeper.GetParams(sdkCtx)
	require.Equal(t, types.DefaultParams(), gotParams)

	newPrams := gotParams
	newPrams.DefaultUnifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("33.33")
	newPrams.PriceTickExponent = -33

	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	// try to update for random address
	require.ErrorIs(t, dexKeeper.UpdateParams(sdkCtx, randomAddr.String(), newPrams), govtypes.ErrInvalidSigner)

	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	require.NoError(t, dexKeeper.UpdateParams(sdkCtx, govAddr, newPrams))
	gotParams = dexKeeper.GetParams(sdkCtx)
	require.Equal(t, newPrams, gotParams)
}

func TestKeeper_PlaceOrder_OrderBookIDs(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)

	type denomsToOrderBookIDs struct {
		baseDenom                   string
		quoteDenom                  string
		expectedSelfOrderBookID     uint32
		expectedOppositeOrderBookID uint32
	}

	for _, item := range []denomsToOrderBookIDs{
		// save with asc denoms ordering
		{
			baseDenom:                   denom1,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(0),
			expectedOppositeOrderBookID: uint32(1),
		},
		// save one more time to check that returns the same
		{
			baseDenom:                   denom1,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(0),
			expectedOppositeOrderBookID: uint32(1),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom1,
			expectedSelfOrderBookID:     uint32(1),
			expectedOppositeOrderBookID: uint32(0),
		},
		// save with desc denoms ordering
		{
			baseDenom:                   denom3,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(3),
			expectedOppositeOrderBookID: uint32(2),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom3,
			expectedSelfOrderBookID:     uint32(2),
			expectedOppositeOrderBookID: uint32(3),
		},
	} {
		price := types.MustNewPriceFromString("1")
		acc, _ := testApp.GenAccount(sdkCtx)
		order := types.Order{
			Creator:     acc.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   item.baseDenom,
			QuoteDenom:  item.quoteDenom,
			Price:       &price,
			Quantity:    sdkmath.NewInt(1),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		}
		lockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))

		require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
		selfOrderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, item.baseDenom, item.quoteDenom)
		require.NoError(t, err)
		oppositeOrderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, item.quoteDenom, item.baseDenom)
		require.NoError(t, err)

		require.Equal(t, item.expectedSelfOrderBookID, selfOrderBookID)
		require.Equal(t, item.expectedOppositeOrderBookID, oppositeOrderBookID)
	}
}

func TestKeeper_PlaceAndGetOrderByID(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	price := lo.ToPtr(types.MustNewPriceFromString("12e-1"))
	acc, _ := testApp.GenAccount(sdkCtx)

	sellOrder := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      price,
		Quantity:   sdkmath.NewInt(10),
		Side:       types.SIDE_SELL,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: 390,
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err := sellOrder.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))

	// try to place the sellOrder one more time
	err = dexKeeper.PlaceOrder(sdkCtx, sellOrder)
	require.ErrorIs(t, err, types.ErrInvalidInput)
	require.ErrorContains(t, err, "is already created")

	gotOrder, err := dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(sellOrder.Creator), sellOrder.ID,
	)
	require.NoError(t, err)

	// set expected values
	sellOrder.RemainingQuantity = sdkmath.NewInt(10)
	sellOrder.RemainingBalance = sdkmath.NewInt(10)
	require.Equal(t, sellOrder, gotOrder)

	// check same buy with the buy order

	buyOrder := types.Order{
		Creator:     acc.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom2,
		QuoteDenom:  denom3,
		Price:       price,
		Quantity:    sdkmath.NewInt(100),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	lockedBalance, err = buyOrder.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))

	gotOrder, err = dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(buyOrder.Creator), buyOrder.ID,
	)
	require.NoError(t, err)

	// set expected values
	buyOrder.RemainingQuantity = sdkmath.NewInt(100)
	buyOrder.RemainingBalance = sdkmath.NewInt(120)
	require.Equal(t, buyOrder, gotOrder)
}

func TestKeeper_PlaceAndCancelOrder(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height: 100,
		Time:   time.Date(2023, 3, 2, 1, 11, 12, 13, time.UTC),
	})
	dexKeeper := testApp.DEXKeeper
	assetFTKeeper := testApp.AssetFTKeeper

	acc, _ := testApp.GenAccount(sdkCtx)

	sellOrder := types.Order{
		Creator:     acc.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err := sellOrder.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(sellLockedBalance))

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))
	dexLockedBalance := assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.Equal(t, sellLockedBalance.String(), dexLockedBalance.String())

	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, sellOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())

	buyOrder := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  denom1,
		QuoteDenom: denom2,
		Price:      lo.ToPtr(types.MustNewPriceFromString("13e-1")),
		Quantity:   sdkmath.NewInt(5_000),
		Side:       types.SIDE_BUY,
		GoodTil: &types.GoodTil{
			GoodTilBlockHeight: uint64(sdkCtx.BlockHeight() + 1),
			GoodTilBlockTime:   lo.ToPtr(sdkCtx.BlockTime().Add(time.Second)),
		},
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	buyLockedBalance, err := buyOrder.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(buyLockedBalance))

	// try to place order with the invalid GoodTilBlockHeight
	buyOrderWithGoodTilHeight := buyOrder
	buyOrderWithGoodTilHeight.GoodTil = &types.GoodTil{
		GoodTilBlockHeight: uint64(sdkCtx.BlockHeight() - 1),
	}
	require.ErrorContains(
		t,
		dexKeeper.PlaceOrder(sdkCtx, buyOrderWithGoodTilHeight),
		"good til block height 99 must be greater than current block height 100: invalid input",
	)

	// try to place order with the invalid GoodTilBlockTime
	buyOrderWithGoodTilTime := buyOrder
	buyOrderWithGoodTilTime.GoodTil = &types.GoodTil{
		GoodTilBlockTime: lo.ToPtr(sdkCtx.BlockTime()),
	}
	require.ErrorContains(
		t,
		dexKeeper.PlaceOrder(sdkCtx, buyOrderWithGoodTilTime),
		"good til block height 2023-03-02 01:11:12.000000013 +0000 UTC must be greater than current block height",
	)

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.Equal(t, buyLockedBalance.String(), dexLockedBalance.String())

	// check unlocking
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())

	// now place both orders to let them match partially
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))

	_, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, sellOrder.ID)
	require.ErrorIs(t, err, types.ErrRecordNotFound)
	buyOrder, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, buyOrder.ID)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(5300).String(), buyOrder.RemainingBalance.String())
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())
}

func TestKeeper_PlaceOrderWithPriceTick(t *testing.T) {
	tests := []struct {
		name                string
		price               types.Price
		baseDenomRefAmount  *sdkmath.LegacyDec
		quoteDenomRefAmount *sdkmath.LegacyDec
		wantTickError       bool
	}{
		{
			name:          "valid_default_price",
			price:         types.MustNewPriceFromString("1e-5"),
			wantTickError: false,
		},
		{
			name:          "invalid_default_price",
			price:         types.MustNewPriceFromString("1e-6"),
			wantTickError: true,
		},
		{
			name:               "valid_base_custom",
			price:              types.MustNewPriceFromString("33e-6"),
			baseDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10000000")),
			wantTickError:      false,
		},
		{
			name:                "valid_quote_custom",
			price:               types.MustNewPriceFromString("1e-6"),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000")),
			wantTickError:       false,
		},
		{
			name:                "valid_both_custom",
			price:               types.MustNewPriceFromString("14e-3"),
			baseDenomRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("1")),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100")),
			wantTickError:       false,
		},
		{
			name:                "valid_both_custom_tick_greater_than_one",
			price:               types.MustNewPriceFromString("14e1"),
			baseDenomRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.01")),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10303.3")),
			wantTickError:       false,
		},
		{
			name:                "invalid_both_custom_tick_greater_than_one",
			price:               types.MustNewPriceFromString("14"),
			baseDenomRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.01")),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10303.3")),
			wantTickError:       true,
		},
		{
			name:                "valid_both_custom_base_less_than_one",
			price:               types.MustNewPriceFromString("3e33"),
			baseDenomRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000000000000000000")),
			wantTickError:       false,
		},
		{
			name:                "invalid_both_custom_base_less_than_one",
			price:               types.MustNewPriceFromString("3e32"),
			baseDenomRefAmount:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
			quoteDenomRefAmount: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000000000000000000")),
			wantTickError:       true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testApp := simapp.New()
			sdkCtx := testApp.BaseApp.NewContext(false)

			if tt.baseDenomRefAmount != nil {
				testApp.AssetFTKeeper.SetDEXSettings(sdkCtx, denom1, assetfttypes.DEXSettings{
					UnifiedRefAmount: *tt.baseDenomRefAmount,
				})
			}

			if tt.quoteDenomRefAmount != nil {
				testApp.AssetFTKeeper.SetDEXSettings(sdkCtx, denom2, assetfttypes.DEXSettings{
					UnifiedRefAmount: *tt.quoteDenomRefAmount,
				})
			}

			acc, _ := testApp.GenAccount(sdkCtx)
			order := types.Order{
				Creator:     acc.String(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denom1,
				QuoteDenom:  denom2,
				Price:       &tt.price,
				Quantity:    sdkmath.NewInt(1_000),
				Side:        types.SIDE_SELL,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			}
			lockedBalance, err := order.ComputeLimitOrderLockedBalance()
			require.NoError(t, err)
			testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))
			err = testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
			if tt.wantTickError {
				require.ErrorContains(t, err, "the price must be multiple of")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestKeeper_GetOrdersAndOrderBookOrders(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)

	orders := []types.Order{
		{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denom1,
			QuoteDenom:  denom2,
			Price:       lo.ToPtr(types.MustNewPriceFromString("13e-1")),
			Quantity:    sdkmath.NewInt(2000),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
		{
			Creator:    acc1.String(),
			Type:       types.ORDER_TYPE_LIMIT,
			ID:         uuid.Generate().String(),
			BaseDenom:  denom3,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(types.MustNewPriceFromString("14e-1")),
			Quantity:   sdkmath.NewInt(3000),
			Side:       types.SIDE_BUY,
			GoodTil: &types.GoodTil{
				GoodTilBlockHeight: 32,
			},
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
		{
			Creator:    acc1.String(),
			Type:       types.ORDER_TYPE_LIMIT,
			ID:         uuid.Generate().String(),
			BaseDenom:  denom1,
			QuoteDenom: denom2,
			Price:      lo.ToPtr(types.MustNewPriceFromString("12e-1")),
			Quantity:   sdkmath.NewInt(1000),
			Side:       types.SIDE_SELL,
			GoodTil: &types.GoodTil{
				GoodTilBlockHeight: 1000,
			},
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
		{
			Creator:     acc2.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denom1,
			QuoteDenom:  denom2,
			Price:       lo.ToPtr(types.MustNewPriceFromString("11e-1")),
			Quantity:    sdkmath.NewInt(100),
			Side:        types.SIDE_BUY,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
	}

	for i, order := range orders {
		lockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))
		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
		// fill order with the remaining quantity for assertions
		order.RemainingQuantity = order.Quantity
		order.RemainingBalance = lockedBalance.Amount
		orders[i] = order
	}

	// get account orders
	acc1Orders, pageRes, err := testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Offset:     0,
		Limit:      2,
		CountTotal: true,
	})
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(3), pageRes.Total)
	require.Len(t, acc1Orders, 2)

	acc1Orders, _, err = testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Limit: query.PaginationMaxLimit,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []types.Order{
		orders[0], orders[1], orders[2],
	}, acc1Orders)

	// get order book orders
	denom1To2Orders, pageRes, err := testApp.DEXKeeper.GetOrderBookOrders(
		sdkCtx,
		denom1,
		denom2,
		types.SIDE_SELL,
		&query.PageRequest{
			Offset:     0,
			Limit:      1,
			CountTotal: true,
		},
	)
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(2), pageRes.Total)
	require.Len(t, denom1To2Orders, 1)

	denom1To2Orders, _, err = testApp.DEXKeeper.GetOrderBookOrders(
		sdkCtx,
		denom1,
		denom2,
		types.SIDE_SELL,
		&query.PageRequest{
			Limit: query.PaginationMaxLimit,
		},
	)
	require.NoError(t, err)
	require.ElementsMatch(t, []types.Order{
		orders[0], orders[2],
	}, denom1To2Orders)
}

func TestKeeper_GetOrderBooks(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	acc1, _ := testApp.GenAccount(sdkCtx)

	orders := []types.Order{
		{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denom1,
			QuoteDenom:  denom2,
			Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
			Quantity:    sdkmath.NewInt(10),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
		{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denom3,
			QuoteDenom:  denom2,
			Price:       lo.ToPtr(types.MustNewPriceFromString("13e-1")),
			Quantity:    sdkmath.NewInt(10),
			Side:        types.SIDE_BUY,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		},
	}

	for _, order := range orders {
		lockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), sdk.NewCoins(lockedBalance))
		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
	}

	orderBooks, pageRes, err := testApp.DEXKeeper.GetOrderBooks(sdkCtx, &query.PageRequest{
		Offset:     0,
		Limit:      3,
		CountTotal: true,
	})
	require.NoError(t, err)
	require.NotNil(t, pageRes.NextKey)
	require.Equal(t, uint64(4), pageRes.Total)
	require.Equal(t, []types.OrderBookData{
		{
			BaseDenom:  denom1,
			QuoteDenom: denom2,
		},
		{
			BaseDenom:  denom2,
			QuoteDenom: denom1,
		},
		{
			BaseDenom:  denom2,
			QuoteDenom: denom3,
		},
	}, orderBooks)
}

func TestKeeper_ComputePriceTick(t *testing.T) {
	tests := []struct {
		name  string
		base  float64
		quote float64
	}{
		{
			name:  "3.0/27.123",
			base:  3.0,
			quote: 27.123,
		},

		{
			name:  "10000.0/10000.0",
			base:  10000.0,
			quote: 10000.0,
		},
		{
			name:  "3000.0/20.0",
			base:  3000.0,
			quote: 20.0,
		},
		{
			name:  "300000.0/20.0",
			base:  300000.0,
			quote: 20.0,
		},
		{
			name:  "2.0/2.0",
			base:  2.0,
			quote: 2.0,
		},
		{
			name:  "100.0/1.0",
			base:  100.0,
			quote: 1.0,
		},
		{
			name:  "3.0/1.0",
			base:  3.0,
			quote: 1.0,
		},

		{
			name:  "3100000.0/8.0",
			base:  3100000.0,
			quote: 8.0,
		},
		{
			name:  "0.00017/100",
			base:  0.00017,
			quote: 100,
		},
		{
			name:  "0.000001/10000000",
			base:  0.000001,
			quote: 10000000,
		},
		{
			name:  "100/1000000000000",
			base:  100,
			quote: 1000000000000,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assertTickCalculations(t, tt.base, tt.quote)
			assertTickCalculations(t, tt.quote, tt.base)
		})
	}
}

func TestKeeper_PlaceAndCancelOrderWithMaxAllowedAccountDenomOrdersCount(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{
		Height: 100,
		Time:   time.Date(2023, 3, 2, 1, 11, 12, 13, time.UTC),
	})

	params := testApp.DEXKeeper.GetParams(sdkCtx)
	params.MaxOrdersPerDenom = 2
	require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

	acc1, _ := testApp.GenAccount(sdkCtx)
	acc2, _ := testApp.GenAccount(sdkCtx)

	order1 := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err := order1.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order1))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denom1: 1,
			denom2: 1,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	order2 := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom2,
		QuoteDenom:  denom3,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err = order2.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order2))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denom1: 1,
			denom2: 2,
			denom3: 1,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	// create order to reach max allowed limit for all denom
	order3 := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom3,
		QuoteDenom:  denom1,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err = order3.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order3))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denom1: 2,
			denom2: 2,
			denom3: 2,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	// try to create one more order to exceed the limit
	// create order to reach max allowed limit for denom1
	trialCtx := simapp.CopyContextWithMultiStore(sdkCtx) // copy in order not to affect the state by the error
	order4 := types.Order{
		Creator:     acc1.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom3,
		QuoteDenom:  denom1,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err = order4.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, trialCtx, acc1, sdk.NewCoins(sellLockedBalance))
	require.ErrorContains(t,
		testApp.DEXKeeper.PlaceOrder(trialCtx, order4),
		"it's prohibited to save more than 2 orders per denom",
	)

	// cancel the order1 VIA matching
	order5 := types.Order{
		Creator:     acc2.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(10_000),
		Side:        types.SIDE_BUY,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err = order5.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc2, sdk.NewCoins(sellLockedBalance))

	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order5))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denom1: 1,
			denom2: 1,
			denom3: 2,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	// cancel order manually
	require.NoError(t, testApp.DEXKeeper.CancelOrder(sdkCtx, acc1, order2.ID))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denom1: 1,
			denom2: 0,
			denom3: 1,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))
}

func getSorterOrderBookOrders(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) []types.Order {
	records := getSorterOrderBookRecords(t, testApp, sdkCtx, orderBookID, side)
	orders := make([]types.Order, 0, len(records))
	authQueryServer := authkeeper.NewQueryServer(testApp.AccountKeeper)
	for _, record := range records {
		resp, err := authQueryServer.AccountAddressByID(
			sdkCtx,
			&authtypes.QueryAccountAddressByIDRequest{AccountId: record.AccountNumber},
		)
		require.NoError(t, err)
		addr := sdk.MustAccAddressFromBech32(resp.AccountAddress)
		order, err := testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, addr, record.OrderID)
		require.NoError(t, err)
		orders = append(orders, order)
	}

	return orders
}

func getSorterOrderBookRecords(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	orderBookID uint32,
	side types.Side,
) []types.OrderBookRecord {
	records := make([]types.OrderBookRecord, 0)
	iterator := testApp.DEXKeeper.NewOrderBookSideIterator(sdkCtx, orderBookID, side)
	defer iterator.Close()

	for {
		record, found, err := iterator.Next()
		require.NoError(t, err)
		if !found {
			break
		}
		records = append(records, record)
	}

	return records
}

func getAccountDenomsOrdersCount(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	acc sdk.AccAddress,
) map[string]uint64 {
	denomToCount := make(map[string]uint64)
	accountsDenomsOrdersCount, _, err := testApp.DEXKeeper.GetPaginatedAccountsDenomsOrdersCounts(
		sdkCtx,
		&query.PageRequest{
			Limit: query.PaginationMaxLimit,
		},
	)
	require.NoError(t, err)

	accNumber := testApp.AccountKeeper.GetAccount(sdkCtx, acc).GetAccountNumber()

	for _, accountDenomsOrdersCount := range accountsDenomsOrdersCount {
		if accountDenomsOrdersCount.AccountNumber != accNumber {
			continue
		}
		denomToCount[accountDenomsOrdersCount.Denom] = accountDenomsOrdersCount.OrdersCount
	}

	return denomToCount
}

func assertTickCalculations(t *testing.T, base, quote float64) {
	tickExponent := -5

	finalTickExp := math.Floor(math.Log10(quote/base)) + float64(tickExponent)
	finalTick := math.Pow(10, finalTickExp)

	baseDenomRefAmount := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%.15f", base))
	quoteRefAmount := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%.15f", quote))
	keeperPriceTick := keeper.ComputePriceTick(baseDenomRefAmount, quoteRefAmount, int32(tickExponent))
	require.Equal(t, fmt.Sprintf("%.15f", finalTick), keeperPriceTick.FloatString(15))
}
