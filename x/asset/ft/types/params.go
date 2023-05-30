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

// KeyIssueFee represents the issue fee param key.
var KeyIssueFee = []byte("IssueFee")

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// of module parameters.
func (m *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyIssueFee, &m.IssueFee, validateIssueFee),
	}
}

// DefaultParams returns params with default values.
func DefaultParams(genesisTime time.Time) Params {
	return Params{
		IssueFee:                    sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
		TokenUpgradeDecisionTimeout: genesisTime.Add(DefaultTokenUpgradeDecisionPeriod),
		TokenUpgradeGracePeriod:     DefaultTokenUpgradeGracePeriod,
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
