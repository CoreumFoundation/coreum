package fee

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	initialGasPrice       int64 = 100000000
	maxDiscount                 = 0.2
	minDiscountedGasPrice       = int64((1. - maxDiscount) * float64(initialGasPrice))

	feeModel = Model{
		InitialGasPrice: sdk.NewInt(initialGasPrice),
		MaxDiscount:     maxDiscount,
		OptimalBlockGas: 700,
		MaxBlockGas:     1000,
	}
)

func TestKeyPoints(t *testing.T) {
	// at point 0 it should be initial price

	nextGasPrice := calculateNextGasPrice(feeModel, 0, 100).Int64()
	assert.Equal(t, initialGasPrice, nextGasPrice)

	// if current average gas is equal to average gas it should be minDiscountedGasPrice

	nextGasPrice = calculateNextGasPrice(feeModel, 100, 100).Int64()
	assert.Equal(t, minDiscountedGasPrice, nextGasPrice)

	// if current average gas equals optimal block gas it still should be minDiscountedGasPrice

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.OptimalBlockGas, 100).Int64()
	assert.Equal(t, minDiscountedGasPrice, nextGasPrice)

	// if current average gas is equal to MaxBlockGas it should be a really big number

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.MaxBlockGas, 100).Int64()
	assert.Equal(t, 300*minDiscountedGasPrice, nextGasPrice)

	// for one point below MaxBlockGas it's also the same

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.MaxBlockGas-1, 100).Int64()
	assert.Equal(t, 300*minDiscountedGasPrice, nextGasPrice)

	// if current average gas is greater than MaxBlockGas it should stay the same

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.MaxBlockGas+100, 100).Int64()
	assert.Equal(t, 300*minDiscountedGasPrice, nextGasPrice)
}

func TestAverageGasBeyondOptimalBlockGas(t *testing.T) {
	// There is a special case when average block gas is higher than optimal block gas.
	// The question is if in such scenario we should offer discounted gas price or escalation should be applied instead.
	// It seems obvious that price should be escalated.

	nextGasPrice := calculateNextGasPrice(feeModel, feeModel.OptimalBlockGas+150, feeModel.OptimalBlockGas+100).Int64()
	assert.Greater(t, nextGasPrice, minDiscountedGasPrice)

	// Next gas price should be the same as for average block gas being below optimal block gas.
	// It means that escalation was turned on.

	nextGasPrice2 := calculateNextGasPrice(feeModel, feeModel.OptimalBlockGas+150, feeModel.OptimalBlockGas-100).Int64()
	assert.Equal(t, nextGasPrice, nextGasPrice2)
}

func TestShapeInDecreasingRegion(t *testing.T) {
	const averageBlockGas = 100

	lastPrice := initialGasPrice
	for i := int64(1); i <= averageBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, averageBlockGas).Int64()
		assert.Less(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}

func TestShapeInFlatRegion(t *testing.T) {
	const averageBlockGas = 100

	for i := int64(averageBlockGas); i <= feeModel.OptimalBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, averageBlockGas).Int64()
		assert.Equal(t, minDiscountedGasPrice, nextPrice)
	}
}

func TestShapeInEscalationRegion(t *testing.T) {
	const averageBlockGas = 100

	lastPrice := minDiscountedGasPrice
	for i := feeModel.OptimalBlockGas + 1; i < feeModel.MaxBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, averageBlockGas).Int64()
		assert.Greater(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}
