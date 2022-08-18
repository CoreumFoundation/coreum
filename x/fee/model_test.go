package fee

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	initialGasPrice       int64 = 100000000
	maxGasPrice           int64 = 200000000
	maxDiscount                 = 0.5
	minDiscountedGasPrice       = int64((1. - maxDiscount) * float64(initialGasPrice))

	feeModel = Model{
		InitialGasPrice:         sdk.NewInt(initialGasPrice),
		MaxGasPrice:             sdk.NewInt(maxGasPrice),
		MaxDiscount:             maxDiscount,
		EscalationStartBlockGas: 700,
		MaxBlockGas:             1000,
	}
)

func TestCalculateNextGasPriceKeyPoints(t *testing.T) {
	// at point 0 it should be initial price

	nextGasPrice := calculateNextGasPrice(feeModel, 0, 100).Int64()
	assert.Equal(t, initialGasPrice, nextGasPrice)

	// if current average gas is equal to average gas it should be minDiscountedGasPrice

	nextGasPrice = calculateNextGasPrice(feeModel, 100, 100).Int64()
	assert.Equal(t, minDiscountedGasPrice, nextGasPrice)

	// if current average gas equals escalation start block gas it still should be minDiscountedGasPrice

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.EscalationStartBlockGas, 100).Int64()
	assert.Equal(t, minDiscountedGasPrice, nextGasPrice)

	// if current average gas is equal to MaxBlockGas it should be max gas price number

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.MaxBlockGas, 100).Int64()
	assert.Equal(t, maxGasPrice, nextGasPrice)

	// if current average gas is greater than MaxBlockGas it should stay the same

	nextGasPrice = calculateNextGasPrice(feeModel, feeModel.MaxBlockGas+100, 100).Int64()
	assert.Equal(t, maxGasPrice, nextGasPrice)
}

func TestAverageGasBeyondEscalationStartBlockGas(t *testing.T) {
	// There is a special case when long average block gas is higher than escalation start block gas.
	// The question is if in such scenario we should offer discounted gas price or escalation should be applied instead.
	// It seems obvious that price should be escalated.

	nextGasPrice := calculateNextGasPrice(feeModel, feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas+100).Int64()
	assert.Greater(t, nextGasPrice, minDiscountedGasPrice)

	// Next gas price should be the same as for long average block gas being below optimal block gas.
	// It means that escalation was turned on.

	nextGasPrice2 := calculateNextGasPrice(feeModel, feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas-100).Int64()
	assert.Equal(t, nextGasPrice, nextGasPrice2)
}

func TestZeroAverageGas(t *testing.T) {
	nextGasPrice := calculateNextGasPrice(feeModel, 0, 0).Int64()
	assert.Equal(t, nextGasPrice, minDiscountedGasPrice)

	nextGasPrice = calculateNextGasPrice(feeModel, 1, 0).Int64()
	assert.Equal(t, nextGasPrice, minDiscountedGasPrice)
}

func TestShapeInDecreasingRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := initialGasPrice
	for i := int64(1); i <= longAverageBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, longAverageBlockGas).Int64()
		assert.Less(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}

func TestShapeInFlatRegion(t *testing.T) {
	const longAverageBlockGas = 100

	for i := int64(longAverageBlockGas); i <= feeModel.EscalationStartBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, longAverageBlockGas).Int64()
		assert.Equal(t, minDiscountedGasPrice, nextPrice)
	}
}

func TestShapeInEscalationRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := minDiscountedGasPrice
	for i := feeModel.EscalationStartBlockGas + 1; i <= feeModel.MaxBlockGas; i++ {
		nextPrice := calculateNextGasPrice(feeModel, i, longAverageBlockGas).Int64()
		assert.Greater(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}
