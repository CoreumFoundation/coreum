package keeper_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

const (
	denom1 = "denom1"
	denom2 = "denom2"
	denom3 = "denom3"
)

type OrderPlacementEvents struct {
	OrderPlaced   types.EventOrderPlaced
	OrdersReduced []types.EventOrderReduced
	OrderCreated  *types.EventOrderCreated
	OrdersClosed  []types.EventOrderClosed
}

func (o OrderPlacementEvents) getOrderReduced(acc, id string) (types.EventOrderReduced, bool) {
	for _, evt := range o.OrdersReduced {
		if evt.Creator == acc && evt.ID == id {
			return evt, true
		}
	}

	return types.EventOrderReduced{}, false
}

func readOrderEvents(
	t *testing.T,
	sdkCtx sdk.Context,
) OrderPlacementEvents {
	events := OrderPlacementEvents{
		OrderCreated:  nil,
		OrdersReduced: make([]types.EventOrderReduced, 0),
		OrdersClosed:  make([]types.EventOrderClosed, 0),
	}

	for _, evt := range sdkCtx.EventManager().Events().ToABCIEvents() {
		if !strings.HasPrefix(evt.Type, "coreum.dex.v1") {
			continue
		}
		msg, err := sdk.ParseTypedEvent(evt)
		require.NoError(t, err)
		switch typedEvt := msg.(type) {
		case *types.EventOrderPlaced:
			require.Empty(t, events.OrderPlaced.Creator, "Only one types.OrderPlaced is expected.")
			events.OrderPlaced = *typedEvt
		case *types.EventOrderReduced:
			events.OrdersReduced = append(events.OrdersReduced, *typedEvt)
		case *types.EventOrderCreated:
			require.Nil(t, events.OrderCreated, "Only one types.EventOrderCreated is expected.")
			events.OrderCreated = typedEvt
		case *types.EventOrderClosed:
			events.OrdersClosed = append(events.OrdersClosed, *typedEvt)
		}
	}

	return events
}

