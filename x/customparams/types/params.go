package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

var (
	// ParamStoreKeyMinSelfDelegation defines the param key for the min_self_delegation param.
	ParamStoreKeyMinSelfDelegation = []byte("minselfdelegation")
)

// StakingParamKeyTable returns the parameter key table.
func StakingParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&StakingParams{})
}

// DefaultStakingParams returns default staking parameters.
func DefaultStakingParams() StakingParams {
	return StakingParams{
		MinSelfDelegation: sdk.OneInt(),
	}
}

// ParamSetPairs returns the parameter set pairs.
func (p *StakingParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMinSelfDelegation, &p.MinSelfDelegation, validateMinSelfDelegation),
	}
}

// ValidateBasic performs basic validation on staking parameters.
func (p StakingParams) ValidateBasic() error {
	return validateMinSelfDelegation(p.MinSelfDelegation)
}

func validateMinSelfDelegation(i interface{}) error {
	v, ok := i.(sdk.Int)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return errors.New("param min_self_delegation must be not nil")
	}
	if v.IsNegative() {
		return errors.Errorf("param min_self_delegation must be positive: %s", v)
	}

	return nil
}
