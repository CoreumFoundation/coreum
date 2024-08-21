package keeper_test

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/docker/distribution/uuid"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func FuzzSaveSellOrderAndReadWithSorting(f *testing.F) {
	f.Add(uint64(0), int8(0))
	f.Add(uint64(123), types.MaxExp)
	f.Add(uint64(4123123123), types.MinExt)
	f.Add(uint64(9999999999999999999), types.MaxExp)
	f.Add(uint64(1), types.MinExt)

	testApp := simapp.New()
	lock := &sync.Mutex{}

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		lock.Lock()
		defer lock.Unlock()
		placeRandomOrderAndAssertOrdering(t, testApp, num, exp, types.SIDE_SELL)
	})
}

func FuzzSaveBuyOrderAndReadWithSorting(f *testing.F) {
	f.Add(uint64(123), int8(-3))
	f.Add(uint64(0), int8(0))
	f.Add(uint64(1), int8(-10))
	f.Add(uint64(9999999999999999999), int8(10))

	testApp := simapp.New()
	lock := &sync.Mutex{}

	f.Fuzz(func(t *testing.T, num uint64, exp int8) {
		// to prevent fast fail, because of out of sdk,Int range in the bank keeper at the time of the funding
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

	// prepare valid price
	var expPart string
	if exp != 0 {
		expPart = types.ExponentSymbol + strconv.Itoa(int(exp))
	}
	numPart := strconv.FormatUint(num, 10)
	if strings.HasSuffix(numPart, "0") || len(numPart) > types.MaxNumLen {
		t.Skip()
	}
	if exp > types.MaxExp || exp < types.MinExt {
		t.Skip()
	}
	priceStr := strconv.FormatUint(num, 10) + expPart
	price := types.MustNewPriceFromString(priceStr)

	sdkCtx, _, _ := testApp.BeginNextBlock(time.Now())
	acc, _ := testApp.GenAccount(sdkCtx)

	order := types.Order{
		Creator:    acc.String(),
		Type:       types.ORDER_TYPE_LIMIT,
		ID:         uuid.Generate().String(),
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
		Price:      &price,
		Quantity:   sdkmath.NewInt(1),
		Side:       side,
	}
	t.Logf("Order to place: %s", order.String())
	lockedBalance, err := order.ComputeLimitOrderLockedBalance()
	if err != nil {
		// the generated balance might overflow the sdkmath.Int type
		t.Skip()
	}
	testApp.MintAndSendCoin(t, sdkCtx, acc, sdk.NewCoins(lockedBalance))

	require.NoError(t, testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	orderBookID, err := testApp.DEXKeeper.GetOrderBookIDByDenoms(sdkCtx, baseDenom, quoteDenom)
	require.NoError(t, err)

	assertOrdersOrdering(t, testApp, sdkCtx, orderBookID, side)

	_, err = testApp.EndBlocker(sdkCtx)
	require.NoError(t, err)
}
