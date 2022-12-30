package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

// KeyMintFee represents the mint fee param key
var KeyMintFee = []byte("MintFee")

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintFee, &m.MintFee, validateMintFee),
	}
}

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		MintFee: sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
	}
}

// ValidateBasic validates parameters.
func (m Params) ValidateBasic() error {
	return validateMintFee(m.MintFee)
}

func validateMintFee(i interface{}) error {
	fee, ok := i.(sdk.Coin)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}
	if fee.IsNil() || !fee.IsValid() {
		return errors.New("mint fee must be a non-negative value")
	}
	return nil
}