func TestKeeper_UpdateParams(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)
	dexKeeper := testApp.DEXKeeper

	gotParams, err := dexKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), gotParams)

	newPrams := gotParams
	newPrams.DefaultUnifiedRefAmount = sdkmath.LegacyMustNewDecFromStr("33.33")
	newPrams.PriceTickExponent = gotParams.PriceTickExponent - 2
	newPrams.QuantityStepExponent = gotParams.QuantityStepExponent - 1
	newPrams.OrderReserve = sdk.NewInt64Coin(sdk.DefaultBondDenom, 313)

	randomAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	// try to update for random address
	require.ErrorIs(t, dexKeeper.UpdateParams(sdkCtx, randomAddr.String(), newPrams), govtypes.ErrInvalidSigner)

	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	require.NoError(t, dexKeeper.UpdateParams(sdkCtx, govAddr, newPrams))
	gotParams, err = dexKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
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
			expectedSelfOrderBookID:     uint32(1),
			expectedOppositeOrderBookID: uint32(2),
		},
		// save one more time to check that returns the same
		{
			baseDenom:                   denom1,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(1),
			expectedOppositeOrderBookID: uint32(2),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom1,
			expectedSelfOrderBookID:     uint32(2),
			expectedOppositeOrderBookID: uint32(1),
		},
		// save with desc denoms ordering
		{
			baseDenom:                   denom3,
			quoteDenom:                  denom2,
			expectedSelfOrderBookID:     uint32(4),
			expectedOppositeOrderBookID: uint32(3),
		},
		// inverse denom
		{
			baseDenom:                   denom2,
			quoteDenom:                  denom3,
			expectedSelfOrderBookID:     uint32(3),
			expectedOppositeOrderBookID: uint32(4),
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
		fundOrderReserve(t, testApp, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator))

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
	fundOrderReserve(t, testApp, sdkCtx, acc)

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
	sellOrder.Sequence = 1
	sellOrder.RemainingBaseQuantity = sdkmath.NewInt(10)
	sellOrder.RemainingSpendableBalance = sdkmath.NewInt(10)
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	orderReserve := params.OrderReserve
	sellOrder.Reserve = orderReserve
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
	fundOrderReserve(t, testApp, sdkCtx, acc)

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))

	gotOrder, err = dexKeeper.GetOrderByAddressAndID(
		sdkCtx, sdk.MustAccAddressFromBech32(buyOrder.Creator), buyOrder.ID,
	)
	require.NoError(t, err)

	// set expected values
	buyOrder.Sequence = 2
	buyOrder.RemainingBaseQuantity = sdkmath.NewInt(100)
	buyOrder.RemainingSpendableBalance = sdkmath.NewInt(120)
	buyOrder.Reserve = orderReserve
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
	issuer, _ := testApp.GenAccount(sdkCtx)

	issuanceSettings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "DEFEXT",
		Subunit:       "defext",
		Precision:     6,
		InitialAmount: sdkmath.NewIntWithDecimal(1, 10),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_whitelisting,
		},
	}
	ft1Whitelisting, err := testApp.AssetFTKeeper.Issue(sdkCtx, issuanceSettings)
	require.NoError(t, err)

	sellOrder := types.Order{
		Creator:     acc.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          "id1",
		BaseDenom:   denom1,
		QuoteDenom:  ft1Whitelisting,
		Price:       lo.ToPtr(types.MustNewPriceFromString("12e-1")),
		Quantity:    sdkmath.NewInt(1_000),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	sellLockedBalance, err := sellOrder.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(sellLockedBalance))
	fundOrderReserve(t, testApp, sdkCtx, acc)
	sellLWhitelistedBalance, err := types.ComputeLimitOrderExpectedToReceiveBalance(
		sellOrder.Side, sellOrder.BaseDenom, sellOrder.QuoteDenom, sellOrder.Quantity, *sellOrder.Price,
	)
	require.NoError(t, err)
	require.NoError(t, testApp.AssetFTKeeper.SetWhitelistedBalance(sdkCtx, issuer, acc, sellLWhitelistedBalance))

	sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))
	events := readOrderEvents(t, sdkCtx)
	require.NotNil(t, events.OrderCreated)

	expectedSellOrderSequence := uint64(1)
	require.Equal(t, types.EventOrderPlaced{
		Creator:  sellOrder.Creator,
		ID:       sellOrder.ID,
		Sequence: expectedSellOrderSequence, // first order sequence
	}, events.OrderPlaced)

	require.Equal(t, types.EventOrderCreated{
		Creator:                   sellOrder.Creator,
		ID:                        sellOrder.ID,
		Sequence:                  expectedSellOrderSequence,
		RemainingBaseQuantity:     sellOrder.Quantity,
		RemainingSpendableBalance: sellLockedBalance.Amount,
	}, *events.OrderCreated)
	require.Empty(t, events.OrdersClosed)
	require.Empty(t, events.OrdersReduced)

	dexLockedBalance := assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.Equal(t, sellLockedBalance.String(), dexLockedBalance.String())
	dexExpectedToReceiveBalance := assetFTKeeper.GetDEXExpectedToReceivedBalance(
		sdkCtx, acc, sellLWhitelistedBalance.Denom,
	)
	require.Equal(t, sellLWhitelistedBalance.String(), dexExpectedToReceiveBalance.String())

	sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, sellOrder.ID))
	events = readOrderEvents(t, sdkCtx)
	require.Nil(t, events.OrderCreated)
	require.EqualValues(t, []types.EventOrderClosed{
		{
			Creator:                   sellOrder.Creator,
			ID:                        sellOrder.ID,
			Sequence:                  expectedSellOrderSequence,
			RemainingBaseQuantity:     sellOrder.Quantity,
			RemainingSpendableBalance: sellLockedBalance.Amount,
		},
	}, events.OrdersClosed)
	require.Empty(t, events.OrdersReduced)

	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, sellLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())
	dexExpectedToReceiveBalance = assetFTKeeper.GetDEXExpectedToReceivedBalance(
		sdkCtx, acc, sellLWhitelistedBalance.Denom,
	)
	require.True(t, dexExpectedToReceiveBalance.IsZero())

	buyOrder := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         "id2",
		BaseDenom:  denom1,
		QuoteDenom: ft1Whitelisting,
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
	// whitelist for both orders
	whitelistedBalance := sellLWhitelistedBalance.Add(buyLockedBalance)
	require.NoError(t, testApp.AssetFTKeeper.SetWhitelistedBalance(sdkCtx, issuer, acc, whitelistedBalance))
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(buyLockedBalance))
	fundOrderReserve(t, testApp, sdkCtx, acc)

	// try to place order with the invalid GoodTilBlockHeight
	buyOrderWithGoodTilHeight := buyOrder
	buyOrderWithGoodTilHeight.GoodTil = &types.GoodTil{
		GoodTilBlockHeight: uint64(sdkCtx.BlockHeight() - 1),
	}
	require.ErrorContains(
		t,
		dexKeeper.PlaceOrder(simapp.CopyContextWithMultiStore(sdkCtx), buyOrderWithGoodTilHeight),
		"good til block height 99 must be greater than current block height 100: invalid input",
	)

	// try to place order with the invalid GoodTilBlockTime
	buyOrderWithGoodTilTime := buyOrder
	buyOrderWithGoodTilTime.GoodTil = &types.GoodTil{
		GoodTilBlockTime: lo.ToPtr(sdkCtx.BlockTime()),
	}
	require.ErrorContains(
		t,
		dexKeeper.PlaceOrder(simapp.CopyContextWithMultiStore(sdkCtx), buyOrderWithGoodTilTime),
		"good til block time 2023-03-02 01:11:12.000000013 +0000 UTC must be greater than current block time",
	)

	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.Equal(t, buyLockedBalance.String(), dexLockedBalance.String())
	// check unlocking
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())
	dexExpectedToReceiveBalance = assetFTKeeper.GetDEXExpectedToReceivedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexExpectedToReceiveBalance.IsZero())

	// now place both orders to let them match partially
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, sellOrder))

	sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
	require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, buyOrder))
	events = readOrderEvents(t, sdkCtx)

	// update sequence
	expectedSellOrderSequence = uint64(3)
	expectedBuyOrderSequence := uint64(4)
	require.Equal(t, types.EventOrderCreated{
		Creator:                   buyOrder.Creator,
		ID:                        buyOrder.ID,
		Sequence:                  expectedBuyOrderSequence,
		RemainingBaseQuantity:     sdkmath.NewInt(4000), // filled partially
		RemainingSpendableBalance: sdkmath.NewInt(5200),
	}, *events.OrderCreated)

	require.EqualValues(t, []types.EventOrderReduced{
		{
			Creator:      sellOrder.Creator,
			ID:           sellOrder.ID,
			Sequence:     expectedSellOrderSequence,
			SentCoin:     sdk.NewCoin(sellOrder.BaseDenom, sdkmath.NewIntFromUint64(1000)),
			ReceivedCoin: sdk.NewCoin(sellOrder.QuoteDenom, sdkmath.NewIntFromUint64(1200)),
		},
		{
			Creator:      buyOrder.Creator,
			ID:           buyOrder.ID,
			Sequence:     expectedBuyOrderSequence,
			SentCoin:     sdk.NewCoin(buyOrder.QuoteDenom, sdkmath.NewIntFromUint64(1200)),
			ReceivedCoin: sdk.NewCoin(buyOrder.BaseDenom, sdkmath.NewIntFromUint64(1000)),
		},
	}, events.OrdersReduced)

	require.EqualValues(t, []types.EventOrderClosed{
		{
			Creator:                   sellOrder.Creator,
			ID:                        sellOrder.ID,
			Sequence:                  expectedSellOrderSequence,
			RemainingBaseQuantity:     sdkmath.ZeroInt(),
			RemainingSpendableBalance: sdkmath.ZeroInt(),
		},
	}, events.OrdersClosed)

	_, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, sellOrder.ID)
	require.ErrorIs(t, err, types.ErrRecordNotFound)
	buyOrder, err = dexKeeper.GetOrderByAddressAndID(sdkCtx, acc, buyOrder.ID)
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(5200).String(), buyOrder.RemainingSpendableBalance.String())
	require.NoError(t, dexKeeper.CancelOrder(sdkCtx, acc, buyOrder.ID))
	// check unlocking
	dexLockedBalance = assetFTKeeper.GetDEXLockedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexLockedBalance.IsZero())
	dexExpectedToReceiveBalance = assetFTKeeper.GetDEXExpectedToReceivedBalance(sdkCtx, acc, buyLockedBalance.Denom)
	require.True(t, dexExpectedToReceiveBalance.IsZero())
}

