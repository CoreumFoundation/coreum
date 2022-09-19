package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// DefaultModel returns model with default params
func DefaultModel() Model {
	return Model{
		params: DefaultParams(),
	}
}

// NewModel creates model
func NewModel(params Params) Model {
	return Model{
		params: params,
	}
}

// Model executes fee model
type Model struct {
	params Params
}

// Params returns fee model params
func (m Model) Params() Params {
	return m.params
}

// CalculateNextGasPrice calculates minimum gas price for next block
// Chart showing a sample output of the fee model: x/feemodel/spec/assets/curve.png
func (m Model) CalculateNextGasPrice(shortEMA int64, longEMA int64) sdk.Dec {
	switch {
	case shortEMA >= m.params.MaxBlockGas:
		return m.params.MaxGasPrice
	case shortEMA > m.params.EscalationStartBlockGas:
		// be cautious: this function panics if shortEMA == m.EscalationStartBlockGas, that's why that case is not served here
		return m.calculateNextGasPriceInEscalationRegion(shortEMA)
	case shortEMA >= longEMA:
		return m.CalculateGasPriceWithMaxDiscount()
	case longEMA > 0:
		// be cautious: this function panics if longEMA == 0, that's why that case is not served here
		return m.calculateNextGasPriceInDiscountRegion(shortEMA, longEMA)
	default:
		return m.params.InitialGasPrice
	}
}

// CalculateGasPriceWithMaxDiscount calculates gas price with maximum discount applied
func (m Model) CalculateGasPriceWithMaxDiscount() sdk.Dec {
	return m.params.InitialGasPrice.Mul(sdk.OneDec().Sub(m.params.MaxDiscount))
}

func (m Model) calculateNextGasPriceInEscalationRegion(shortEMA int64) sdk.Dec {
	gasPriceWithMaxDiscount := m.CalculateGasPriceWithMaxDiscount()
	// exponent defines how slow gas price goes up after triggering escalation algorithm (the lower the exponent,
	// the faster price goes up)
	const exponent = 2
	height := m.params.MaxGasPrice.Sub(gasPriceWithMaxDiscount)
	width := sdk.NewInt(m.params.MaxBlockGas - m.params.EscalationStartBlockGas).ToDec()
	x := sdk.NewInt(shortEMA - m.params.EscalationStartBlockGas).ToDec()

	offset := height.Mul(x.Quo(width).Power(exponent))
	return gasPriceWithMaxDiscount.Add(offset)
}

func (m Model) calculateNextGasPriceInDiscountRegion(shortEMA int64, longEMA int64) sdk.Dec {
	gasPriceWithMaxDiscount := m.CalculateGasPriceWithMaxDiscount()
	// exponent defines how slow gas price goes up after triggering escalation algorithm (the lower the exponent,
	// the faster price goes up)
	const exponent = 2
	height := m.params.InitialGasPrice.Sub(gasPriceWithMaxDiscount)
	width := sdk.NewInt(longEMA).ToDec()
	x := sdk.NewInt(shortEMA).ToDec()

	offset := height.Mul(x.Quo(width).Sub(sdk.OneDec()).Abs().Power(exponent))
	return gasPriceWithMaxDiscount.Add(offset)
}

// CalculateEMA calculates next EMA value
func CalculateEMA(previousEMA, newValue int64, numOfBlocks uint32) int64 {
	return int64((uint64(numOfBlocks-1)*uint64(previousEMA) + uint64(newValue)) / uint64(numOfBlocks))
}
