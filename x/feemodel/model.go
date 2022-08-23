package feemodel

import (
	"math"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Model stores parameters defining fee model of coreum blockchain
type Model struct {
	FeeDenom                string
	InitialGasPrice         sdk.Int
	MaxGasPrice             sdk.Int
	MaxDiscount             float64
	EscalationStartBlockGas int64
	MaxBlockGas             int64
	ShortAverageInertia     uint
	LongAverageInertia      uint
}

// CalculateNextGasPrice calculates minimum gas price for next block
// Chart showing a sample output of the fee mdoel: https://docs.google.com/spreadsheets/d/1YTvt06CIgHpx5kgOXk2BK-kuJ63DwVYtGfDEHLxvCZQ/edit#gid=0
func (m Model) CalculateNextGasPrice(shortEMA int64, longEMA int64) sdk.Int {
	switch {
	case shortEMA >= m.MaxBlockGas:
		return m.MaxGasPrice
	case shortEMA > m.EscalationStartBlockGas:
		return m.calculateNextGasPriceInEscalationRegion(shortEMA)
	case shortEMA >= longEMA:
		return m.computeMGasPriceWithMaxDiscount()
	case longEMA > 0:
		return m.calculateNextGasPriceInDiscountRegion(shortEMA, longEMA)
	default:
		return m.InitialGasPrice
	}
}

func (m Model) calculateNextGasPriceInEscalationRegion(shortEMA int64) sdk.Int {
	gasPriceWithMaxDiscount := m.computeMGasPriceWithMaxDiscount()

	// inertia defines how slow gas price goes up after triggering escalation algorithm (the lower the inertia,
	// the faster price goes up)
	const inertia = 2.0
	height := m.MaxGasPrice.Sub(gasPriceWithMaxDiscount)
	width := float64(m.MaxBlockGas - m.EscalationStartBlockGas)
	x := float64(shortEMA - m.EscalationStartBlockGas)

	escalationOffsetFloat := new(big.Float).SetInt(height.BigInt())
	escalationOffsetFloat.Mul(escalationOffsetFloat, big.NewFloat(math.Pow(x/width, inertia)))
	escalationOffset, _ := escalationOffsetFloat.Int(nil)

	return gasPriceWithMaxDiscount.Add(sdk.NewIntFromBigInt(escalationOffset))
}

func (m Model) calculateNextGasPriceInDiscountRegion(shortEMA int64, longEMA int64) sdk.Int {
	discountFactor := math.Pow(1.-m.MaxDiscount, float64(shortEMA)/float64(longEMA))

	gasPriceFloat := big.NewFloat(0).SetInt(m.InitialGasPrice.BigInt())
	gasPriceFloat.Mul(gasPriceFloat, big.NewFloat(discountFactor))
	minGasPrice, _ := gasPriceFloat.Int(nil)

	return sdk.NewIntFromBigInt(minGasPrice)
}

func (m Model) computeMGasPriceWithMaxDiscount() sdk.Int {
	gasPriceWithMaxDiscountFloat := big.NewFloat(0).SetInt(m.InitialGasPrice.BigInt())
	gasPriceWithMaxDiscountFloat.Mul(gasPriceWithMaxDiscountFloat, big.NewFloat(1.-m.MaxDiscount))
	gasPriceWithMaxDiscount, _ := gasPriceWithMaxDiscountFloat.Int(nil)
	return sdk.NewIntFromBigInt(gasPriceWithMaxDiscount)
}

func calculateMovingAverage(previousAverage, newValue int64, numOfBlocks uint) int64 {
	return int64((uint64(numOfBlocks-1)*uint64(previousAverage) + uint64(newValue)) / uint64(numOfBlocks))
}
