package feemodel

import (
	"math"
	"math/rand"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	feeModel = Model{
		InitialGasPrice:         sdk.NewInt(1500),
		MaxGasPrice:             sdk.NewInt(1500000),
		MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
		EscalationStartBlockGas: 700,
		MaxBlockGas:             1000,
	}

	gasPriceWithMaxDiscount = feeModel.computeGasPriceWithMaxDiscount()
)

func TestCalculateNextGasPriceKeyPoints(t *testing.T) {
	// at point 0 it should be initial price

	nextGasPrice := feeModel.CalculateNextGasPrice(0, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.InitialGasPrice))

	// if current average gas is equal to average gas it should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(100, 100)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	// if current average gas equals escalation start block gas it still should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas, 100)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	// if current average gas is equal to MaxBlockGas it should be max gas price number

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.MaxBlockGas, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.MaxGasPrice))

	// if current average gas is greater than MaxBlockGas it should stay the same

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.MaxBlockGas+100, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.MaxGasPrice))
}

func TestAverageGasBeyondEscalationStartBlockGas(t *testing.T) {
	// There is a special case when long average block gas is higher than escalation start block gas.
	// The question is if in such scenario we should offer discounted gas price or escalation should be applied instead.
	// It seems obvious that price should be escalated.

	nextGasPrice := feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas+100)
	assert.True(t, nextGasPrice.GT(gasPriceWithMaxDiscount))

	// Next gas price should be the same as for long average block gas being below optimal block gas.
	// It means that escalation was turned on.

	nextGasPrice2 := feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas-100)
	assert.True(t, nextGasPrice2.Equal(nextGasPrice))
}

func TestZeroAverageGas(t *testing.T) {
	nextGasPrice := feeModel.CalculateNextGasPrice(0, 0)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	nextGasPrice = feeModel.CalculateNextGasPrice(1, 0)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))
}

func TestShapeInDecreasingRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := feeModel.InitialGasPrice
	for i := int64(1); i <= longAverageBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas)
		assert.True(t, nextPrice.LTE(lastPrice))
		lastPrice = nextPrice
	}
}

func TestShapeInFlatRegion(t *testing.T) {
	const longAverageBlockGas = 100

	for i := int64(longAverageBlockGas); i <= feeModel.EscalationStartBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas)
		assert.True(t, nextPrice.Equal(gasPriceWithMaxDiscount))
	}
}

func TestShapeInEscalationRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := gasPriceWithMaxDiscount
	for i := feeModel.EscalationStartBlockGas + 1; i <= feeModel.MaxBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas)
		assert.True(t, nextPrice.GT(lastPrice))

		lastPrice = nextPrice
	}
}

func TestWithRandomModels(t *testing.T) {
	t.Parallel()

	for i := 0; i < 100; i++ {
		t.Run("RandomCase", func(t *testing.T) {
			t.Parallel()

			model, shortEMA, longEMA := generateRandomizedModel()
			logParameters(t, model, shortEMA, longEMA)
			nextGasPrice := model.CalculateNextGasPrice(shortEMA, longEMA)

			assert.True(t, nextGasPrice.GT(sdk.ZeroInt()))
			assert.True(t, nextGasPrice.LTE(model.MaxGasPrice))

			switch {
			case shortEMA >= model.MaxBlockGas:
				assert.True(t, nextGasPrice.Equal(model.MaxGasPrice))
			case shortEMA > model.EscalationStartBlockGas:
				assert.True(t, nextGasPrice.LT(model.MaxGasPrice))
				assert.True(t, nextGasPrice.GTE(model.computeGasPriceWithMaxDiscount()))
			case shortEMA >= longEMA:
				assert.True(t, nextGasPrice.Equal(model.computeGasPriceWithMaxDiscount()))
			case shortEMA > 0:
				assert.True(t, nextGasPrice.LT(model.InitialGasPrice))
			default:
				assert.True(t, nextGasPrice.Equal(model.InitialGasPrice))
			}
		})
	}
}

func generateRandomizedModel() (model Model, shortEMA, longEMA int64) {
	// fuzz engine of go test is not used because it doesn't allow to maintain relationships between generated parameters

	initialGasPrice := rand.Uint64()/2 + 10
	maxGasPrice := initialGasPrice + uint64(float64(math.MaxUint64-initialGasPrice-10)*rand.Float64()+10.0)
	shortAverageBlockLength := uint(rand.Intn(1000) + 1)
	longAverageBlockLength := shortAverageBlockLength + uint(rand.Intn(10000))
	maxBlockGas := rand.Int63()
	escalationStartBlockGas := rand.Int63n(maxBlockGas-2) + 1
	// rand.Float64() returns random number in [0.0, 1.0), but we want (0.0, 1.0],
	// calculating 1.0 - rand.Float64() is the simplest way of achieving this.
	maxDiscount := (1.0 - rand.Float64()) * 0.8 // this gives a number in range (0.0, 0.8]

	shortEMA = rand.Int63()
	longEMA = rand.Int63()

	return Model{
		InitialGasPrice:         sdk.NewIntFromUint64(initialGasPrice),
		MaxGasPrice:             sdk.NewIntFromUint64(maxGasPrice),
		MaxDiscount:             sdk.MustNewDecFromStr(strconv.FormatFloat(maxDiscount, 'f', 4, 64)),
		EscalationStartBlockGas: escalationStartBlockGas,
		MaxBlockGas:             maxBlockGas,
		ShortAverageBlockLength: shortAverageBlockLength,
		LongAverageBlockLength:  longAverageBlockLength,
	}, shortEMA, longEMA
}

func logParameters(t *testing.T, model Model, shortEMA, longEMA int64) {
	t.Logf(`InitialGasPrice: %s
MaxGasPrice: %s
MaxDiscount: %s
EscalationStartBlockGas: %d
MaxBlockGas: %d
ShortAverageBlockLength: %d
LongAverageBlockLength: %d
ShortEMA: %d
LongEMA: %d`,
		model.InitialGasPrice,
		model.MaxGasPrice,
		model.MaxDiscount,
		model.EscalationStartBlockGas,
		model.MaxBlockGas,
		model.ShortAverageBlockLength,
		model.LongAverageBlockLength,
		shortEMA,
		longEMA)
}
