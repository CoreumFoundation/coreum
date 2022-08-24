//go:build simulation
// +build simulation

package feemodel

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	feeModelSim = Model{
		InitialGasPrice:         sdk.NewInt(1500),
		MaxGasPrice:             sdk.NewInt(15000),
		MaxDiscount:             0.5,
		EscalationStartBlockGas: 37500000, // 300 * BankSend message
		MaxBlockGas:             50000000, // 400 * BankSend message
		ShortAverageInertia:     10,
		LongAverageInertia:      1000,
	}
)

func ExampleGasPricePerBlockGas() {
	const longAverageBlockGas = 5000000

	for i := int64(0); i <= feeModelSim.MaxBlockGas+5000000; i += 10000 {
		fmt.Printf("%d\t%d\n", i, feeModelSim.CalculateNextGasPrice(i, longAverageBlockGas).Int64())
	}
	// Output: list of gas prices for each gas usage
	// FIXME (wojtek): refer picture containing a chart once doc is provided for the module
}

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
		shortAverage = calculateMovingAverage(shortAverage, gas, feeModelSim.ShortAverageInertia)
		longAverage = calculateMovingAverage(longAverage, gas, feeModelSim.LongAverageInertia)
		gasPrice := feeModelSim.CalculateNextGasPrice(shortAverage, longAverage)

		if i%10 != 0 {
			continue
		}

		fmt.Printf("%d\t%d\t%d\t%s\n", gas, shortAverage, longAverage, gasPrice)
	}

	// Output: list of gas prices over time
	// FIXME (wojtek): refer picture containing a chart once doc is provided for the module
}
