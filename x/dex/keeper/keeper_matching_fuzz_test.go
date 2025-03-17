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
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/testutil/simapp"
	testcontracts "github.com/CoreumFoundation/coreum/v5/x/asset/ft/keeper/test-contracts"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

const (
	IDDEXOrderSuffixTrigger = "blocked"
)

type FuzzAppConfig struct {
	AccountsCount                 int
	AssetFTDefaultDenomsCount     int
	AssetFTWhitelistingCount      int
	AssetFTFreezingDenomsCount    int
	AssetFTExtensionDenomsCount   int
	AssetFTAllFeaturesDenomsCount int
	NativeDenomCount              int

	// used for both default and asset ft
	UnifiedRefAmountChangePercent  int
	WhitelistedDenomsChangePercent int

	OrdersCount                int
	CancelOrdersPercent        int
	CancelOrdersByDenomPercent int

	MarketOrdersPercent             int
	TimeInForceIOCPercent           int
	TimeInForceFOKPercent           int
	GoodTilBlockHeightPercent       int
	GoodTilBlockTimePercent         int
	ProhibitedExtensionOrderPercent int

	FundOrderReservePercent     int
	CreateVestingAccountPercent int

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

	failedPlaceOrderCount int
}

func NewFuzzApp(
	t *testing.T,
	cfg FuzzAppConfig,
) FuzzApp {
	testApp := simapp.New()

	sdkCtx, _, _ := testApp.BeginNextBlock()

	params, err := testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	// use smaller values than default ones in case we decide to decrease it
	params.PriceTickExponent -= 10
	params.QuantityStepExponent -= 10

	require.NoError(t, testApp.DEXKeeper.SetParams(sdkCtx, params))

	accounts := lo.RepeatBy(cfg.AccountsCount, func(_ int) sdk.AccAddress {
		// gen address but don't register it to allow the test register vesting accounts conditionally
		return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	})

	issuer, _ := testApp.GenAccount(sdkCtx)

	mintedAmount := sdkmath.NewIntWithDecimal(1, 77)
	denoms := lo.RepeatBy(cfg.NativeDenomCount, func(i int) string {
		denom := fmt.Sprintf("native-denom-%d", i)
		testApp.MintAndSendCoin(t, sdkCtx, issuer, sdk.NewCoins(sdk.NewCoin(denom, mintedAmount)))
		return denom
	})

	extensionCodeID, _, err := testApp.WasmPermissionedKeeper.Create(
		sdkCtx, issuer, testcontracts.AssetExtensionWasm, &wasmtypes.AllowEverybody,
	)
	require.NoError(t, err)

	ftDenoms := make([]string, 0)
	defaultFTSettings := assetfttypes.IssueSettings{
		Issuer:        issuer,
		Precision:     1,
		InitialAmount: mintedAmount,
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_dex_order_cancellation,
			assetfttypes.Feature_dex_unified_ref_amount_change,
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
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTFreezingDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("WHD%d", i)
		settings.Subunit = fmt.Sprintf("whd%d", i)
		settings.Features = append(settings.Features, assetfttypes.Feature_dex_whitelisted_denoms)
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTExtensionDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("EXT%d", i)
		settings.Subunit = fmt.Sprintf("ext%d", i)
		settings.Features = append(
			settings.Features,
			assetfttypes.Feature_extension,
		)
		settings.ExtensionSettings = &assetfttypes.ExtensionIssueSettings{
			CodeId: extensionCodeID,
		}
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	ftDenoms = append(ftDenoms, lo.RepeatBy(cfg.AssetFTAllFeaturesDenomsCount, func(i int) string {
		settings := defaultFTSettings
		settings.Symbol = fmt.Sprintf("ALL%d", i)
		settings.Subunit = fmt.Sprintf("all%d", i)
		settings.Features = append(
			settings.Features,
			assetfttypes.Feature_whitelisting,
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_dex_whitelisted_denoms,
			assetfttypes.Feature_extension,
		)
		settings.ExtensionSettings = &assetfttypes.ExtensionIssueSettings{
			CodeId: extensionCodeID,
		}
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)
	denoms = append(denoms, ftDenoms...)

	sides := []types.Side{
		types.SIDE_BUY,
		types.SIDE_SELL,
	}

	_, err = testApp.EndBlocker(sdkCtx)
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

	for i := range fa.cfg.OrdersCount {
		_, err := fa.testApp.BeginBlocker(sdkCtx)
		require.NoError(t, err)

		orderSeed := rootRnd.Int63()
		orderRnd := rand.New(rand.NewSource(orderSeed))

		// every iteration we might change some app/denoms params
		fa.AdjustAppState(t, sdkCtx, orderRnd)

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

	require.LessOrEqual(t, float64(fa.failedPlaceOrderCount), float64(fa.cfg.OrdersCount)*0.8, "More than 80% of orders failed to be placed")
}

func (fa *FuzzApp) GenOrder(
	t *testing.T,
	rnd *rand.Rand,
) types.Order {
	t.Helper()

	creator := getAnyItemByIndex(fa.accounts, uint8(rnd.Uint32()))
	orderType := types.ORDER_TYPE_LIMIT
	if randBoolWithPercent(rnd, fa.cfg.MarketOrdersPercent) {
		orderType = types.ORDER_TYPE_MARKET
	}

	// take next if same
	var baseDenom, quoteDenom string
	baseDenomInd := uint8(rnd.Uint32())
	for {
		baseDenom = getAnyItemByIndex(fa.denoms, uint8(rnd.Uint32()))
		quoteDenom = getAnyItemByIndex(fa.denoms, baseDenomInd)
		if baseDenom != quoteDenom {
			break
		}
		baseDenomInd++
	}

	side := getAnyItemByIndex(fa.sides, uint8(rnd.Uint32()))

	var (
		price       *types.Price
		goodTil     *types.GoodTil
		timeInForce = types.TIME_IN_FORCE_UNSPECIFIED
	)

	if orderType == types.ORDER_TYPE_LIMIT { //nolint:nestif  // the ifs are simple to check the percents mostly
		priceNum := rnd.Uint32()

		// generate price exponent in order not to overflow the sdkmath.Int when fund accounts
		priceExp := int8(randIntInRange(rnd, 5, 10))

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
	} else if orderType == types.ORDER_TYPE_MARKET {
		timeInForce = types.TIME_IN_FORCE_IOC
	}

	// the quantity can't be zero
	quantity := uint64((rnd.Int63n(1_000_000) + 1) * 1_000_000_000)

	var orderIDSuffix string
	if randBoolWithPercent(rnd, fa.cfg.ProhibitedExtensionOrderPercent) {
		orderIDSuffix = IDDEXOrderSuffixTrigger
	}

	return types.Order{
		Creator:     creator.String(),
		Type:        orderType,
		ID:          randString(20, rnd) + orderIDSuffix,
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
			"Failed to fund, insufficient issuer balance, balance:%s, recipient:%s, coin:%s",
			issuerBalance.String(), creator.String(), fundCoin.String(),
		)
	} else {
		if spendDef.IsFeatureEnabled(assetfttypes.Feature_whitelisting) {
			dexExpectedToReceiveBalance := fa.testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				sdkCtx, creator, fundCoin.Denom,
			)
			balance := fa.testApp.BankKeeper.GetBalance(sdkCtx, creator, fundCoin.Denom)
			whitelistBalance := fundCoin.Add(balance).Add(dexExpectedToReceiveBalance)
			t.Logf("Whitelisting initial account's balance: %s, %s", creator.String(), whitelistBalance.String())
			require.NoError(t, fa.testApp.AssetFTKeeper.SetWhitelistedBalance(sdkCtx, fa.issuer, creator, whitelistBalance))
		}
		if randBoolWithPercent(orderRnd, fa.cfg.CreateVestingAccountPercent) &&
			// acc doesn't exist so we can create vesting account
			fa.testApp.AccountKeeper.GetAccount(sdkCtx, creator) == nil {
			t.Logf("Creating vesting account: %s", creator.String())
			// create acc with permanently vesting locked coins
			baseVestingAccount, err := vestingtypes.NewDelayedVestingAccount(
				authtypes.NewBaseAccountWithAddress(creator),
				sdk.NewCoins(fundCoin),
				math.MaxInt64,
			)
			require.NoError(t, err)
			account := fa.testApp.App.AccountKeeper.NewAccount(sdkCtx, baseVestingAccount)
			fa.testApp.AccountKeeper.SetAccount(sdkCtx, account)
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
		params, err := fa.testApp.DEXKeeper.GetParams(sdkCtx)
		require.NoError(t, err)
		reserve := params.OrderReserve

		spendableBalance, err := fa.testApp.AssetFTKeeper.GetSpendableBalance(sdkCtx, creator, reserve.Denom)
		require.NoError(t, err)
		if spendableBalance.IsLT(reserve) {
			t.Logf("Funding order reserve, account: %s coin: %s", creator.String(), reserve.String())
			fa.testApp.MintAndSendCoin(t, sdkCtx, creator, sdk.NewCoins(reserve))
		}
	}
}

func (fa *FuzzApp) AdjustAppState(t *testing.T, sdkCtx sdk.Context, rnd *rand.Rand) {
	// change unified ref amount
	if randBoolWithPercent(rnd, fa.cfg.UnifiedRefAmountChangePercent) {
		// change globally
		if randBool(rnd) {
			params, err := fa.testApp.DEXKeeper.GetParams(sdkCtx)
			require.NoError(t, err)

			params.DefaultUnifiedRefAmount = randPositiveSDKDec(rnd)
			t.Logf("Updating new default unified ref amount: %s", params.DefaultUnifiedRefAmount.String())
			require.NoError(t, fa.testApp.DEXKeeper.SetParams(
				sdkCtx,
				params,
			))
		} else {
			denom := getAnyItemByIndex(fa.ftDenoms, uint8(rnd.Uint32()))
			unifiedRefAmount := randPositiveSDKDec(rnd)
			t.Logf("Updating new denom %s unified ref amount: %s", denom, unifiedRefAmount.String())
			require.NoError(t, fa.testApp.AssetFTKeeper.UpdateDEXUnifiedRefAmount(
				sdkCtx,
				fa.issuer,
				denom,
				unifiedRefAmount,
			))
		}
	}

	for _, denom := range fa.ftDenoms {
		def, err := fa.testApp.AssetFTKeeper.GetDefinition(sdkCtx, denom)
		require.NoError(t, err)
		if !def.IsFeatureEnabled(assetfttypes.Feature_dex_whitelisted_denoms) {
			continue
		}
		if randBoolWithPercent(rnd, fa.cfg.WhitelistedDenomsChangePercent) {
			var whitelistedDenoms []string
			if randBool(rnd) {
				// change whitelisted denoms
				whitelistedDenoms = randItemsFromSlice(fa.ftDenoms, rnd)
			} else {
				// remove whitelisted denoms
				whitelistedDenoms = make([]string, 0)
			}
			t.Logf("Updating %s whitelisted denoms: %v", denom, whitelistedDenoms)

			require.NoError(t, fa.testApp.AssetFTKeeper.UpdateDEXWhitelistedDenoms(
				sdkCtx, sdk.MustAccAddressFromBech32(def.Admin), denom, whitelistedDenoms,
			))
		}
	}
}

func (fa *FuzzApp) PlaceOrder(t *testing.T, sdkCtx sdk.Context, order types.Order) {
	t.Helper()

	trialCtx := simapp.CopyContextWithMultiStore(sdkCtx) // copy to dry run and don't change state if error
	if err := fa.testApp.DEXKeeper.PlaceOrder(trialCtx, order); err != nil {
		fa.failedPlaceOrderCount++
		t.Logf("Placement failed, err: %s", err.Error())
		creator := sdk.MustAccAddressFromBech32(order.Creator)
		switch {
		case sdkerrors.IsOf(err, assetfttypes.ErrDEXInsufficientSpendableBalance):
			// check that the order can't be placed because of the lack of balance
			if order.Type != types.ORDER_TYPE_LIMIT {
				return
			}
			// check failed because of reserve
			params, err := fa.testApp.DEXKeeper.GetParams(sdkCtx)
			require.NoError(t, err)
			reserve := params.OrderReserve
			reserveDenomSpendableBalance, err := fa.testApp.AssetFTKeeper.GetSpendableBalance(
				sdkCtx, creator, reserve.Denom,
			)
			require.NoError(t, err)
			if reserveDenomSpendableBalance.Amount.LT(reserve.Amount) {
				t.Logf("Placement is failed due to insufficient reserve, reserve: %s, reserveDenomSpendableBalance: %s",
					reserve.String(), reserveDenomSpendableBalance.String())
				return
			}

			// check spendable balance
			spendableBalance, err := fa.testApp.AssetFTKeeper.GetSpendableBalance(
				sdkCtx, creator, order.GetSpendDenom(),
			)
			require.NoError(t, err)
			orderLockedBalance, err := order.ComputeLimitOrderLockedBalance()
			require.NoError(t, err)
			require.True(
				t,
				spendableBalance.IsLT(orderLockedBalance),
				"availableBalance: %s, orderLockedBalance: %s", spendableBalance.String(), orderLockedBalance.String(),
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
				requiredWhitelistedBalance, err = types.ComputeLimitOrderExpectedToReceiveBalance(
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
			dexExpectedToReceiveBalance := fa.testApp.AssetFTKeeper.GetDEXExpectedToReceivedBalance(
				sdkCtx, creator, requiredWhitelistedBalance.Denom,
			)
			receivableAmt := whitelistedBalance.Amount.Sub(balance.Amount).Sub(dexExpectedToReceiveBalance.Amount)
			require.True(
				t,
				receivableAmt.LT(requiredWhitelistedAmt),
				"receivableAmt: %s, requiredWhitelistedAmt: %s",
				receivableAmt.String(), requiredWhitelistedAmt.String(),
			)
			return
		case sdkerrors.IsOf(err, assetfttypes.ErrExtensionCallFailed):
			t.Logf("Placement has failed due to extension call error: %v", err.Error())
			return
		case strings.Contains(err.Error(), "has to be multiple of price tick"),
			strings.Contains(err.Error(), "has to be multiple of quantity step"),
			strings.Contains(err.Error(), "good til block"),
			strings.Contains(err.Error(), "it's prohibited to save more than"),
			strings.Contains(err.Error(), "not whitelisted for"): // whitelisted denoms
			t.Logf("Placement has failed due to expected error: %v", err.Error())
			return
		default:
			require.NoError(t, err)
		}
	}

	availableBalancesBefore, err := getAvailableBalances(sdkCtx, fa.testApp, sdk.MustAccAddressFromBech32(order.Creator))
	require.NoError(t, err)

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

	params, err := fa.testApp.DEXKeeper.GetParams(sdkCtx)
	require.NoError(t, err)
	require.LessOrEqual(t, count, params.MaxOrdersPerDenom)

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
				AssetFTExtensionDenomsCount:   2,
				AssetFTAllFeaturesDenomsCount: 2,
				NativeDenomCount:              2,

				UnifiedRefAmountChangePercent:  10,
				WhitelistedDenomsChangePercent: 5,

				OrdersCount:                500,
				CancelOrdersPercent:        5,
				CancelOrdersByDenomPercent: 2,

				MarketOrdersPercent:             8,
				TimeInForceIOCPercent:           4,
				TimeInForceFOKPercent:           4,
				GoodTilBlockHeightPercent:       10,
				GoodTilBlockTimePercent:         10,
				ProhibitedExtensionOrderPercent: 10,

				FundOrderReservePercent:     80,
				CreateVestingAccountPercent: 25, // 25% of accounts will be vesting accounts

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
	if exp > types.MaxExp || exp < types.MinExp {
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

func randBoolWithPercent(orderRnd *rand.Rand, cancellationPercent int) bool {
	if cancellationPercent == 0 {
		return false
	}
	shouldCancelOrder := randIntInRange(orderRnd, 1, 100) < cancellationPercent
	return shouldCancelOrder
}

func randBool(orderRnd *rand.Rand) bool {
	return randIntInRange(orderRnd, 1, 100)%2 == 0
}

func randPositiveSDKDec(rnd *rand.Rand) sdkmath.LegacyDec {
	v := sdkmath.LegacyMustNewDecFromStr(fmt.Sprintf("%.18f", rnd.NormFloat64()))
	multiplier := sdkmath.NewIntWithDecimal(1, randIntInRange(rnd, 1, 20))
	if v.IsNegative() {
		v = v.QuoInt(multiplier).Neg()
	} else {
		v = v.MulInt(multiplier)
	}
	if v.IsZero() {
		v = sdkmath.LegacyOneDec()
	}

	return v
}

func randString(length int, rnd *rand.Rand) string {
	const charset = "abcdefghijklmnopqrstuvwxyz1234567890"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}
	return string(b)
}

func randItemsFromSlice[T any](slice []T, rnd *rand.Rand) []T {
	// shuffle the slice
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
	n := randIntInRange(rnd, 0, len(slice))
	return slice[:n]
}

func randIntInRange(rnd *rand.Rand, minRange, maxRange int) int {
	return rnd.Intn(maxRange-minRange+1) + minRange
}

func getAnyItemByIndex[T any](slice []T, ind uint8) T {
	return slice[int(ind)%len(slice)]
}
