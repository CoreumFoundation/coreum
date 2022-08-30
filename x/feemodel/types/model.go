package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// String implements the stringer interface.
func (m Model) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of model's parameters.
func (m *Model) ParamSetPairs() paramtypes.ParamSetPairs {
	modelValidator := func(value interface{}) error {
		return m.Validate()
	}

	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair([]byte("InitialGasPrice"), &m.InitialGasPrice, modelValidator),
		paramtypes.NewParamSetPair([]byte("MaxGasPrice"), &m.MaxGasPrice, modelValidator),
		paramtypes.NewParamSetPair([]byte("MaxDiscount"), &m.MaxDiscount, modelValidator),
		paramtypes.NewParamSetPair([]byte("EscalationStartBlockGas"), &m.EscalationStartBlockGas, modelValidator),
		paramtypes.NewParamSetPair([]byte("MaxBlockGas"), &m.MaxBlockGas, modelValidator),
		paramtypes.NewParamSetPair([]byte("ShortAverageBlockLength"), &m.ShortAverageBlockLength, modelValidator),
		paramtypes.NewParamSetPair([]byte("LongAverageBlockLength"), &m.LongAverageBlockLength, modelValidator),
	}
}

// DefaultModel returns model with default values
func DefaultModel() Model {
	return Model{
		// TODO: Find good parameters before lunching mainnet
		InitialGasPrice:         sdk.NewInt(1500),
		MaxGasPrice:             sdk.NewInt(1500000),
		MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
		EscalationStartBlockGas: 37500000, // 300 * BankSend message
		// TODO: adjust MaxBlockGas before creating testnet & mainnet
		MaxBlockGas:             50000000, // 400 * BankSend message
		ShortAverageBlockLength: 10,
		LongAverageBlockLength:  1000,
	}
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

// Validate validates parameters of the model
func (m Model) Validate() error {
	if m.InitialGasPrice.IsNil() {
		return errors.New("initial gas price is not set")
	}
	if m.MaxGasPrice.IsNil() {
		return errors.New("max gas price is not set")
	}
	if m.MaxDiscount.IsNil() {
		return errors.New("max discount is not set")
	}

	if m.InitialGasPrice.Sign() != 1 {
		return errors.New("initial gas price must be positive")
	}
	if m.MaxGasPrice.Sign() != 1 {
		return errors.New("max gas price must be positive")
	}
	if m.MaxGasPrice.LTE(m.InitialGasPrice) {
		return errors.New("max gas price must be greater than initial gas price")
	}
	if m.MaxDiscount.LTE(sdk.ZeroDec()) {
		return errors.New("max discount must be greater than 0")
	}
	if m.MaxDiscount.GTE(sdk.OneDec()) {
		return errors.New("max discount must be less than 1")
	}
	if m.EscalationStartBlockGas <= 0 {
		return errors.New("escalation start block gas must be greater than 0")
	}
	if m.MaxBlockGas <= m.EscalationStartBlockGas {
		return errors.New("max block gas must be greater than escalation start block gas")
	}
	if m.ShortAverageBlockLength == 0 {
		return errors.New("short average block length must be greater than 0")
	}
	if m.LongAverageBlockLength <= m.ShortAverageBlockLength {
		return errors.New("long average block length must be greater than short average block length")
	}

	return nil
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

// CalculateEMA calculates next EMA value
func CalculateEMA(previousAverage, newValue int64, numOfBlocks uint32) int64 {
	return int64((uint64(numOfBlocks-1)*uint64(previousAverage) + uint64(newValue)) / uint64(numOfBlocks))
}
