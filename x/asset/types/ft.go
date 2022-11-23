package types

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

var (
	subunitRegexStr = `^[a-z][a-z0-9]{2,70}$`
	subunitRegex    *regexp.Regexp
)

const (
	denomSeparator = "-"
)

func init() {
	subunitRegex = regexp.MustCompile(subunitRegexStr)
}

// IssueFungibleTokenSettings is the model which represents the params for the fungible token issuance.
type IssueFungibleTokenSettings struct {
	Issuer        sdk.AccAddress
	Symbol        string
	Subunit       string
	Precision     uint32
	Description   string
	Recipient     sdk.AccAddress
	InitialAmount sdk.Int
	Features      []FungibleTokenFeature
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(prefix string, issuer sdk.AccAddress) string {
	return strings.ToLower(prefix) + denomSeparator + issuer.String()
}

// DeconstructFungibleTokenDenom splits the denom string into the symbol and issuer address.
func DeconstructFungibleTokenDenom(denom string) (prefix string, issuer sdk.Address, err error) {
	denomParts := strings.Split(denom, denomSeparator)
	if len(denomParts) != 2 {
		return "", nil, sdkerrors.Wrap(ErrInvalidDenom, "symbol must match format [subunit]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(denomParts[1])
	if err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidDenom, "invalid issuer address in denom,err:%s", err)
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

// ValidateSubunit checks the provide symbol is valid
func ValidateSubunit(subunit string) error {
	subunit = strings.ToLower(subunit)
	if lo.Contains(reserved, subunit) {
		return sdkerrors.Wrapf(ErrInvalidSubunit, "%s is a reserved symbol", subunit)
	}

	if !subunitRegex.MatchString(subunit) {
		return sdkerrors.Wrapf(ErrInvalidSubunit, "subunit must match regex format '%s'", subunitRegexStr)
	}

	return nil
}
