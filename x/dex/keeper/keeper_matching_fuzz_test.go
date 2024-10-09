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
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

type FuzzApp struct {
	testApp  *simapp.App
	issuer   sdk.AccAddress
	accounts []sdk.AccAddress
	denoms   []string
	ftDenoms []string
	sides    []types.Side

	marketOrdersPercent int

	timeInForceIOCPercent int
	timeInForceFOKPercent int

	goodTilBlockHeightPercent int
	goodTilBlockTimePercent   int

	initialBlockTime   time.Time
	initialBlockHeight uint64
	blockTime          time.Duration
}

func NewFuzzApp(
	t *testing.T,
	accountsCount,
	assetFTDefaultDenomsCount,
	assetFTWhitelistingCount,
	assetFTFreezingDenomsCount,
	assetFTAllFeaturesDenomsCount,
	nativeDenomCount int,

	marketOrdersPercent int,

	timeInForceIOCPercent int,
	timeInForceFOKPercent int,

	goodTilBlockHeightPercent int,
	goodTilBlockTimePercent int,

	initialBlockHeight uint64,
	initialBlockTime time.Time,
	blockTime time.Duration,
) FuzzApp {
	testApp := simapp.New()

	sdkCtx, _, _ := testApp.BeginNextBlock()

	params := testApp.DEXKeeper.GetParams(sdkCtx)
	params.PriceTickExponent = int32(types.MinExt)
	require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

	accounts := lo.RepeatBy(accountsCount, func(_ int) sdk.AccAddress {
		acc, _ := testApp.GenAccount(sdkCtx)
		return acc
	})

	issuer, _ := testApp.GenAccount(sdkCtx)

	mintedAmount := sdkmath.NewIntWithDecimal(1, 77)
	denoms := lo.RepeatBy(nativeDenomCount, func(i int) string {
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
	ftDenoms = append(ftDenoms, lo.RepeatBy(assetFTDefaultDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("DEF%d", i)
		settings.Subunit = fmt.Sprintf("def%d", i)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(assetFTWhitelistingCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("WLS%d", i)
		settings.Subunit = fmt.Sprintf("wls%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_whitelisting)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(assetFTFreezingDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("FRZ%d", i)
		settings.Subunit = fmt.Sprintf("frz%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_freezing)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(assetFTAllFeaturesDenomsCount, func(i int) string {
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
		testApp:  testApp,
		issuer:   issuer,
		accounts: accounts,
		denoms:   denoms,
		ftDenoms: ftDenoms,
		sides:    sides,

		marketOrdersPercent: marketOrdersPercent,

		timeInForceIOCPercent: timeInForceIOCPercent,
		timeInForceFOKPercent: timeInForceFOKPercent,

		goodTilBlockHeightPercent: goodTilBlockHeightPercent,
		goodTilBlockTimePercent:   goodTilBlockTimePercent,

		initialBlockHeight: initialBlockHeight,
		initialBlockTime:   initialBlockTime,
		blockTime:          blockTime,
	}
}

func (fa *FuzzApp) GenOrder(
	t *testing.T,
	rnd *rand.Rand,
) types.Order {
	t.Helper()

	creator := genAnyItemByIndex(fa.accounts, uint8(rnd.Uint32()))
	orderType := types.ORDER_TYPE_LIMIT
	if randBoolWithPercent(rnd, fa.marketOrdersPercent) {
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

		if randBoolWithPercent(rnd, fa.goodTilBlockHeightPercent) {
			goodTil = &types.GoodTil{
				GoodTilBlockHeight: fa.initialBlockHeight + uint64(randIntInRange(rnd, 0, 2000)),
			}
		}

		if randBoolWithPercent(rnd, fa.goodTilBlockTimePercent) {
			if goodTil == nil {
				goodTil = &types.GoodTil{}
			}
			goodTil.GoodTilBlockTime = lo.ToPtr(
				fa.initialBlockTime.Add(time.Duration(randIntInRange(rnd, 0, 2000)) * fa.blockTime),
			)
		}

		timeInForce = types.TIME_IN_FORCE_GTC

		if randBoolWithPercent(rnd, fa.timeInForceIOCPercent) {
			timeInForce = types.TIME_IN_FORCE_IOC
		}
		if randBoolWithPercent(rnd, fa.timeInForceFOKPercent) {
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
		ID:          uuid.Generate().String(),
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

	spendableBalancesBefore := getSpendableBalances(sdkCtx, fa.testApp, sdk.MustAccAddressFromBech32(order.Creator))
	require.NoError(t, fa.testApp.DEXKeeper.PlaceOrder(sdkCtx, order))

	// check if order is placed
	t.Log("Placement passed")
	assertOrderPlacementResult(t, sdkCtx, fa.testApp, spendableBalancesBefore, order)
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
			const (
				accountsCount                 = 4
				assetFTDefaultDenomsCount     = 2
				assetFTWhitelistingCount      = 2
				assetFTFreezingDenomsCount    = 2
				assetFTAllFeaturesDenomsCount = 2
				nativeDenomCount              = 2
				ordersCount                   = 500

				marketOrdersPercent = 8

				timeInForceIOCPercent = 4
				timeInForceFOKPercent = 4

				goodTilBlockHeightPercent = 10
				goodTilBlockTimePercent   = 10

				cancelOrdersPercent        = 5
				cancelOrdersByDenomPercent = 2

				initialBlockHeight = 1
				blockTime          = time.Second
			)
			initialBlockTime := time.Date(2023, 1, 2, 3, 4, 5, 6, time.UTC)

			fuzzApp := NewFuzzApp(
				t,
				accountsCount,
				assetFTDefaultDenomsCount,
				assetFTWhitelistingCount,
				assetFTFreezingDenomsCount,
				assetFTAllFeaturesDenomsCount,
				nativeDenomCount,

				marketOrdersPercent,

				timeInForceIOCPercent,
				timeInForceFOKPercent,

				goodTilBlockHeightPercent,
				goodTilBlockTimePercent,

				initialBlockHeight,
				initialBlockTime,
				blockTime,
			)
			rootRnd := rand.New(rand.NewSource(int64(rootSeed)))

			sdkCtx := fuzzApp.testApp.NewContextLegacy(false, tmproto.Header{
				Height: initialBlockHeight,
				Time:   initialBlockTime,
			})

			t.Logf("Generating orders with rootSeed: %d", rootSeed)
			for i := 0; i < ordersCount; i++ {
				_, err := fuzzApp.testApp.BeginBlocker(sdkCtx)
				require.NoError(t, err)

				orderSeed := rootRnd.Int63()
				orderRnd := rand.New(rand.NewSource(orderSeed))

				order := fuzzApp.GenOrder(t, orderRnd)

				t.Logf("Placing order, i:%d, seed:%d, order: %s", i, orderSeed, order.String())
				fuzzApp.FundAccountAndApplyFTFeatures(t, sdkCtx, order, orderRnd)
				fuzzApp.PlaceOrder(t, sdkCtx, order)

				if randBoolWithPercent(orderRnd, cancelOrdersPercent) {
					fuzzApp.CancelFirstOrder(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator))
				}

				if randBoolWithPercent(orderRnd, cancelOrdersByDenomPercent) {
					fuzzApp.CancelOrdersByDenom(t, sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.BaseDenom)
				}

				_, err = fuzzApp.testApp.EndBlocker(sdkCtx)
				require.NoError(t, err)

				sdkCtx = fuzzApp.testApp.NewContextLegacy(false, tmproto.Header{
					Height: sdkCtx.BlockHeight() + 1,
					Time:   sdkCtx.BlockTime().Add(blockTime),
				})
			}
			cancelAllOrdersAndAssertState(t, sdkCtx, fuzzApp.testApp)
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
