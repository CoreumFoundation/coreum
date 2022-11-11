package types

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	symbolRegexStr = `^[a-z][a-z0-9]{2,70}$`
	symbolRegex    *regexp.Regexp
)

func init() {
	symbolRegex = regexp.MustCompile(symbolRegexStr)
}

// IssueFungibleTokenSettings is the model which represents the params for the fungible token issuance.
type IssueFungibleTokenSettings struct {
	Issuer        sdk.AccAddress
	Symbol        string
	Description   string
	Recipient     sdk.AccAddress
	InitialAmount sdk.Int
	Features      []FungibleTokenFeature
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(symbol string, issuer sdk.AccAddress) string {
	return strings.ToLower(symbol) + "-" + issuer.String()
}

// ValidateSymbol checks the provide symbol is valid
func ValidateSymbol(symbol string) error {
	symbol = strings.ToLower(symbol)
	if symbol == "core" || symbol == "ucore" {
		return sdkerrors.Wrapf(ErrInvalidSymbol, "%s is a reserved symbol", symbol)
	}

	if !symbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidSymbol, "symbol must match regex format '%s'", symbolRegexStr)
	}

	return nil
}
