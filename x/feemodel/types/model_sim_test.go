//go:build simulation
// +build simulation

package types

import (
	"fmt"
)

var feeModelSim = DefaultModel()

//nolint:govet // This example does not refer to any identifier
func ExampleGasPricePerBlockGas() {
	const longEMABlockGas = 5000000

	for i := int64(0); i <= feeModelSim.Params().MaxBlockGas+5000000; i += 10000 {
		fmt.Printf("%d\t%s\n", i, feeModelSim.CalculateNextGasPrice(i, longEMABlockGas))
	}
	// Output: list of gas prices for each gas usage
	// Check x/feemodel/spec/assets/curve.png
}

//nolint:govet // This example does not refer to any identifier
func ExampleGasPriceOverTime() {
	var (
		blockGas []int64
		shortEMA int64
		longEMA  int64
		params   = feeModelSim.Params()
	)
	for i := int64(0.4 * float64(params.EscalationStartBlockGas)); i <= params.MaxBlockGas; i += 1000 {
		blockGas = append(blockGas, i)
	}
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, params.MaxBlockGas)
	}
	gas := int64(0.7 * float64(params.EscalationStartBlockGas))
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.5 * float64(params.EscalationStartBlockGas))
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.2 * float64(params.EscalationStartBlockGas))
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, gas)
	}
	gas = int64(0.9 * float64(params.EscalationStartBlockGas))
	for i := 0; i < 5000; i++ {
		blockGas = append(blockGas, gas)
	}
	for i := 0; i < 3000; i++ {
		blockGas = append(blockGas, 0)
	}

	for i, gas := range blockGas {
		shortEMA = CalculateEMA(shortEMA, gas, params.ShortEmaBlockLength)
		longEMA = CalculateEMA(longEMA, gas, params.LongEmaBlockLength)
		gasPrice := feeModelSim.CalculateNextGasPrice(shortEMA, longEMA)

		if i%10 != 0 {
			continue
		}

		fmt.Printf("%d\t%d\t%d\t%d\t%s\n", i, gas, shortEMA, longEMA, gasPrice)
	}

	// Output: list of gas prices over time
	// Check x/feemodel/spec/assets/time_series.png
}
