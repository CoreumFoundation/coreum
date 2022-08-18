//go:build simulation
// +build simulation

package fee

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	feeModelSim = Model{
		InitialGasPrice:                    sdk.NewInt(1500),
		MaxGasPrice:                        sdk.NewInt(15000),
		MaxDiscount:                        0.5,
		EscalationStartBlockGas:            37500000, // 300 * BankSend message
		MaxBlockGas:                        50000000, // 400 * BankSend message
		NumOfBlocksForShortAverageBlockGas: 10,
		NumOfBlocksForLongAverageBlockGas:  1000,
	}
)

func ExampleGasPricePerBlockGas() {
	const longAverageBlockGas = 5000000

	for i := int64(0); i <= feeModelSim.MaxBlockGas+5000000; i += 10000 {
		fmt.Println(calculateNextGasPrice(feeModelSim, i, longAverageBlockGas).Int64())
	}
	// Output: list of gas prices for each gas usage
	// https://docs.google.com/spreadsheets/d/1YTvt06CIgHpx5kgOXk2BK-kuJ63DwVYtGfDEHLxvCZQ/edit#gid=0
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
		shortAverage = calculateMovingAverage(shortAverage, gas, feeModelSim.NumOfBlocksForShortAverageBlockGas)
		longAverage = calculateMovingAverage(longAverage, gas, feeModelSim.NumOfBlocksForLongAverageBlockGas)
		gasPrice := calculateNextGasPrice(feeModelSim, shortAverage, longAverage)

		if i%10 != 0 {
			continue
		}

		fmt.Printf("%d\t%d\t%d\t%s\n", gas, shortAverage, longAverage, gasPrice)
	}

	// Output: list of gas prices over time
	// https://docs.google.com/spreadsheets/d/1YTvt06CIgHpx5kgOXk2BK-kuJ63DwVYtGfDEHLxvCZQ/edit#gid=940400407
}
