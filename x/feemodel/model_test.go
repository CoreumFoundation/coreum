package feemodel

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	initialGasPrice         int64 = 100000000
	maxGasPrice             int64 = 200000000
	maxDiscount                   = 0.5
	gasPriceWithMaxDiscount       = int64((1. - maxDiscount) * float64(initialGasPrice))

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

	nextGasPrice := feeModel.CalculateNextGasPrice(0, 100).Int64()
	assert.Equal(t, initialGasPrice, nextGasPrice)

	// if current average gas is equal to average gas it should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(100, 100).Int64()
	assert.Equal(t, gasPriceWithMaxDiscount, nextGasPrice)

	// if current average gas equals escalation start block gas it still should be gasPriceWithMaxDiscount

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas, 100).Int64()
	assert.Equal(t, gasPriceWithMaxDiscount, nextGasPrice)

	// if current average gas is equal to MaxBlockGas it should be max gas price number

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.MaxBlockGas, 100).Int64()
	assert.Equal(t, maxGasPrice, nextGasPrice)

	// if current average gas is greater than MaxBlockGas it should stay the same

	nextGasPrice = feeModel.CalculateNextGasPrice(feeModel.MaxBlockGas+100, 100).Int64()
	assert.Equal(t, maxGasPrice, nextGasPrice)
}

func TestAverageGasBeyondEscalationStartBlockGas(t *testing.T) {
	// There is a special case when long average block gas is higher than escalation start block gas.
	// The question is if in such scenario we should offer discounted gas price or escalation should be applied instead.
	// It seems obvious that price should be escalated.

	nextGasPrice := feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas+100).Int64()
	assert.Greater(t, nextGasPrice, gasPriceWithMaxDiscount)

	// Next gas price should be the same as for long average block gas being below optimal block gas.
	// It means that escalation was turned on.

	nextGasPrice2 := feeModel.CalculateNextGasPrice(feeModel.EscalationStartBlockGas+150, feeModel.EscalationStartBlockGas-100).Int64()
	assert.Equal(t, nextGasPrice, nextGasPrice2)
}

func TestZeroAverageGas(t *testing.T) {
	nextGasPrice := feeModel.CalculateNextGasPrice(0, 0).Int64()
	assert.Equal(t, nextGasPrice, gasPriceWithMaxDiscount)

	nextGasPrice = feeModel.CalculateNextGasPrice(1, 0).Int64()
	assert.Equal(t, nextGasPrice, gasPriceWithMaxDiscount)
}

func TestShapeInDecreasingRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := initialGasPrice
	for i := int64(1); i <= longAverageBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas).Int64()
		assert.Less(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}

func TestShapeInFlatRegion(t *testing.T) {
	const longAverageBlockGas = 100

	for i := int64(longAverageBlockGas); i <= feeModel.EscalationStartBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas).Int64()
		assert.Equal(t, gasPriceWithMaxDiscount, nextPrice)
	}
}

func TestShapeInEscalationRegion(t *testing.T) {
	const longAverageBlockGas = 100

	lastPrice := gasPriceWithMaxDiscount
	for i := feeModel.EscalationStartBlockGas + 1; i <= feeModel.MaxBlockGas; i++ {
		nextPrice := feeModel.CalculateNextGasPrice(i, longAverageBlockGas).Int64()
		assert.Greater(t, nextPrice, lastPrice)

		lastPrice = nextPrice
	}
}
