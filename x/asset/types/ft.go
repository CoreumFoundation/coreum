package types

import (
	"math"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// IssueFungibleTokenSettings is the model which represents the params for the fungible token issuance.
type IssueFungibleTokenSettings struct {
	Issuer        sdk.AccAddress
	Symbol        string
	Subunit       string
	Precision     uint32
	Description   string
	InitialAmount sdk.Int
	Features      []FungibleTokenFeature
	BurnRate      sdk.Dec
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(subunit string, issuer sdk.AccAddress) string {
	return strings.ToLower(subunit) + denomSeparator + issuer.String()
}

// DeconstructFungibleTokenDenom splits the denom string into the symbol and issuer address.
func DeconstructFungibleTokenDenom(denom string) (prefix string, issuer sdk.Address, err error) {
	denomParts := strings.Split(denom, denomSeparator)
	if len(denomParts) != 2 {
		return "", nil, sdkerrors.Wrap(ErrInvalidInput, "symbol must match format [subunit]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(denomParts[1])
	if err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidInput, "invalid issuer address in denom,err:%s", err)
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

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (ftd *FungibleTokenDefinition) IsFeatureEnabled(feature FungibleTokenFeature) bool {
	return lo.Contains(ftd.Features, feature)
}

// ValidateBurnRate checks the provide burn rate is valid
func ValidateBurnRate(burnRate sdk.Dec) error {
	if burnRate.IsNil() {
		return nil
	}

	if !isDecPrecisionValid(burnRate, 4) {
		return sdkerrors.Wrap(ErrInvalidInput, "burn rate precision should not be more than 4 decimal places")
	}

	if burnRate.LT(sdk.NewDec(0)) || burnRate.GT(sdk.NewDec(1)) {
		return sdkerrors.Wrap(ErrInvalidInput, "burn rate is not within acceptable range")
	}

	return nil
}

// checks that dec precision is limited to the provided value
func isDecPrecisionValid(dec sdk.Dec, prec uint) bool {
	return dec.Mul(sdk.NewDecFromInt(sdk.NewInt(int64(math.Pow10(int(prec)))))).IsInteger()
}

// CalculateBurnRateAmount returns the coins to be burned
func (ftd FungibleTokenDefinition) CalculateBurnRateAmount(coin sdk.Coin) sdk.Int {
	return ftd.BurnRate.MulInt(coin.Amount).Ceil().RoundInt()
}
