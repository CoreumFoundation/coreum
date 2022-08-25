package feemodel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Model stores parameters defining fee model of coreum blockchain
// There are four regions on the fee model curve
// - between 0 and "long average block gas" where gas price goes down exponentially from InitialGasPrice to gas price with maximum discount (InitialGasPrice * (1 -MaxDiscount))
// - between "long average block gas" and EscalationStartBlockGas where we offer gas price with maximum discount all the time
// - between EscalationStartBlockGas and MaxBlockGas where price goes up rapidly (being an output of a power function) from gas price with maximum discount to MaxGasPrice
// - above MaxBlockGas (if it happens for any reason) where price is equal to MaxGasPrice
//
// The input (x value) for that function is calculated by taking short block gas average.
// Price (y value) being an output of the fee model is used as the minimum gas price for next block.
type Model struct {
	// InitialGasPrice is used when block gas short average is 0. It happens when there are no transactions being broadcasted. This value is also used to initialize gas price on brand-new chain.
	InitialGasPrice sdk.Int

	// MaxGasPrice is used when block gas short average is greater than or equal to MaxBlockGas. This value is used to limit gas price escalation to avoid having possible infinity GasPrice value otherwise.
	MaxGasPrice sdk.Int

	// MaxDiscount is th maximum discount we offer on top of initial gas price if short average block gas is between long average block gas and escalation start block gas.
	MaxDiscount sdk.Dec

	// EscalationStartBlockGas defines block gas usage where gas price escalation starts if short average block gas is higher than this value.
	EscalationStartBlockGas int64

	// MaxBlockGas sets the maximum capacity of block. This is enforced on tendermint level in genesis configuration. Once short average block gas goes above this value, gas price is a flat line equal to MaxGasPrice.
	MaxBlockGas int64

	// ShortAverageBlockLength defines inertia for short average long gas in EMA model. The equation is: NewAverage = ((ShortAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / ShortAverageBlockLength
	// The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.
	ShortAverageBlockLength uint

	// LongAverageBlockLength defines inertia for long average block gas in EMA model. The equation is: NewAverage = ((LongAverageBlockLength - 1)*PreviousAverage + GasUsedByCurrentBlock) / LongAverageBlockLength
	// The value might be interpreted as the number of blocks which are taken to calculate the average. It would be exactly like that in SMA model, in EMA this is an approximation.
	LongAverageBlockLength uint
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
