package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

// KeyIssueFee represents the issue fee param key
var KeyIssueFee = []byte("IssueFee")

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIssueFee, &m.IssueFee, validateIssueFee),
	}
}

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		IssueFee: sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
	}
}

// ValidateBasic validates parameters.
func (m Params) ValidateBasic() error {
	return validateIssueFee(m.IssueFee)
}

func validateIssueFee(i interface{}) error {
	fee, ok := i.(sdk.Coin)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}
	if fee.IsNil() || !fee.IsValid() {
		return errors.New("issue fee must be a non-negative value")
	}
	return nil
}
