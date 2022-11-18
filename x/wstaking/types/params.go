package types

import (
	"fmt"
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	// ParamStoreKeyMinSelfDelegation defines the param key for the min_self_delegation param.
	ParamStoreKeyMinSelfDelegation = []byte("minselfdelegation")
)

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		MinSelfDelegation: sdk.OneInt(),
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMinSelfDelegation, &p.MinSelfDelegation, validateMinSelfDelegation),
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if err := validateMinSelfDelegation(p.MinSelfDelegation); err != nil {
		return err
	}

	return nil
}

func validateMinSelfDelegation(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return errors.New(fmt.Sprintf("invalid parameter type: %T", i))
	}

	if v.IsNil() {
		return errors.New("param min_self_delegation tax must be not nil")
	}
	if v.IsNegative() {
		return errors.New(fmt.Sprintf("param min_self_delegation must be positive: %s", v))
	}

	return nil
}
