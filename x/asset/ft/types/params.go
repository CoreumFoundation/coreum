package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/pkg/errors"
)

const (
	// DefaultTokenUpgradeDecisionPeriod is the period when issuer must decide if token should be upgraded.
	DefaultTokenUpgradeDecisionPeriod = time.Hour * 24 * 30

	// DefaultTokenUpgradeGracePeriod is the period after which upgrade is effectively executed.
	DefaultTokenUpgradeGracePeriod = time.Hour * 24 * 7
)

var (
	// KeyIssueFee represents the issue fee param key.
	KeyIssueFee = []byte("IssueFee")

	// KeyTokenUpgradeDecisionTimeout represents the token upgrade decision timeout param key.
	KeyTokenUpgradeDecisionTimeout = []byte("TokenUpgradeDecisionTimeout")

	// KeyTokenUpgradeGracePeriod represents the token upgrade grace period param key.
	KeyTokenUpgradeGracePeriod = []byte("TokenUpgradeGracePeriod")
)

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *ParamsV1) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIssueFee, &m.IssueFee, validateIssueFee),
	}
}

// ValidateBasic validates parameters.
func (m ParamsV1) ValidateBasic() error {
	return validateIssueFee(m.IssueFee)
}

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		IssueFee:                    sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
		TokenUpgradeDecisionTimeout: time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC),
		TokenUpgradeGracePeriod:     DefaultTokenUpgradeGracePeriod,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIssueFee, &m.IssueFee, validateIssueFee),
		paramtypes.NewParamSetPair(KeyTokenUpgradeDecisionTimeout, &m.TokenUpgradeDecisionTimeout, validateTokenUpgradeDecisionTimeout),
		paramtypes.NewParamSetPair(KeyTokenUpgradeGracePeriod, &m.TokenUpgradeGracePeriod, validateTokenUpgradeGracePeriod),
	}
}

// ValidateBasic validates parameters.
func (m Params) ValidateBasic() error {
	if err := validateIssueFee(m.IssueFee); err != nil {
		return err
	}
	if err := validateTokenUpgradeDecisionTimeout(m.TokenUpgradeDecisionTimeout); err != nil {
		return err
	}
	return validateTokenUpgradeGracePeriod(m.TokenUpgradeGracePeriod)
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

func validateTokenUpgradeDecisionTimeout(i interface{}) error {
	return nil
}

func validateTokenUpgradeGracePeriod(i interface{}) error {
	gracePeriod, ok := i.(time.Duration)
	if !ok {
		return errors.Errorf("invalid parameter type: %T", i)
	}
	if gracePeriod <= 0 {
		return errors.New("grace period must be greater than 0")
	}
	return nil
}
