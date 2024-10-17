package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// KeyDefaultUnifiedRefAmount represents the default unified ref amount param key.
	KeyDefaultUnifiedRefAmount = []byte("DefaultUnifiedRefAmount")

	// KeyPriceTickExponent represents the price tick exponent param key.
	KeyPriceTickExponent = []byte("PriceTickExponent")

	// KeyMaxOrdersPerDenom represents the max orders per denom param key.
	KeyMaxOrdersPerDenom = []byte("MaxOrdersPerDenom")

	// KeyOrderReserve represents the order reserve param key.
	KeyOrderReserve = []byte("OrderReserve")
)

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		DefaultUnifiedRefAmount: sdkmath.LegacyMustNewDecFromStr("1000000"),
		PriceTickExponent:       -8,
		MaxOrdersPerDenom:       100,
		OrderReserve:            sdk.NewInt64Coin(sdk.DefaultBondDenom, 10_000_000),
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
		paramtypes.NewParamSetPair(
			KeyMaxOrdersPerDenom,
			&m.MaxOrdersPerDenom,
			validateMaxOrdersPerDenom,
		),
		paramtypes.NewParamSetPair(
			KeyOrderReserve,
			&m.OrderReserve,
			validateOrderReserve,
		),
	}
}

// ValidateBasic validates parameters.
func (m Params) ValidateBasic() error {
	if err := validateDefaultUnifiedRefAmount(m.DefaultUnifiedRefAmount); err != nil {
		return err
	}

	if err := validatePriceTickExponent(m.PriceTickExponent); err != nil {
		return err
	}

	if err := validateMaxOrdersPerDenom(m.MaxOrdersPerDenom); err != nil {
		return err
	}

	return validateOrderReserve(m.OrderReserve)
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

func validateMaxOrdersPerDenom(i interface{}) error {
	maxOrders, ok := i.(uint64)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if maxOrders == 0 {
		return sdkerrors.Wrap(
			ErrInvalidInput,
			"max orders per denom must be positive",
		)
	}

	return nil
}

func validateOrderReserve(i interface{}) error {
	orderReserve, ok := i.(sdk.Coin)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if !orderReserve.IsValid() {
		return sdkerrors.Wrap(
			ErrInvalidInput,
			"invalid order reserve",
		)
	}

	return nil
}
