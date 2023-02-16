package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

// KeyModel represents the Model param key with which the ModelParams will be stored.
var KeyModel = []byte("Model")

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of model's parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyModel, &m.Model, validateModelParams),
	}
}

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		Model: ModelParams{
			// TODO: Find good parameters before launching mainnet
			InitialGasPrice:         sdk.MustNewDecFromStr("0.0625"),
			MaxGasPriceMultiplier:   sdk.MustNewDecFromStr("1000.0"),
			MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
			EscalationStartFraction: sdk.MustNewDecFromStr("0.8"),
			// TODO: adjust MaxBlockGas before creating testnet & mainnet
			MaxBlockGas:         50000000, // 400 * BankSend message
			ShortEmaBlockLength: 50,
			LongEmaBlockLength:  1000,
		},
	}
}

// ValidateBasic validates parameters of the model.
func (m Params) ValidateBasic() error {
	return validateModelParams(m.Model)
}

// ValidateBasic validates parameters of the model params.
func (m ModelParams) ValidateBasic() error {
	return validateModelParams(m)
}

func validateModelParams(i interface{}) error {
	m, ok := i.(ModelParams)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}

	if m.InitialGasPrice.IsNil() {
		return errors.New("initial gas price is not set")
	}
	if m.MaxGasPriceMultiplier.IsNil() {
		return errors.New("max gas price multiplier is not set")
	}
	if m.MaxDiscount.IsNil() {
		return errors.New("max discount is not set")
	}

	if !m.InitialGasPrice.IsPositive() {
		return errors.New("initial gas price must be positive")
	}
	if m.MaxGasPriceMultiplier.LTE(sdk.OneDec()) {
		return errors.New("max gas price multiplier must be greater than one")
	}
	if m.MaxDiscount.LTE(sdk.ZeroDec()) {
		return errors.New("max discount must be greater than 0")
	}
	if m.MaxDiscount.GTE(sdk.OneDec()) {
		return errors.New("max discount must be less than 1")
	}

	if m.EscalationStartFraction.IsNil() {
		return errors.New("escalation start fraction is not set")
	}
	if m.EscalationStartFraction.LTE(sdk.ZeroDec()) {
		return errors.New("escalation start fraction must be greater than 0")
	}
	if m.EscalationStartFraction.GTE(sdk.OneDec()) {
		return errors.New("escalation start fraction must be less than 1")
	}
	if m.ShortEmaBlockLength == 0 {
		return errors.New("short EMA block length must be greater than 0")
	}
	if m.LongEmaBlockLength <= m.ShortEmaBlockLength {
		return errors.New("long EMA block length must be greater than short EMA block length")
	}

	if m.MaxBlockGas < 1 {
		return errors.New("max block gas must be bigger than 0")
	}

	return nil
}
