package fee

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Model stores parameters defining fee model of coreum blockchain
type Model struct {
	FeeDenom                             string
	InitialGasPrice                      sdk.Int
	MaxGasPrice                          sdk.Int
	MaxDiscount                          float64
	EscalationStartBlockGas              int64
	MaxBlockGas                          int64
	NumOfBlocksForCurrentAverageBlockGas uint
	NumOfBlocksForAverageBlockGas        uint
}

func calculateNextGasPrice(feeModel Model, currentAverageGas int64, averageGas int64) *big.Int {
	switch {
	case currentAverageGas >= feeModel.MaxBlockGas:
		return feeModel.MaxGasPrice.BigInt()
	case currentAverageGas > feeModel.EscalationStartBlockGas:
		maxDiscountedGasPrice := computeMaxDiscountedGasPrice(feeModel.InitialGasPrice.BigInt(), feeModel.MaxDiscount)

		// inertia defines how slow gas price goes up after triggering escalation algorithm (the lower the inertia,
		// the faster price goes up)
		const inertia = 2.
		height := new(big.Int).Sub(feeModel.MaxGasPrice.BigInt(), maxDiscountedGasPrice)
		width := float64(feeModel.MaxBlockGas - feeModel.EscalationStartBlockGas)
		x := float64(currentAverageGas - feeModel.EscalationStartBlockGas)

		escalationOffsetFloat := new(big.Float).SetInt(height)
		escalationOffsetFloat.Mul(escalationOffsetFloat, new(big.Float).SetFloat64(math.Pow(x/width, inertia)))
		escalationOffset, _ := escalationOffsetFloat.Int(nil)

		return maxDiscountedGasPrice.Add(maxDiscountedGasPrice, escalationOffset)
	case currentAverageGas >= averageGas:
		return computeMaxDiscountedGasPrice(feeModel.InitialGasPrice.BigInt(), feeModel.MaxDiscount)
	case averageGas > 0:
		discountFactor := math.Pow(1.-feeModel.MaxDiscount, float64(currentAverageGas)/float64(averageGas))

		gasPriceFloat := big.NewFloat(0).SetInt(feeModel.InitialGasPrice.BigInt())
		gasPriceFloat.Mul(gasPriceFloat, big.NewFloat(discountFactor))
		minGasPrice, _ := gasPriceFloat.Int(nil)

		return minGasPrice
	default:
		return feeModel.InitialGasPrice.BigInt()
	}
}

func computeMaxDiscountedGasPrice(initialGasPrice *big.Int, maxDiscount float64) *big.Int {
	maxDiscountedGasPriceFloat := big.NewFloat(0).SetInt(initialGasPrice)
	maxDiscountedGasPriceFloat.Mul(maxDiscountedGasPriceFloat, big.NewFloat(1.-maxDiscount))
	maxDiscountedGasPrice, _ := maxDiscountedGasPriceFloat.Int(nil)

	return maxDiscountedGasPrice
}