func TestKeeper_PlaceOrder_PriceTickAndQuantityStep(t *testing.T) {
	tests := []struct {
		name              string
		price             *types.Price
		quantity          *sdkmath.Int
		baseURA           *sdkmath.LegacyDec
		quoteURA          *sdkmath.LegacyDec
		wantQuantityError bool
		wantPriceError    bool
	}{
		{
			name:           "valid_default_URAs",
			price:          lo.ToPtr(types.MustNewPriceFromString("1e-6")),
			wantPriceError: false,
		},
		{
			name:              "invalid_quantity_default_URAs",
			quantity:          lo.ToPtr(sdkmath.NewInt(10)),
			wantQuantityError: true,
		},
		{
			name:           "invalid_price_default_URAs",
			price:          lo.ToPtr(types.MustNewPriceFromString("1e-7")),
			wantPriceError: true,
		},
		{
			name:           "invalid_price2_default_URAs",
			price:          lo.ToPtr(types.MustNewPriceFromString("100000001e-7")),
			wantPriceError: true,
		},
		{
			name:    "valid_custom_base_URA",
			price:   lo.ToPtr(types.MustNewPriceFromString("33e-6")),
			baseURA: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10000000")),
		},
		{
			name:     "valid_custom_quote_URA",
			price:    lo.ToPtr(types.MustNewPriceFromString("1e-6")),
			quoteURA: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000")),
		},
		{
			name:     "valid_custom_both_URA",
			price:    lo.ToPtr(types.MustNewPriceFromString("14e-3")),
			baseURA:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("1")),
			quoteURA: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100")),
		},
		{
			name:     "valid_custom_both_URA_tick_greater_than_one",
			price:    lo.ToPtr(types.MustNewPriceFromString("14e1")),
			baseURA:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.01")),
			quoteURA: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10303.3")),
		},
		{
			name:           "invalid_price_custom_both_URA_tick_greater_than_one",
			price:          lo.ToPtr(types.MustNewPriceFromString("14")),
			baseURA:        lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.00001")),
			quoteURA:       lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10303.3")),
			wantPriceError: true,
		},
		{
			name:              "invalid_quantity_custom_base_URA_quantity_step_greater_than_one",
			quantity:          lo.ToPtr(sdkmath.NewInt(123)),
			baseURA:           lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("10000000")),
			wantQuantityError: true,
		},
		{
			name:     "valid_custom_both_URA_base_URA_less_than_one",
			price:    lo.ToPtr(types.MustNewPriceFromString("3e33")),
			baseURA:  lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
			quoteURA: lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000000000000000000")),
		},
		{
			name:           "invalid_price_custom_both_URA_base_URA_less_than_one",
			price:          lo.ToPtr(types.MustNewPriceFromString("3e32")),
			baseURA:        lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("0.000000000000000001")),
			quoteURA:       lo.ToPtr(sdkmath.LegacyMustNewDecFromStr("100000000000000000000000")),
			wantPriceError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testApp := simapp.New()
			sdkCtx := testApp.BaseApp.NewContext(false)

			if tt.baseURA != nil {
				require.NoError(t, testApp.AssetFTKeeper.SetDEXSettings(sdkCtx, denom1, assetfttypes.DEXSettings{
					UnifiedRefAmount: tt.baseURA,
				}))
			}

			if tt.quoteURA != nil {
				require.NoError(t, testApp.AssetFTKeeper.SetDEXSettings(sdkCtx, denom2, assetfttypes.DEXSettings{
					UnifiedRefAmount: tt.quoteURA,
				}))
			}

			price := types.MustNewPriceFromString("1e3")
			quantity := sdkmath.NewInt(1_000_000)

			if tt.price != nil {
				price = *tt.price
			}
			if tt.quantity != nil {
				quantity = *tt.quantity
			}

			acc, _ := testApp.GenAccount(sdkCtx)
			order := types.Order{
				Creator:     acc.String(),
				Type:        types.ORDER_TYPE_LIMIT,
				ID:          uuid.Generate().String(),
				BaseDenom:   denom1,
				QuoteDenom:  denom2,
				Price:       &price,
				Quantity:    quantity,
				Side:        types.SIDE_SELL,
				TimeInForce: types.TIME_IN_FORCE_GTC,
			}
			lockedBalance, err := order.ComputeLimitOrderLockedBalance()
			require.NoError(t, err)
			testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))
			fundOrderReserve(t, testApp, sdkCtx, acc)
			err = testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
			if tt.wantPriceError {
				require.ErrorContains(t, err, "has to be multiple of price tick")
			} else if tt.wantQuantityError {
				require.ErrorContains(t, err, "has to be multiple of quantity step")
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
		fundOrderReserve(t, testApp, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator))
		require.NoError(t, dexKeeper.PlaceOrder(sdkCtx, order))
		// fill order with the remaining quantity for assertions
		order.RemainingBaseQuantity = order.Quantity
		order.RemainingSpendableBalance = lockedBalance.Amount
		orders[i] = order
	}
	orders = fillReserveAndOrderSequence(t, sdkCtx, testApp, orders)

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
		fundOrderReserve(t, testApp, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator))
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

