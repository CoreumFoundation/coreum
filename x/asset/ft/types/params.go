package types

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultTokenUpgradeGracePeriod is the period after which upgrade is effectively executed.
const DefaultTokenUpgradeGracePeriod = time.Hour * 24 * 7

// DefaultTokenUpgradeDecisionTimeout is the timeout for a decision to upgrade the token.
var DefaultTokenUpgradeDecisionTimeout = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

var (
	// KeyIssueFee represents the issue fee param key.
	KeyIssueFee = []byte("IssueFee")

	// KeyTokenUpgradeDecisionTimeout represents the token upgrade decision timeout param key.
	KeyTokenUpgradeDecisionTimeout = []byte("TokenUpgradeDecisionTimeout")

	// KeyTokenUpgradeGracePeriod represents the token upgrade grace period param key.
	KeyTokenUpgradeGracePeriod = []byte("TokenUpgradeGracePeriod")
)

// DefaultParams returns params with default values.
func DefaultParams() Params {
	return Params{
		IssueFee:                    sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
		TokenUpgradeDecisionTimeout: DefaultTokenUpgradeDecisionTimeout,
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
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if fee.IsNil() || !fee.IsValid() {
		return sdkerrors.Wrap(ErrInvalidInput, "issue fee must be a non-negative value")
	}
	return nil
}

func validateTokenUpgradeDecisionTimeout(i interface{}) error {
	decisionTimeout, ok := i.(time.Time)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if decisionTimeout.Before(DefaultTokenUpgradeDecisionTimeout) {
		return sdkerrors.Wrapf(ErrInvalidInput, "decision timeout cannot be set before %s", DefaultTokenUpgradeDecisionTimeout)
	}

	return nil
}

func validateTokenUpgradeGracePeriod(i interface{}) error {
	gracePeriod, ok := i.(time.Duration)
	if !ok {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid parameter type: %T", i)
	}
	if gracePeriod <= 0 {
		return sdkerrors.Wrap(ErrInvalidInput, "grace period must be greater than 0")
	}
	return nil
}
