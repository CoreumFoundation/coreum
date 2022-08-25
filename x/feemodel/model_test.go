package feemodel

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

var (
	feeModel = Model{
		InitialGasPrice:         sdk.NewInt(500000000000000),
		MaxGasPrice:             sdk.NewInt(1000000000000000),
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
		assert.True(t, nextPrice.LT(lastPrice))

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
