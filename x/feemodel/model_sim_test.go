//go:build simulation
// +build simulation

package feemodel

import (
	"fmt"
)

var feeModelSim = DefaultModel()

//nolint:govet // This example does not refer to any identifier
func ExampleGasPricePerBlockGas() {
	const longAverageBlockGas = 5000000

	for i := int64(0); i <= feeModelSim.MaxBlockGas+5000000; i += 10000 {
		fmt.Printf("%d\t%s\n", i, feeModelSim.CalculateNextGasPrice(i, longAverageBlockGas))
	}
	// Output: list of gas prices for each gas usage
	// FIXME (wojtek): refer picture containing a chart once doc is provided for the module
}

//nolint:govet // This example does not refer to any identifier
func ExampleGasPriceOverTime() {
	var (
		blockGas     []int64
		shortAverage int64 = 0
		longAverage  int64 = 0
	)
	for i := int64(0.4 * float64(feeModelSim.EscalationStartBlockGas)); i <= feeModelSim.MaxBlockGas; i += 1000 {
		blockGas = append(blockGas, i)
	}
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, feeModelSim.MaxBlockGas)
	}
	gas := int64(0.7 * float64(feeModelSim.EscalationStartBlockGas))
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.5 * float64(feeModelSim.EscalationStartBlockGas))
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.2 * float64(feeModelSim.EscalationStartBlockGas))
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.9 * float64(feeModelSim.EscalationStartBlockGas))
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, gas)
	}
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, 0)
	}

	for i, gas := range blockGas {
		shortAverage = calculateMovingAverage(shortAverage, gas, feeModelSim.ShortAverageBlockLength)
		longAverage = calculateMovingAverage(longAverage, gas, feeModelSim.LongAverageBlockLength)
		gasPrice := feeModelSim.CalculateNextGasPrice(shortAverage, longAverage)

		if i%10 != 0 {
			continue
		}

		fmt.Printf("%d\t%d\t%d\t%d\t%s\n", i, gas, shortAverage, longAverage, gasPrice)
	}

	// Output: list of gas prices over time
	// FIXME (wojtek): refer picture containing a chart once doc is provided for the module
}
