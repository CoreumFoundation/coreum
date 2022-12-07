package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

var (
	symbolRegexStr = `^[a-z][a-z0-9]{2,70}$`
	symbolRegex    *regexp.Regexp
)

const (
	denomSeparator = "-"
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
	BurnRate      float32
}

// BuildFungibleTokenDenom builds the denom string from the symbol and issuer address.
func BuildFungibleTokenDenom(symbol string, issuer sdk.AccAddress) string {
	return strings.ToLower(symbol) + "-" + issuer.String()
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

// ValidateSymbol checks the provide symbol is valid
func ValidateSymbol(symbol string) error {
	symbol = strings.ToLower(symbol)
	if lo.Contains(reserved, symbol) {
		return sdkerrors.Wrapf(ErrInvalidSymbol, "%s is a reserved symbol", symbol)
	}

	if !symbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidSymbol, "symbol must match regex format '%s'", symbolRegexStr)
	}

	return nil
}

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (ftd *FungibleTokenDefinition) IsFeatureEnabled(feature FungibleTokenFeature) bool {
	return lo.Contains(ftd.Features, feature)
}

// ValidateBurnRate checks the provide burn rate is valid
func ValidateBurnRate(burnRate float32) error {
	if burnRate < 0 || burnRate > 1 {
		return sdkerrors.Wrap(ErrInvalidBurnRate, "burn rate is not within acceptable rate")
	}

	return nil
}

// CalculateBurnCoin returns the coins to be burned
func (ftd FungibleTokenDefinition) CalculateBurnCoin(coin sdk.Coin) sdk.Coin {
	if ftd.BurnRate > 1 {
		panic("burn rate cannot be bigger than 1")
	}

	// limit precision to 4 decimal places
	burnRateStr := fmt.Sprintf("%.4f", ftd.BurnRate)
	// we convert float32 to string and parse it to big.Float because direct conversion from
	// float32 to big.Float leads to some rounding errors
	burnRate, _, err := big.ParseFloat(burnRateStr, 10, 100, big.ToNearestAway)
	if err != nil {
		panic(err)
	}

	amount := big.NewFloat(0).SetInt(coin.Amount.BigInt())
	burnAmountFloat := big.NewFloat(0).Mul(amount, burnRate)
	burnAmount, accuracy := burnAmountFloat.Int(nil)
	str := burnAmountFloat.String()
	burnRateStr = burnRate.String()
	fmt.Print(str, burnRateStr)
	if accuracy != big.Exact {
		burnAmount = big.NewInt(0).Add(burnAmount, big.NewInt(1))
	}

	return sdk.NewCoin(coin.Denom, sdk.NewIntFromBigInt(burnAmount))
}