func TestKeeper_PlaceAndCancelOrderWithMaxAllowedAccountDenomOrdersCount(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)

	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
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
	fundOrderReserve(t, testApp, sdkCtx, acc1)
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
	fundOrderReserve(t, testApp, sdkCtx, acc1)
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
	fundOrderReserve(t, testApp, sdkCtx, acc1)
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
	fundOrderReserve(t, testApp, sdkCtx, acc1)
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
	fundOrderReserve(t, testApp, sdkCtx, acc2)
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

func TestKeeper_PlaceAndCancelOrdersByDenom(t *testing.T) {
	testApp := simapp.New()
	sdkCtx := testApp.BaseApp.NewContext(false)

	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

	acc1, _ := testApp.GenAccount(sdkCtx)
	issuer, _ := testApp.GenAccount(sdkCtx)

	settings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Symbol:        "INVD",
		Subunit:       "invd",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(100),
	}
	denomWithProhibitedCancellation, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
	require.NoError(t, err)
	// check access for not FT denom
	require.ErrorIs(t, testApp.DEXKeeper.CancelOrdersByDenom(
		sdkCtx, issuer, acc1, "nativedenom"), assetfttypes.ErrInvalidDenom,
	)
	// check access with disabled feature
	require.ErrorIs(t, testApp.DEXKeeper.CancelOrdersByDenom(
		sdkCtx, issuer, acc1, denomWithProhibitedCancellation), cosmoserrors.ErrUnauthorized,
	)

	denoms := make([]string, 0)
	for i := range 3 {
		settings := assetfttypes.IssueSettings{
			Issuer:        issuer,
			Symbol:        fmt.Sprintf("SMB%d", i),
			Subunit:       fmt.Sprintf("sut%d", i),
			Precision:     1,
			InitialAmount: sdkmath.NewInt(100),
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_dex_order_cancellation,
			},
		}
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		denoms = append(denoms, denom)
	}

	// place 5 limit orders to denom0/denom1
	for range 5 {
		order := types.Order{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denoms[0],
			QuoteDenom:  denoms[1],
			Price:       lo.ToPtr(types.MustNewPriceFromString("1e-1")),
			Quantity:    sdkmath.NewInt(1_000),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		}
		sellLockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
		fundOrderReserve(t, testApp, sdkCtx, acc1)
		require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	}
	// place 5 limit orders to denom1/denom0
	for range 5 {
		order := types.Order{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denoms[1],
			QuoteDenom:  denoms[0],
			Price:       lo.ToPtr(types.MustNewPriceFromString("11e1")),
			Quantity:    sdkmath.NewInt(1_000),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		}
		sellLockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
		fundOrderReserve(t, testApp, sdkCtx, acc1)
		require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	}
	// place 5 limit orders to denom1/denom2
	for range 5 {
		order := types.Order{
			Creator:     acc1.String(),
			Type:        types.ORDER_TYPE_LIMIT,
			ID:          uuid.Generate().String(),
			BaseDenom:   denoms[0],
			QuoteDenom:  denoms[2],
			Price:       lo.ToPtr(types.MustNewPriceFromString("12e1")),
			Quantity:    sdkmath.NewInt(1_000),
			Side:        types.SIDE_SELL,
			TimeInForce: types.TIME_IN_FORCE_GTC,
		}
		sellLockedBalance, err := order.ComputeLimitOrderLockedBalance()
		require.NoError(t, err)
		testApp.MintAndSendCoin(t, sdkCtx, acc1, sdk.NewCoins(sellLockedBalance))
		fundOrderReserve(t, testApp, sdkCtx, acc1)
		require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))
	}

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denoms[0]: 15,
			denoms[1]: 10,
			denoms[2]: 5,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	orders, _, err := testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Limit: query.PaginationMaxLimit,
	})
	require.NoError(t, err)
	require.Len(t, orders, 15)

	// try to cancel from not admin
	require.ErrorIs(t, testApp.DEXKeeper.CancelOrdersByDenom(sdkCtx, acc1, acc1, denoms[1]), cosmoserrors.ErrUnauthorized)

	// cancel orders fro admin
	require.NoError(t, testApp.DEXKeeper.CancelOrdersByDenom(sdkCtx, issuer, acc1, denoms[1]))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denoms[0]: 5,
			denoms[1]: 0,
			denoms[2]: 5,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))

	orders, _, err = testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Limit: query.PaginationMaxLimit,
	})
	require.NoError(t, err)
	require.Len(t, orders, 5)
	// check that there are not orders with the denom2
	for _, order := range orders {
		for _, denom := range order.Denoms() {
			require.NotEqual(t, denom2, denom)
		}
	}

	// cancel remaining
	require.NoError(t, testApp.DEXKeeper.CancelOrdersByDenom(sdkCtx, issuer, acc1, denoms[0]))

	require.True(t, reflect.DeepEqual(
		map[string]uint64{
			denoms[0]: 0,
			denoms[1]: 0,
			denoms[2]: 0,
		},
		getAccountDenomsOrdersCount(t, testApp, sdkCtx, acc1),
	))
	orders, _, err = testApp.DEXKeeper.GetOrders(sdkCtx, acc1, &query.PageRequest{
		Limit: query.PaginationMaxLimit,
	})
	require.NoError(t, err)
	require.Empty(t, orders)

	// cancel empty list, should not fail
	require.NoError(t, testApp.DEXKeeper.CancelOrdersByDenom(sdkCtx, issuer, acc1, denoms[0]))
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
	accountsDenomsOrdersCount, _, err := testApp.DEXKeeper.GetAccountsDenomsOrdersCounts(
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

func fundOrderReserve(
	t *testing.T,
	testApp *simapp.App,
	sdkCtx sdk.Context,
	acc sdk.AccAddress,
) {
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	orderReserve := params.OrderReserve
	if !orderReserve.IsPositive() {
		return
	}
	require.NoError(t, testApp.FundAccount(sdkCtx, acc, sdk.NewCoins(orderReserve)))
}
