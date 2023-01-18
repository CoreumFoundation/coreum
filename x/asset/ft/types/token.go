package types

import (
	"math"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

var (
	subunitRegexStr = `^[a-z][a-z0-9]{0,70}$`
	subunitRegex    *regexp.Regexp

	symbolRegexStr = `^[a-zA-Z][a-zA-Z0-9-.]{0,127}$`
	symbolRegex    *regexp.Regexp
)

const (
	denomSeparator = "-"
)

func init() {
	subunitRegex = regexp.MustCompile(subunitRegexStr)
	symbolRegex = regexp.MustCompile(symbolRegexStr)
}

// IssueSettings is the model which represents the params for the fungible token issuance.
type IssueSettings struct {
	Issuer             sdk.AccAddress
	Symbol             string
	Subunit            string
	Precision          uint32
	Description        string
	InitialAmount      sdk.Int
	Features           []Feature
	BurnRate           sdk.Dec
	SendCommissionRate sdk.Dec
}

// BuildDenom builds the denom string from the symbol and issuer address.
func BuildDenom(subunit string, issuer sdk.AccAddress) string {
	return strings.ToLower(subunit) + denomSeparator + issuer.String()
}

// DeconstructDenom splits the denom string into the symbol and issuer address.
func DeconstructDenom(denom string) (prefix string, issuer sdk.AccAddress, err error) {
	denomParts := strings.Split(denom, denomSeparator)
	if len(denomParts) != 2 {
		return "", nil, sdkerrors.Wrap(ErrInvalidDenom, "denom must match format [subunit]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(denomParts[1])
	if err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidDenom, "invalid issuer address in denom, err:%s", err)
	}

	return denomParts[0], address, nil
}

var reserved = []string{
	strings.ToLower(constant.DenomDev),
	strings.ToLower(constant.DenomDevDisplay),
	strings.ToLower(constant.DenomTest),
	strings.ToLower(constant.DenomTestDisplay),
	strings.ToLower(constant.DenomMain),
	strings.ToLower(constant.DenomMainDisplay),
}

// ValidateSubunit checks the provide subunit is valid
func ValidateSubunit(subunit string) error {
	if lo.Contains(reserved, strings.ToLower(subunit)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "%s is a reserved subunit", subunit)
	}

	if !subunitRegex.MatchString(subunit) {
		return sdkerrors.Wrapf(ErrInvalidInput, "subunit must match regex format '%s'", subunitRegexStr)
	}

	return nil
}

// ValidateSymbol checks the provided symbol is valid
func ValidateSymbol(symbol string) error {
	if lo.Contains(reserved, strings.ToLower(symbol)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "%s is a reserved symbol", symbol)
	}

	if !symbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidInput, "symbol must match regex format '%s'", symbolRegexStr)
	}

	return nil
}

// NormalizeSymbolForKey normalizes the symbol string
func NormalizeSymbolForKey(in string) string {
	return strings.ToLower(in)
}

// CheckFeatureAllowed returns error if feature isn't allowed for the address.
func (def Definition) CheckFeatureAllowed(addr sdk.AccAddress, feature Feature) error {
	if def.IsFeatureAllowed(addr, feature) {
		return nil
	}

	if !def.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(ErrFeatureDisabled, "feature %s is disabled", feature.String())
	}

	return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %s is unauthorized to perform %q related operations", addr.String(), feature.String())
}

// IsFeatureAllowed returns true if feature is allowed for the address.
//
//nolint:nosnakecase
func (def Definition) IsFeatureAllowed(addr sdk.Address, feature Feature) bool {
	featureEnabled := def.IsFeatureEnabled(feature)
	// issuer can use any enabled feature and burning even if it is disabled
	if def.IsIssuer(addr) {
		return featureEnabled || feature == Feature_burning
	}

	// non-issuer can use only burning and only if it is enabled
	return featureEnabled && feature == Feature_burning
}

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (def Definition) IsFeatureEnabled(feature Feature) bool {
	return lo.Contains(def.Features, feature)
}

// IsIssuer returns true if the addr is the issuer.
func (def Definition) IsIssuer(addr sdk.Address) bool {
	return def.Issuer == addr.String()
}

// ValidateBurnRate checks that provided burn rate is valid
func ValidateBurnRate(burnRate sdk.Dec) error {
	if err := validateRate(burnRate); err != nil {
		return errors.Wrap(err, "burn rate is invalid")
	}
	return nil
}

// ValidateSendCommissionRate checks that provided send commission rate is valid
func ValidateSendCommissionRate(sendCommissionRate sdk.Dec) error {
	if err := validateRate(sendCommissionRate); err != nil {
		return errors.Wrap(err, "send commission rate is invalid")
	}
	return nil
}

func validateRate(rate sdk.Dec) error {
	const maxRatePrecisionAllowed = 4

	if rate.IsNil() {
		return nil
	}

	if !isDecPrecisionValid(rate, maxRatePrecisionAllowed) {
		return sdkerrors.Wrap(ErrInvalidInput, "rate precision should not be more than 4 decimal places")
	}

	if rate.LT(sdk.NewDec(0)) || rate.GT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(ErrInvalidInput, "rate is not within acceptable range")
	}

	return nil
}

// checks that dec precision is limited to the provided value
func isDecPrecisionValid(dec sdk.Dec, prec uint) bool {
	return dec.Mul(sdk.NewDecFromInt(sdk.NewInt(int64(math.Pow10(int(prec)))))).IsInteger()
}
