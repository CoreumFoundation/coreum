package keeper_test

import (
	"sync"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func FuzzSaveSellOrderAndReadWithSorting(f *testing.F) {
	f.Add(uint64(1), int8(0))
	f.Add(uint64(123), int8(-5))
	f.Add(uint64(123456789), int8(5))
	f.Add(uint64(9999999999999999999), int8(10))

	testApp := simapp.New()
	lock := sync.Mutex{}

	sdkCtx, _, _ := testApp.BeginNextBlock()

	// don't limit the price tick
	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(f, err)
	params.PriceTickExponent = int32(types.MinExp)
	require.NoError(f, testApp.DEXKeeper.SetParams(sdkCtx, params))

	_, err = testApp.EndBlocker(sdkCtx)
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		// to prevent fast fail, because of out of sdkmath.Int range in the bank keeper at the time of the funding
		// we limit the exponent.
		if exp < -10 || exp > 10 {
			t.Skip()
		}
		lock.Lock()
		defer lock.Unlock()
		placeRandomOrderAndAssertOrdering(t, testApp, num, exp, types.SIDE_SELL)
	})
}

func FuzzSaveBuyOrderAndReadWithSorting(f *testing.F) {
	f.Add(uint64(1), int8(0))
	f.Add(uint64(123), int8(-5))
	f.Add(uint64(123456789), int8(5))
	f.Add(uint64(9999999999999999999), int8(10))

	testApp := simapp.New()
	lock := sync.Mutex{}

	// don't limit the price tick
	sdkCtx, _, _ := testApp.BeginNextBlock()

	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(f, err)
	params.PriceTickExponent = int32(types.MinExp)
	require.NoError(f, testApp.DEXKeeper.SetParams(sdkCtx, params))

	_, err = testApp.EndBlocker(sdkCtx)
	require.NoError(f, err)

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		// to prevent fast fail, because of out of sdkmath.Int range in the bank keeper at the time of the funding
		// we limit the exponent.
		if exp < -10 || exp > 10 {
			t.Skip()
		}
		lock.Lock()
		defer lock.Unlock()
		placeRandomOrderAndAssertOrdering(t, testApp, num, exp, types.SIDE_BUY)
	})
}

func placeRandomOrderAndAssertOrdering(
	t *testing.T,
	testApp *simapp.App,
	num uint64,
	exp int8,
	side types.Side,
) {
	baseDenom := denom1
	quoteDenom := denom2

	price, ok := buildNumExpPrice(num, exp)
	if !ok {
		t.Skip()
	}

	sdkCtx, _, _ := testApp.BeginNextBlock()
	acc, _ := testApp.GenAccount(sdkCtx)

	order := types.Order{
		Creator:     acc.String(),
		Type:        types.ORDER_TYPE_LIMIT,
		ID:          uuid.Generate().String(),
		BaseDenom:   baseDenom,
		QuoteDenom:  quoteDenom,
		Price:       &price,
		Quantity:    sdkmath.NewInt(1),
		Side:        side,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}
	t.Logf("Order to place: %s", order.String())
	lockedBalance, err := order.ComputeLimitOrderLockedBalance()
	require.NoError(t, err)
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))
	fundOrderReserve(t, testApp, sdkCtx, acc)
	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	orderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, baseDenom, quoteDenom)
	require.NoError(t, err)

	assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, side)

	_, err = testApp.EndBlocker(sdkCtx)
	require.NoError(t, err)
}
