package keeper_test

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

type FuzzAppConfig struct {
	AccountsCount                 int
	AssetFTDefaultDenomsCount     int
	AssetFTWhitelistingCount      int
	AssetFTFreezingDenomsCount    int
	AssetFTAllFeaturesDenomsCount int
	NativeDenomCount              int

	OrdersCount                int
	CancelOrdersPercent        int
	CancelOrdersByDenomPercent int

	MarketOrdersPercent       int
	TimeInForceIOCPercent     int
	TimeInForceFOKPercent     int
	GoodTilBlockHeightPercent int
	GoodTilBlockTimePercent   int
	FundOrderReservePercent   int

	InitialBlockHeight uint64
	InitialBlockTime   time.Time
	BlockTime          time.Duration
}

type FuzzApp struct {
	cfg FuzzAppConfig

	testApp  *simapp.App
	issuer   sdk.AccAddress
	accounts []sdk.AccAddress
	denoms   []string
	ftDenoms []string
	sides    []types.Side
}

func NewFuzzApp(
	t *testing.T,
	cfg FuzzAppConfig,
) FuzzApp {
	testApp := simapp.New()

	sdkCtx, _, _ := testApp.BeginNextBlock()

	params := testApp.DEXKeeper.GetParams(sdkCtx)
	params.PriceTickExponent = int32(types.MinExt)

	require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

	accounts := lo.RepeatBy(cfg.AccountsCount, func(_ int) sdk.AccAddress {
		acc, _ := testApp.GenAccount(sdkCtx)
		return acc
	})

	issuer, _ := testApp.GenAccount(sdkCtx)

	mintedAmount := sdkmath.NewIntWithDecimal(1, 77)
	denoms := lo.RepeatBy(cfg.NativeDenomCount, func(i int) string {
		denom := fmt.Sprintf("native-denom-%d", i)
		testApp.MintAndSendCoin(t, sdkCtx, issuer, sdk.NewCoins(sdk.NewCoin(denom, mintedAmount)))
		return denom
	})

	ftDenoms := make([]string, 0)

	defaultFTSettings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Precision:     1,
		InitialAmount: mintedAmount,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_order_cancellation,
		},
	}
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTDefaultDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("DEF%d", i)
		settings.Subunit = fmt.Sprintf("def%d", i)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTWhitelistingCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("WLS%d", i)
		settings.Subunit = fmt.Sprintf("wls%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_whitelisting)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTFreezingDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("FRZ%d", i)
		settings.Subunit = fmt.Sprintf("frz%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_freezing)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTAllFeaturesDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("ALL%d", i)
		settings.Subunit = fmt.Sprintf("all%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_whitelisting)
		settings.Features = append(settings.Features, assetfttypes.Feature_freezing)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	denoms = append(denoms, ftDenoms...)

	sides := []types.Side{
		types.SIDE_BUY,
		types.SIDE_SELL,
	}

	_, err := testApp.EndBlocker(sdkCtx)
	require.NoError(t, err)

	return FuzzApp{
		cfg:      cfg,
		testApp:  testApp,
		issuer:   issuer,
		accounts: accounts,
		denoms:   denoms,
		ftDenoms: ftDenoms,
		sides:    sides,
	}
}

func (fa *FuzzApp) PlaceOrdersAndAssertFinalState(
	t *testing.T,
	rootRnd *rand.Rand,
) {
	sdkCtx := fa.testApp.NewContextLegacy(false, tmproto.Header{
		Height: int64(fa.cfg.InitialBlockHeight),
		Time:   fa.cfg.InitialBlockTime,
	})

	for i := 0; i < fa.cfg.OrdersCount; i++ {
		_, err := fa.testApp.BeginBlocker(sdkCtx)
		require.NoError(t, err)

		orderSeed := rootRnd.Int63()
		orderRnd := rand.New(rand.NewSource(orderSeed))

		order := fa.GenOrder(t, orderRnd)

		t.Logf("Placing order, i:%d, seed:%d, order: %s", i, orderSeed, order.String())
		fa.FundAccountAndApplyFTFeatures(t, sdkCtx, order, orderRnd)
		fa.PlaceOrder(t, sdkCtx, order)

		if randBoolWithPercent(orderRnd, fa.cfg.CancelOrdersPercent) {
			fa.CancelFirstOrder(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator))
		}

		if randBoolWithPercent(orderRnd, fa.cfg.CancelOrdersByDenomPercent) {
			fa.CancelOrdersByDenom(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.BaseDenom)
		}

		_, err = fa.testApp.EndBlocker(sdkCtx)
		require.NoError(t, err)

		sdkCtx = fa.testApp.NewContextLegacy(false, tmproto.Header{
			Height: sdkCtx.BlockHeight() + 1,
			Time:   sdkCtx.BlockTime().Add(fa.cfg.BlockTime),
		})
	}
	cancelAllOrdersAndAssertState(t, sdkCtx, fa.testApp)
}

func (fa *FuzzApp) GenOrder(
	t *testing.T,
	rnd *rand.Rand,
) types.Order {
	t.Helper()

	creator := genAnyItemByIndex(fa.accounts, uint8(rnd.Uint32()))
	orderType := types.ORDER_TYPE_LIMIT
	if randBoolWithPercent(rnd, fa.cfg.MarketOrdersPercent) {
		orderType = types.ORDER_TYPE_MARKET
	}

	// take next if same
	var baseDenom, quoteDenom string
	baseDenomInd := uint8(rnd.Uint32())
	for {
		baseDenom = genAnyItemByIndex(fa.denoms, uint8(rnd.Uint32()))
		quoteDenom = genAnyItemByIndex(fa.denoms, baseDenomInd)
		if baseDenom != quoteDenom {
			break
		}
		baseDenomInd++
	}

	side := genAnyItemByIndex(fa.sides, uint8(rnd.Uint32()))

	priceNum := rnd.Uint32()

	// generate price exponent in order not to overflow the sdkmath.Int when fund accounts
	priceExp := int8(randIntInRange(rnd, -10, 10))

	var (
		price       *types.Price
		goodTil     *types.GoodTil
		timeInForce = types.TIME_IN_FORCE_UNSPECIFIED
	)

	if orderType == types.ORDER_TYPE_LIMIT { //nolint:nestif  // the ifs are simple to check the percents mostly
		v, ok := buildNumExpPrice(uint64(priceNum), priceExp)
		// since we use Uint32 as num it never exceed the max num length
		require.True(t, ok)
		price = &v

		if randBoolWithPercent(rnd, fa.cfg.GoodTilBlockHeightPercent) {
			goodTil = &types.GoodTil{
				GoodTilBlockHeight: fa.cfg.InitialBlockHeight + uint64(randIntInRange(rnd, 0, 2000)),
			}
		}

		if randBoolWithPercent(rnd, fa.cfg.GoodTilBlockTimePercent) {
			if goodTil == nil {
				goodTil = &types.GoodTil{}
			}
			goodTil.GoodTilBlockTime = lo.ToPtr(
				fa.cfg.InitialBlockTime.Add(time.Duration(randIntInRange(rnd, 0, 2000)) * fa.cfg.BlockTime),
			)
		}

		timeInForce = types.TIME_IN_FORCE_GTC

		if randBoolWithPercent(rnd, fa.cfg.TimeInForceIOCPercent) {
			timeInForce = types.TIME_IN_FORCE_IOC
		}
		if randBoolWithPercent(rnd, fa.cfg.TimeInForceFOKPercent) {
			timeInForce = types.TIME_IN_FORCE_FOK
		}
	}

	// the quantity can't be zero
	quantity := rnd.Uint64()
	if quantity == 0 {
		quantity = 1
	}

	return types.Order{
		Creator:     creator.String(),
		Type:        orderType,
		ID:          genString(20, rnd),
		BaseDenom:   baseDenom,
		QuoteDenom:  quoteDenom,
		Price:       price,
		Quantity:    sdkmath.NewIntFromUint64(quantity),
		Side:        side,
		GoodTil:     goodTil,
		TimeInForce: timeInForce,
	}
}

func (fa *FuzzApp) FundAccountAndApplyFTFeatures(
	t *testing.T,
	sdkCtx sdk.Context,
	order types.Order,
	orderRnd *rand.Rand,
) {
	t.Helper()

	creator := sdk.MustAccAddressFromBech32(order.Creator)

	spendDef, err := fa.testApp.AssetFTKeeper.GetDefinition(sdkCtx, order.GetSpendDenom())
	if err != nil {
		require.True(t, sdkerrors.IsOf(err, assetfttypes.ErrInvalidDenom, assetfttypes.ErrTokenNotFound))
	}
	receiveDef, err := fa.testApp.AssetFTKeeper.GetDefinition(sdkCtx, order.GetReceiveDenom())
	if err != nil {
		require.True(t, sdkerrors.IsOf(err, assetfttypes.ErrInvalidDenom, assetfttypes.ErrTokenNotFound))
	}

	fundCoin := sdk.NewCoin(order.GetSpendDenom(), sdkmath.NewIntFromUint64(orderRnd.Uint64()))
	issuerBalance := fa.testApp.BankKeeper.GetBalance(sdkCtx, fa.issuer, fundCoin.Denom)
	if issuerBalance.IsLT(fundCoin) {
		t.Logf(
			"Failed to fund, insufficient issuer balance, balance:%s, recipient:%s,  coin:%s",
			issuerBalance.String(), creator.String(), fundCoin.String(),
		)
	} else {
		if spendDef.IsFeatureEnabled(assetfttypes.Feature_whitelisting) {
			dexWhitelistingReservedBalance := fa.testApp.AssetFTKeeper.GetDEXWhitelistingReservedBalance(
				sdkCtx, creator, fundCoin.Denom,
			)
			balance := fa.testApp.BankKeeper.GetBalance(sdkCtx, creator, fundCoin.Denom)
			whitelistBalance := fundCoin.Add(balance).Add(dexWhitelistingReservedBalance)
			t.Logf("Whitelisting initial account's balance: %s, %s", creator.String(), whitelistBalance.String())
			require.NoError(t, fa.testApp.AssetFTKeeper.SetWhitelistedBalance(sdkCtx, fa.issuer, creator, whitelistBalance))
		}
		t.Logf("Funding account: %s, %s", creator.String(), fundCoin.String())
		require.NoError(t, fa.testApp.BankKeeper.SendCoins(sdkCtx, fa.issuer, creator, sdk.NewCoins(fundCoin)))
	}

	if spendDef.IsFeatureEnabled(assetfttypes.Feature_freezing) {
		freezeCoin := sdk.NewCoin(order.GetSpendDenom(), sdkmath.NewIntFromUint64(orderRnd.Uint64()))
		t.Logf("Freezing account's coin: %s, %s", creator.String(), freezeCoin.String())
		require.NoError(t, fa.testApp.AssetFTKeeper.SetFrozen(sdkCtx, fa.issuer, creator, freezeCoin))
	}

	if receiveDef.IsFeatureEnabled(assetfttypes.Feature_whitelisting) {
		whitelistBalance := sdk.NewCoin(order.GetReceiveDenom(), sdkmath.NewIntFromUint64(orderRnd.Uint64()))
		t.Logf("Whitelisting account's coin: %s, %s", creator.String(), whitelistBalance.String())
		require.NoError(t, fa.testApp.AssetFTKeeper.SetWhitelistedBalance(sdkCtx, fa.issuer, creator, whitelistBalance))
	}

	if order.Type == types.ORDER_TYPE_LIMIT &&
		order.TimeInForce == types.TIME_IN_FORCE_GTC &&
		randBoolWithPercent(orderRnd, fa.cfg.FundOrderReservePercent) {
		reserve := fa.testApp.DEXKeeper.GetParams(sdkCtx).OrderReserve
		spendableBalance := fa.testApp.AssetFTKeeper.GetSpendableBalance(sdkCtx, creator, reserve.Denom)
		if spendableBalance.IsLT(reserve) {
			t.Logf("Funding order reserve, account: %s coin: %s", creator.String(), reserve.String())
			fa.testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(reserve))
		}
	}
}

