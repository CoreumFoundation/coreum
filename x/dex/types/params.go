package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyDefaultUnifiedRefAmount represents the default unified ref amount param key.
	KeyDefaultUnifiedRefAmount = []byte("DefaultUnifiedRefAmount")

	// KeyPriceTickExponent represents the price tick exponent param key.
	KeyPriceTickExponent = []byte("PriceTickExponent")
)

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		DefaultUnifiedRefAmount: sdkmath.LegacyMustNewDecFromStr("1000000"),
		PriceTickExponent:       -5,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(
			KeyDefaultUnifiedRefAmount,
			&m.DefaultUnifiedRefAmount,
			validateDefaultUnifiedRefAmount,
		),
		paramtypes.NewParamSetPair(
			KeyPriceTickExponent,
			&m.PriceTickExponent,
			validatePriceTickExponent,
		),
	}
}

// ValidateBasic validates parameters.
func (m Params) ValidateBasic() error {
	if err := validateDefaultUnifiedRefAmount(m.DefaultUnifiedRefAmount); err != nil {
		return err
	}
	return validatePriceTickExponent(m.PriceTickExponent)
}

func validateDefaultUnifiedRefAmount(i interface{}) error {
	amt, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if !amt.IsPositive() {
		return sdkerrors.Wrap(ErrInvalidInput, "default unified ref amount be a positive value")
	}
	return nil
}

func validatePriceTickExponent(i interface{}) error {
	exp, ok := i.(int32)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if exp >= 0 {
		return sdkerrors.Wrap(
			ErrInvalidInput,
			"price tick exponent must be negative",
		)
	}

	return nil
}
