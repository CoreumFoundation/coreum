package keeper_test

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/docker/distribution/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

type FuzzApp struct {
	testApp    *simapp.App
	issuer     sdk.AccAddress
	accounts   []sdk.AccAddress
	orderTypes []types.OrderType
	denoms     []string
	sides      []types.Side
}

func NewFuzzApp(t *testing.T, accountsCount, assetFTDenomsCount, nativeDenomCount int) FuzzApp {
	testApp := simapp.New()

	sdkCtx, _, _ := testApp.BeginNextBlock(time.Now())
	accounts := lo.RepeatBy(accountsCount, func(_ int) sdk.AccAddress {
		acc, _ := testApp.GenAccount(sdkCtx)
		return acc
	})

	orderTypes := []types.OrderType{
		types.ORDER_TYPE_LIMIT,
		types.ORDER_TYPE_MARKET,
	}

	issuer, _ := testApp.GenAccount(sdkCtx)

	mintedAmount := sdkmath.NewIntWithDecimal(1, 77)
	denoms := lo.RepeatBy(nativeDenomCount, func(i int) string {
		denom := fmt.Sprintf("native-denom-%d", i)
		testApp.MintAndSendCoin(t, sdkCtx, issuer, sdk.NewCoins(sdk.NewCoin(denom, mintedAmount)))
		return denom
	})

	denoms = append(denoms, lo.RepeatBy(assetFTDenomsCount, func(i int) string {
		settings := assetfttypes.IssueSettings{
			Issuer:        issuer,
			Symbol:        fmt.Sprintf("SMB%d", i),
			Subunit:       fmt.Sprintf("sut%d", i),
			Precision:     1,
			InitialAmount: mintedAmount,
		}
		denom, err := testApp.AssetFTKeeper.Issue(sdkCtx, settings)
		require.NoError(t, err)
		return denom
	})...)

	sides := []types.Side{
		types.SIDE_BUY,
		types.SIDE_SELL,
	}

	_, err := testApp.EndBlocker(sdkCtx)
	require.NoError(t, err)

	return FuzzApp{
		testApp:    testApp,
		issuer:     issuer,
		accounts:   accounts,
		orderTypes: orderTypes,
		denoms:     denoms,
		sides:      sides,
	}
}

func (fa *FuzzApp) GenOrder(
	t *testing.T,
	rnd *rand.Rand,
) types.Order {
	t.Helper()

	creator := genAnyItemByIndex(fa.accounts, uint8(rnd.Uint32()))
	orderType := genAnyItemByIndex(fa.orderTypes, uint8(rnd.Uint32()))
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
	const minExp, maxExp = -10, 10
	priceExp := int8(rnd.Intn(maxExp-minExp+1) + minExp)

	var price *types.Price
	if orderType == types.ORDER_TYPE_LIMIT {
		v, ok := buildNumExpPrice(uint64(priceNum), priceExp)
		// since we use Uint32 as num it never exceed the max num length
		require.True(t, ok)
		price = &v
	}

	// the quantity can't be zero
	quantity := rnd.Uint64()
	if quantity == 0 {
		quantity = 1
	}

	return types.Order{
		Creator:    creator.String(),
		Type:       orderType,
		ID:         uuid.Generate().String(),
		BaseDenom:  baseDenom,
		QuoteDenom: quoteDenom,
		Price:      price,
		Quantity:   sdkmath.NewIntFromUint64(quantity),
		Side:       side,
	}
}

func (fa *FuzzApp) FundAccount(t *testing.T, sdkCtx sdk.Context, recipient sdk.AccAddress, coin sdk.Coin) {
	t.Helper()

	issuerBalance := fa.testApp.BankKeeper.GetBalance(sdkCtx, fa.issuer, coin.Denom)
	if issuerBalance.IsLT(coin) {
		t.Logf(
			"Failed to fund, insufficient issuer balance, balance:%s, recipient:%s,  coin:%s",
			issuerBalance.String(), recipient.String(), coin.String(),
		)
		return
	}
	require.NoError(t, fa.testApp.BankKeeper.SendCoins(sdkCtx, fa.issuer, recipient, sdk.NewCoins(coin)))
}

func (fa *FuzzApp) PlaceOrder(t *testing.T, sdkCtx sdk.Context, order types.Order) {
	t.Helper()

	spendableBalancesBefore := getSpendableBalances(sdkCtx, fa.testApp, sdk.MustAccAddressFromBech32(order.Creator))
	err := fa.testApp.DEXKeeper.PlaceOrder(sdkCtx, order)
	if err != nil {
		t.Logf("Placement failed, err: %s", err.Error())
		// expected fail
		require.ErrorIs(t, err, types.ErrFailedToLockCoin)
		return
	}
	require.NoError(t, err)
	t.Log("Placement passed")
	assertOrderPlacementResult(t, sdkCtx, fa.testApp, spendableBalancesBefore, order)
}

func (fa *FuzzApp) CancelOrder(t *testing.T, sdkCtx sdk.Context, order types.Order) {
	t.Helper()

	_, err := fa.testApp.DEXKeeper.GetOrderByAddressAndID(sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID)
	if err != nil {
		require.ErrorIs(t, err, types.ErrRecordNotFound)
		t.Logf("Order to cancel not found.")
		return
	}
	t.Logf("Cancelling order.")
	require.NoError(t, fa.testApp.DEXKeeper.CancelOrder(sdkCtx, sdk.MustAccAddressFromBech32(order.Creator), order.ID))
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
				accountsCount       = 10
				assetFTDenomsCount  = 3
				nativeDenomCount    = 3
				ordersCount         = 1000
				cancellationPercent = 5 // cancel 5% of limit orders
			)
			fuzzApp := NewFuzzApp(t, accountsCount, assetFTDenomsCount, nativeDenomCount)
			rootRnd := rand.New(rand.NewSource(int64(rootSeed)))

			// process generated orders
			sdkCtx, _, _ := fuzzApp.testApp.BeginNextBlock(time.Now())

			t.Logf("Generating orders with rootSeed: %d", rootSeed)
			for i := 0; i < ordersCount; i++ {
				orderSeed := rootRnd.Int63()
				orderRnd := rand.New(rand.NewSource(orderSeed))

				order := fuzzApp.GenOrder(t, orderRnd)

				var fundDenom string
				switch order.Side {
				case types.SIDE_BUY:
					fundDenom = order.QuoteDenom
				case types.SIDE_SELL:
					fundDenom = order.BaseDenom
				default:
					t.Fatalf("Unsupported order side: %s", order.Side.String())
				}

				// fund creator with random quantity,
				creator := sdk.MustAccAddressFromBech32(order.Creator)

				balance := orderRnd.Uint64()
				coin := sdk.NewCoin(fundDenom, sdkmath.NewIntFromUint64(balance))
				t.Logf("Funding account for the order placement, addr:%s, coin:%s", order.Creator, coin.String())
				fuzzApp.FundAccount(t, sdkCtx, creator, coin)
				t.Logf("Placing order, i:%d, seed:%d, order: %s", i, orderSeed, order.String())
				fuzzApp.PlaceOrder(t, sdkCtx, order)

				if order.Type == types.ORDER_TYPE_LIMIT {
					// simulate percent based cancellation
					const minRange, maxRange = 1, 100
					shouldCancelOrder := orderRnd.Intn(maxRange-minRange+1)+minRange < cancellationPercent
					if shouldCancelOrder {
						fuzzApp.CancelOrder(t, sdkCtx, order)
					}
				}
			}

			_, err := fuzzApp.testApp.EndBlocker(sdkCtx)
			require.NoError(t, err)
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
