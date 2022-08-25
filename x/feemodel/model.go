package feemodel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Model stores parameters defining fee model of coreum blockchain
type Model struct {
	FeeDenom                string
	InitialGasPrice         sdk.Int
	MaxGasPrice             sdk.Int
	MaxDiscount             sdk.Dec
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
		// be cautious: this function panics if shortEMA == m.EscalationStartBlockGas, that's why that case is not served here
		return m.calculateNextGasPriceInEscalationRegion(shortEMA)
	case shortEMA >= longEMA:
		return m.computeGasPriceWithMaxDiscount()
	case longEMA > 0:
		// be cautious: this function panics if longEMA == 0, that's why that case is not served here
		return m.calculateNextGasPriceInDiscountRegion(shortEMA, longEMA)
	default:
		return m.InitialGasPrice
	}
}

func (m Model) calculateNextGasPriceInEscalationRegion(shortEMA int64) sdk.Int {
	gasPriceWithMaxDiscount := m.computeGasPriceWithMaxDiscount()
	// exponent defines how slow gas price goes up after triggering escalation algorithm (the lower the exponent,
	// the faster price goes up)
	const exponent = 2
	height := m.MaxGasPrice.Sub(gasPriceWithMaxDiscount).ToDec()
	width := sdk.NewInt(m.MaxBlockGas - m.EscalationStartBlockGas).ToDec()
	x := sdk.NewInt(shortEMA - m.EscalationStartBlockGas).ToDec()

	offset := height.Mul(x.Quo(width).Power(exponent)).TruncateInt()
	return gasPriceWithMaxDiscount.Add(offset)
}

func (m Model) calculateNextGasPriceInDiscountRegion(shortEMA int64, longEMA int64) sdk.Int {
	gasPriceWithMaxDiscount := m.computeGasPriceWithMaxDiscount()
	// exponent defines how slow gas price goes up after triggering escalation algorithm (the lower the exponent,
	// the faster price goes up)
	const exponent = 2
	height := m.InitialGasPrice.Sub(gasPriceWithMaxDiscount).ToDec()
	width := sdk.NewInt(longEMA).ToDec()
	x := sdk.NewInt(shortEMA).ToDec()

	offset := height.Mul(x.Quo(width).Sub(sdk.OneDec()).Abs().Power(exponent)).TruncateInt()
	return gasPriceWithMaxDiscount.Add(offset)
}

func (m Model) computeGasPriceWithMaxDiscount() sdk.Int {
	return m.InitialGasPrice.ToDec().Mul(sdk.OneDec().Sub(m.MaxDiscount)).TruncateInt()
}

func calculateMovingAverage(previousAverage, newValue int64, numOfBlocks uint) int64 {
	return int64((uint64(numOfBlocks-1)*uint64(previousAverage) + uint64(newValue)) / uint64(numOfBlocks))
}