func (fa *FuzzApp) PlaceOrder(t *testing.T, sdkCtx sdk.Context, order types.Order) {
	t.Helper()

	trialCtx := simapp.CopyContextWithMultiStore(sdkCtx) // copy to dry run and don't change state if error
	if err := fa.testApp.DEXKeeper.PlaceOrder(trialCtx, order); err != nil {
		t.Logf("Placement failed, err: %s", err.Error())
		creator := sdk.MustAccAddressFromBech32(order.Creator)
		switch {
		case sdkerrors.IsOf(err, assetfttypes.ErrDEXLockFailed):
			// check that the order can't be placed because of the lack of balance
			if order.Type != types.ORDER_TYPE_LIMIT {
				return
			}
			// check failed because of reserve
			reserve := fa.testApp.DEXKeeper.GetParams(sdkCtx).OrderReserve
			reserveDenomSpendableBalance := fa.testApp.AssetFTKeeper.GetSpendableBalance(
				sdkCtx, creator, reserve.Denom,
			)
			if reserveDenomSpendableBalance.Amount.LT(reserve.Amount) {
				t.Logf("Placement is failed due to insufficient reserve, reserve: %s, reserveDenomSpendableBalance: %s",
					reserve.String(), reserveDenomSpendableBalance.String())
				return
			}

			// check spendable balance
			spendableBalance := fa.testApp.AssetFTKeeper.GetSpendableBalance(
				sdkCtx, creator, order.GetSpendDenom(),
			)
			orderLockedBalance, err := order.ComputeLimitOrderLockedBalance()
			require.NoError(t, err)
			require.True(
				t,
				spendableBalance.IsLT(orderLockedBalance),
				fmt.Sprintf("availableBalance: %s, orderLockedBalance: %s", spendableBalance.String(), orderLockedBalance.String()),
			)
			t.Logf("Placement is failed due to lack of spendable balance, spendableBalance: %s, orderLockedBalance: %s",
				spendableBalance.String(), orderLockedBalance.String())
			return
		case sdkerrors.IsOf(err, assetfttypes.ErrWhitelistedLimitExceeded):
			if order.Side != types.SIDE_BUY {
				// check for the buy side only, since we don't know what was the execution quantity for the SELL
				return
			}

			var requiredWhitelistedBalance sdk.Coin

			switch order.Type {
			case types.ORDER_TYPE_LIMIT:
				requiredWhitelistedBalance, err = types.ComputeLimitOrderWhitelistingReservedBalance(
					order.Side, order.BaseDenom, order.QuoteDenom, order.Quantity, *order.Price,
				)
				require.NoError(t, err)
			case types.ORDER_TYPE_MARKET:
				requiredWhitelistedBalance = sdk.NewCoin(order.GetReceiveDenom(), order.Quantity)
			default:
				t.Fatalf("Unexpected order type: %s", order.Type.String())
			}

			requiredWhitelistedAmt := requiredWhitelistedBalance.Amount

			balance := fa.testApp.BankKeeper.GetBalance(sdkCtx, creator, requiredWhitelistedBalance.Denom)
			whitelistedBalance := fa.testApp.AssetFTKeeper.GetWhitelistedBalance(
				sdkCtx, creator, requiredWhitelistedBalance.Denom,
			)
			dexWhitelistingReservedBalance := fa.testApp.AssetFTKeeper.GetDEXWhitelistingReservedBalance(
				sdkCtx, creator, requiredWhitelistedBalance.Denom,
			)
			receivableAmt := whitelistedBalance.Amount.Sub(balance.Amount).Sub(dexWhitelistingReservedBalance.Amount)
			require.True(
				t,
				receivableAmt.LT(requiredWhitelistedAmt),
				fmt.Sprintf(
					"receivableAmt: %s, requiredWhitelistedAmt: %s",
					receivableAmt.String(), requiredWhitelistedAmt.String()),
			)
			return
		case strings.Contains(err.Error(), "good til"),
			strings.Contains(err.Error(), "it's prohibited to save more than"):
			return
		default:
			require.NoError(t, err)
		}
	}

	availableBalancesBefore := getAvailableBalances(sdkCtx, fa.testApp, sdk.MustAccAddressFromBech32(order.Creator))

	// use empty event manager for each order placement to check events properly
	sdkCtx = sdkCtx.WithEventManager(sdk.NewEventManager())
	require.NoError(t, fa.testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	// check if order is placed
	t.Log("Placement passed")
	assertOrderPlacementResult(t, sdkCtx, fa.testApp, availableBalancesBefore, order)
}

func (fa *FuzzApp) CancelFirstOrder(t *testing.T, sdkCtx sdk.Context, creator sdk.AccAddress) {
	t.Helper()

	orders, _, err := fa.testApp.DEXKeeper.GetOrders(
		sdkCtx,
		creator,
		&query.PageRequest{Limit: 1},
	)
	require.NoError(t, err)
	if len(orders) == 0 {
		return
	}

	t.Logf("Cancelling order.")
	require.NoError(t, fa.testApp.DEXKeeper.CancelOrder(sdkCtx, creator, orders[0].ID))
}

func (fa *FuzzApp) CancelOrdersByDenom(t *testing.T, sdkCtx sdk.Context, account sdk.AccAddress, denom string) {
	t.Helper()

	// if not ft skip
	if !lo.Contains(fa.ftDenoms, denom) {
		return
	}

	count, err := fa.testApp.DEXKeeper.GetAccountDenomOrdersCount(sdkCtx, account, denom)
	require.NoError(t, err)
	if count == 0 {
		return
	}

	require.LessOrEqual(t, count, fa.testApp.DEXKeeper.GetParams(sdkCtx).MaxOrdersPerDenom)

	require.NoError(t, fa.testApp.DEXKeeper.CancelOrdersByDenom(
		sdkCtx,
		fa.issuer,
		account,
		denom,
	))

	t.Logf("Cancelling orders by denom, count: %d", count)
	count, err = fa.testApp.DEXKeeper.GetAccountDenomOrdersCount(sdkCtx, account, denom)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
}

func FuzzPlaceCancelOrder(f *testing.F) {
	f.Add(uint32(1))
	f.Add(uint32(math.MaxUint32 / 2))
	f.Add(uint32(math.MaxUint32))

	f.Fuzz(
		func(
			t *testing.T,
			rootSeed uint32,
		) {
			fuzzAppConfig := FuzzAppConfig{
				AccountsCount:                 4,
				AssetFTDefaultDenomsCount:     2,
				AssetFTWhitelistingCount:      2,
				AssetFTFreezingDenomsCount:    2,
				AssetFTAllFeaturesDenomsCount: 2,
				NativeDenomCount:              2,

				OrdersCount:                500,
				CancelOrdersPercent:        5,
				CancelOrdersByDenomPercent: 2,

				MarketOrdersPercent:       8,
				TimeInForceIOCPercent:     4,
				TimeInForceFOKPercent:     4,
				GoodTilBlockHeightPercent: 10,
				GoodTilBlockTimePercent:   10,
				FundOrderReservePercent:   80,

				InitialBlockHeight: 1,
				InitialBlockTime:   time.Date(2023, 1, 2, 3, 4, 5, 6, time.UTC),
				BlockTime:          time.Second,
			}

			fuzzApp := NewFuzzApp(
				t,
				fuzzAppConfig,
			)
			rootRnd := rand.New(rand.NewSource(int64(rootSeed)))
			t.Logf("Placing orders and assert final state, rootSeed: %d", rootSeed)
			fuzzApp.PlaceOrdersAndAssertFinalState(t, rootRnd)
		})
}

func randBoolWithPercent(orderRnd *rand.Rand, cancellationPercent int) bool {
	if cancellationPercent == 0 {
		return false
	}
	shouldCancelOrder := randIntInRange(orderRnd, 1, 100) < cancellationPercent
	return shouldCancelOrder
}

func randIntInRange(rnd *rand.Rand, minRange, maxRange int) int {
	return rnd.Intn(maxRange-minRange+1) + minRange
}

func buildNumExpPrice(
	num uint64,
	exp int8,
) (types.Price, bool) {
	numPart := strconv.FormatUint(num, 10)
	// make the price valid if it ends with 0
	validNumPart := strings.TrimRight(numPart, "0")
	if validNumPart == "" {
		// zero price
		return types.Price{}, false
	}
	correction := len(numPart) - len(validNumPart)
	// invalid is exceeds the max int8 value
	if int(exp)+correction > math.MaxInt8 {
		return types.Price{}, false
	}
	numPart = validNumPart
	exp += int8(correction)

	if len(numPart) > types.MaxNumLen {
		return types.Price{}, false
	}
	if exp > types.MaxExp || exp < types.MinExt {
		return types.Price{}, false
	}
	// prepare valid price
	var expPart string
	if exp != 0 {
		expPart = types.ExponentSymbol + strconv.Itoa(int(exp))
	}

	priceStr := numPart + expPart
	return types.MustNewPriceFromString(priceStr), true
}

func genAnyItemByIndex[T any](slice []T, ind uint8) T {
	return slice[int(ind)%len(slice)]
}

func genString(length int, rnd *rand.Rand) string {
	const charset = "abcdefghijklmnopqrstuvwxyz1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b)
}
