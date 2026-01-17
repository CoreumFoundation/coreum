package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types" // <--- 1. ADD THIS IMPORT
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

// 2. ADD THE NEW KEY HERE
var (
	ParamStoreKeyMinSelfDelegation = []byte("minselfdelegation")
	ParamStoreKeyMinCommissionRate = []byte("mincommissionrate")
)

// StakingParamKeyTable returns the parameter key table.
func StakingParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&StakingParams{})
}

// 3. UPDATE DEFAULT PARAMS HERE
func DefaultStakingParams() StakingParams {
	return StakingParams{
		MinSelfDelegation: sdkmath.OneInt(),
		MinCommissionRate: sdk.NewDecWithPrec(5, 2), // Sets default to 5%
	}
}

// 4. REGISTER THE NEW PARAMETER HERE
func (p *StakingParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyMinSelfDelegation, &p.MinSelfDelegation, validateMinSelfDelegation),
		paramtypes.NewParamSetPair(ParamStoreKeyMinCommissionRate, &p.MinCommissionRate, validateMinCommissionRate),
	}
}

// ValidateBasic performs basic validation on staking parameters.
func (p StakingParams) ValidateBasic() error {
	if err := validateMinSelfDelegation(p.MinSelfDelegation); err != nil {
		return err
	}
	return validateMinCommissionRate(p.MinCommissionRate)
}

func validateMinSelfDelegation(i interface{}) error {
	v, ok := i.(sdkmath.Int)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return errors.New("param min_self_delegation must be not nil")
	}
	if !v.IsPositive() {
		return errors.Errorf("param min_self_delegation must be positive: %s", v)
	}

	return nil
}

// validateMinCommissionRate enforces the 5% floor
func validateMinCommissionRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return errors.New("param min_commission_rate must be not nil")
	}

	// Define the 5% floor (0.05)
	minAllowed := sdk.NewDecWithPrec(5, 2)

	if v.LT(minAllowed) {
		return errors.Errorf("param min_commission_rate cannot be lower than %s (5%%)", minAllowed)
	}

	if v.GT(sdk.OneDec()) {
		return errors.New("param min_commission_rate cannot be greater than 1 (100%)")
	}

	return nil
}
