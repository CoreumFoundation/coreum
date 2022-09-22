package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var (
	KeyModel = []byte("Model")
)

// String implements the stringer interface.
func (m Params) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

// String implements the stringer interface.
func (m ModelParams) String() string {
	out, _ := yaml.Marshal(m)
	return string(out)
}

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
			// TODO: Find good parameters before lunching mainnet
			InitialGasPrice:         sdk.NewDec(1500),
			MaxGasPrice:             sdk.NewDec(1500000),
			MaxDiscount:             sdk.MustNewDecFromStr("0.5"),
			EscalationStartBlockGas: 37500000, // 300 * BankSend message
			// TODO: adjust MaxBlockGas before creating testnet & mainnet
			MaxBlockGas:         50000000, // 400 * BankSend message
			ShortEmaBlockLength: 10,
			LongEmaBlockLength:  1000,
		},
	}
}

// ValidateBasic validates parameters of the model
func (m Params) ValidateBasic() error {
	return validateModelParams(m.Model)
}

func validateModelParams(i interface{}) error {
	m, ok := i.(ModelParams)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}

	if m.InitialGasPrice.IsNil() {
		return errors.New("initial gas price is not set")
	}
	if m.MaxGasPrice.IsNil() {
		return errors.New("max gas price is not set")
	}
	if m.MaxDiscount.IsNil() {
		return errors.New("max discount is not set")
	}

	if !m.InitialGasPrice.IsPositive() {
		return errors.New("initial gas price must be positive")
	}
	if !m.MaxGasPrice.IsPositive() {
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
	if m.ShortEmaBlockLength == 0 {
		return errors.New("short EMA block length must be greater than 0")
	}
	if m.LongEmaBlockLength <= m.ShortEmaBlockLength {
		return errors.New("long EMA block length must be greater than short EMA block length")
	}

	return nil
}
