package types

import (
	"math/rand"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	feeModel = Model{
		params: params.Model,
	}

	gasPriceWithMaxDiscount = feeModel.CalculateGasPriceWithMaxDiscount()
)

func TestCalculateNextGasPriceKeyPoints(t *testing.T) {
	// at point 0 it should be initial price

	nextGasPrice := feeModel.CalculateNextGasPrice(0, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.params.InitialGasPrice))

	// if current average gas is equal to average gas it should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(100, 100)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	// if current average gas equals escalation start block gas it still should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.CalculateEscalationStartBlockGas(), 100)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	// if current average gas is equal to MaxBlockGas it should be max gas price number

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.params.MaxBlockGas, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.CalculateMaxGasPrice()))

	// if current average gas is greater than MaxBlockGas it should stay the same

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.params.MaxBlockGas+100, 100)
	assert.True(t, nextGasPrice.Equal(feeModel.CalculateMaxGasPrice()))
}

func TestEMAGasBeyondEscalationStartBlockGas(t *testing.T) {
	// There is a special case when long average block gas is higher than escalation start block gas.
	// The question is if in such scenario we should offer discounted gas price or escalation should be applied instead.
	// It seems obvious that price should be escalated.

	nextGasPrice := feeModel.CalculateNextGasPrice(feeModel.CalculateEscalationStartBlockGas()+150, feeModel.CalculateEscalationStartBlockGas()+100)
	assert.True(t, nextGasPrice.GT(gasPriceWithMaxDiscount))

	// Next gas price should be the same as for long average block gas being below optimal block gas.
	// It means that escalation was turned on.

	nextGasPrice2 := feeModel.CalculateNextGasPrice(feeModel.CalculateEscalationStartBlockGas()+150, feeModel.CalculateEscalationStartBlockGas()-100)
	assert.True(t, nextGasPrice2.Equal(nextGasPrice))
}

func TestZeroEMAGas(t *testing.T) {
	nextGasPrice := feeModel.CalculateNextGasPrice(0, 0)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))

	nextGasPrice = feeModel.CalculateNextGasPrice(1, 0)
	assert.True(t, nextGasPrice.Equal(gasPriceWithMaxDiscount))
}

func TestShapeInDecreasingRegion(t *testing.T) {
	const longEMABlockGas = 100

	lastPrice := feeModel.params.InitialGasPrice
	for i := int64(1); i <= longEMABlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longEMABlockGas)
		assert.True(t, nextPrice.LTE(lastPrice))
		lastPrice = nextPrice
	}
}

func TestShapeInFlatRegion(t *testing.T) {
	const longEMABlockGas = 100

	for i := int64(longEMABlockGas); i <= feeModel.CalculateEscalationStartBlockGas(); i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longEMABlockGas)
		assert.True(t, nextPrice.Equal(gasPriceWithMaxDiscount))
	}
}

func TestShapeInEscalationRegion(t *testing.T) {
	const longEMABlockGas = 100

	lastPrice := gasPriceWithMaxDiscount
	for i := feeModel.CalculateEscalationStartBlockGas() + 1; i <= feeModel.params.MaxBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longEMABlockGas)
		assert.True(t, nextPrice.GT(lastPrice))

		lastPrice = nextPrice
	}
}

func TestWithRandomModels(t *testing.T) {
	t.Parallel()

	for i := 0; i < 100; i++ {
		t.Run("RandomCase", func(t *testing.T) {
			t.Parallel()

			params, shortEMA, longEMA := generateRandomizedParams()
			modelParams := params.Model
			model := Model{
				params: modelParams,
			}
			logParameters(t, modelParams, shortEMA, longEMA)
			nextGasPrice := model.CalculateNextGasPrice(shortEMA, longEMA)

			assert.True(t, nextGasPrice.GT(sdk.ZeroDec()))
			assert.True(t, nextGasPrice.LTE(model.CalculateMaxGasPrice()))

			switch {
			case shortEMA >= modelParams.MaxBlockGas:
				assert.True(t, nextGasPrice.Equal(model.CalculateMaxGasPrice()))
			case shortEMA > model.CalculateEscalationStartBlockGas():
				assert.True(t, nextGasPrice.LT(model.CalculateMaxGasPrice()))
				assert.True(t, nextGasPrice.GTE(model.CalculateGasPriceWithMaxDiscount()))
			case shortEMA >= longEMA:
				assert.True(t, nextGasPrice.Equal(model.CalculateGasPriceWithMaxDiscount()))
			case shortEMA > 0:
				assert.True(t, nextGasPrice.LT(modelParams.InitialGasPrice))
			default:
				assert.True(t, nextGasPrice.Equal(modelParams.InitialGasPrice))
			}
		})
	}
}

func generateRandomizedParams() (params Params, shortEMA, longEMA int64) {
	// fuzz engine of go test is not used because it doesn't allow to maintain relationships between generated parameters

	initialGasPrice := rand.Uint64()/2 + 10
	maxGasPriceMultiplier := uint64(1000.0*rand.Float64() + 10.0)
	shortEMABlockLength := uint32(rand.Int31n(1000) + 1)
	longEMABlockLength := shortEMABlockLength + uint32(rand.Int31n(10000))
	maxBlockGas := rand.Int63()
	escalationStartFraction := 0.8*rand.Float64() + 0.1
	// rand.Float64() returns random number in [0.0, 1.0), but we want (0.0, 1.0],
	// calculating 1.0 - rand.Float64() is the simplest way of achieving this.
	maxDiscount := (1.0 - rand.Float64()) * 0.8 // this gives a number in range (0.0, 0.8]

	shortEMA = rand.Int63()
	longEMA = rand.Int63()

	return Params{
		Model: ModelParams{
			InitialGasPrice:         sdk.NewIntFromUint64(initialGasPrice).ToDec(),
			MaxGasPriceMultiplier:   sdk.NewIntFromUint64(maxGasPriceMultiplier).ToDec(),
			MaxDiscount:             sdk.MustNewDecFromStr(strconv.FormatFloat(maxDiscount, 'f', 4, 64)),
			EscalationStartFraction: sdk.MustNewDecFromStr(strconv.FormatFloat(escalationStartFraction, 'f', 4, 64)),
			MaxBlockGas:             maxBlockGas,
			ShortEmaBlockLength:     shortEMABlockLength,
			LongEmaBlockLength:      longEMABlockLength,
		},
	}, shortEMA, longEMA
}

func logParameters(t *testing.T, params ModelParams, shortEMA, longEMA int64) {
	t.Logf(`InitialGasPrice: %s
MaxGasPriceMultiplier: %s
MaxDiscount: %s
EscalationStartFraction: %d
MaxBlockGas: %d
ShortEMABlockLength: %d
LongEMABlockLength: %d
ShortEMA: %d
LongEMA: %d`,
		params.InitialGasPrice,
		params.MaxGasPriceMultiplier,
		params.MaxDiscount,
		params.EscalationStartFraction,
		params.MaxBlockGas,
		params.ShortEmaBlockLength,
		params.LongEmaBlockLength,
		shortEMA,
		longEMA)
}
