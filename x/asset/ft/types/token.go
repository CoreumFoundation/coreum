package types

import (
	"math"
	"math/big"
	"regexp"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

var (
	// The length is 51 since the demon is {subunit}-{address} and
	// the address length might be up to 66 symbols and the demon length must be less than 127 symbols
	// according to bank validation. This leaves 61 spaces, and we choose 51 to leave some room for
	// future changes.
	subunitRegexStr = `^[a-z][a-z0-9/:._]{0,50}$`
	subunitRegex    *regexp.Regexp

	symbolRegexStr = `^[a-zA-Z][a-zA-Z0-9/:._-]{2,127}$`
	symbolRegex    *regexp.Regexp
)

const (
	// CurrentTokenVersion is the version of the token produced by the current version of the app.
	CurrentTokenVersion = 1

	denomSeparator = "-"
	// MaxPrecision used when issuing a token.
	MaxPrecision = 20
)

// MaxMintableAmount is the maximum amount of a coin that can be minted at a time.
var MaxMintableAmount = sdkmath.NewIntFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(50), nil))

// MaxUnifiedRefAmount is the maximum value for unified ref amount.
var MaxUnifiedRefAmount = sdkmath.LegacyNewDecFromBigInt(big.NewInt(0).Exp(big.NewInt(10), big.NewInt(50), nil))

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
	URI                string
	URIHash            string
	InitialAmount      sdkmath.Int
	Features           []Feature
	BurnRate           sdkmath.LegacyDec
	SendCommissionRate sdkmath.LegacyDec
	ExtensionSettings  *ExtensionIssueSettings
	DEXSettings        *DEXSettings
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
		return "", nil, sdkerrors.Wrapf(
			ErrInvalidDenom,
			"invalid issuer address %q in denom: %s, err:%s",
			denomParts[1],
			denom,
			err,
		)
	}

	if err := ValidateSubunit(denomParts[0]); err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidDenom, "invalid subunit, err:%s", err)
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

// ValidateSubunit checks the provided subunit is valid.
func ValidateSubunit(subunit string) error {
	if lo.Contains(reserved, strings.ToLower(subunit)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "%s is a reserved subunit", subunit)
	}

	if strings.HasPrefix(strings.ToLower(subunit), ibctypes.DenomPrefix) {
		return sdkerrors.Wrapf(ErrInvalidInput, "subunit cannot start with ibc")
	}

	if !subunitRegex.MatchString(subunit) {
		return sdkerrors.Wrapf(ErrInvalidInput, "subunit must match regex format '%s'", subunitRegexStr)
	}

	return nil
}

// ValidateAssetCoin checks that the coin is a valid coin according to asset ft module restrictions.
func ValidateAssetCoin(coin sdk.Coin) error {
	if _, _, err := DeconstructDenom(coin.Denom); err != nil {
		return err
	}

	return coin.Validate()
}

// ValidateAssetCoins checks that the coins is valid according to asset ft module restrictions.
func ValidateAssetCoins(coins sdk.Coins) error {
	for _, coin := range coins {
		if err := ValidateAssetCoin(coin); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePrecision checks the provided precision is valid.
func ValidatePrecision(precision uint32) error {
	if precision > MaxPrecision {
		return sdkerrors.Wrapf(ErrInvalidInput, "precision must be between 0 and %d", MaxPrecision)
	}
	return nil
}

// ValidateSymbol checks the provided symbol is valid.
func ValidateSymbol(symbol string) error {
	if lo.Contains(reserved, strings.ToLower(symbol)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "%s is a reserved symbol", symbol)
	}

	if !symbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidInput, "symbol must match regex format '%s'", symbolRegexStr)
	}

	return nil
}

// NormalizeSymbolForKey normalizes the symbol string.
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

	return sdkerrors.Wrapf(
		cosmoserrors.ErrUnauthorized,
		"address %s is unauthorized to perform %q related operations",
		addr.String(),
		feature.String(),
	)
}

// IsFeatureAllowed returns true if feature is allowed for the address.
func (def Definition) IsFeatureAllowed(addr sdk.Address, feature Feature) bool {
	featureEnabled := def.IsFeatureEnabled(feature)
	// token admin and asset extension contract can use any enabled feature and burning even if it is disabled
	if def.HasAdminPrivileges(addr) {
		return featureEnabled || feature == Feature_burning
	}

	// non-issuer can use only burning and only if it is enabled
	return featureEnabled && feature == Feature_burning
}

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (def Definition) IsFeatureEnabled(feature Feature) bool {
	return lo.Contains(def.Features, feature)
}

// IsAdmin returns true if the addr is the admin.
func (def Definition) IsAdmin(addr sdk.Address) bool {
	return def.Admin == addr.String()
}

// HasAdminPrivileges returns true if the addr is the admin or the asset extension contract address.
func (def Definition) HasAdminPrivileges(addr sdk.Address) bool {
	return def.Admin == addr.String() || def.ExtensionCWAddress == addr.String()
}

// ValidateFeatures verifies that provided features belong to the defined set.
func ValidateFeatures(features []Feature) error {
	present := map[Feature]struct{}{}
	hasExtension := false
	hasDEXBlock := false
	for _, f := range features {
		name, exists := Feature_name[int32(f)]
		if !exists {
			return sdkerrors.Wrapf(ErrInvalidInput, "non-existing feature provided: %d", f)
		}
		if _, exists := present[f]; exists {
			return sdkerrors.Wrapf(ErrInvalidInput, "duplicated feature: %s", name)
		}
		if f == Feature_extension {
			hasExtension = true
		}
		if f == Feature_dex_block {
			hasDEXBlock = true
		}
		present[f] = struct{}{}
	}
	for _, feature := range features {
		if hasExtension && feature == Feature_block_smart_contracts {
			return sdkerrors.Wrapf(ErrInvalidInput, "extension is not allowed in combination with %s", feature.String())
		}
		// if dex is blocked those features make not sense
		if hasDEXBlock && (feature == Feature_dex_whitelisted_denoms ||
			feature == Feature_dex_order_cancellation ||
			feature == Feature_dex_unified_ref_amount_change) {
			return sdkerrors.Wrapf(
				ErrInvalidInput,
				"%s is not allowed in combination with %s", Feature_dex_block.String(), feature.String(),
			)
		}
	}

	return nil
}

// ValidateBurnRate checks that the provided burn rate is valid.
func ValidateBurnRate(burnRate sdkmath.LegacyDec) error {
	if err := validateRate(burnRate); err != nil {
		return errors.Wrap(err, "burn rate is invalid")
	}
	return nil
}

// ValidateSendCommissionRate checks that provided send commission rate is valid.
func ValidateSendCommissionRate(sendCommissionRate sdkmath.LegacyDec) error {
	if err := validateRate(sendCommissionRate); err != nil {
		return errors.Wrap(err, "send commission rate is invalid")
	}
	return nil
}

func validateRate(rate sdkmath.LegacyDec) error {
	const maxRatePrecisionAllowed = 4

	if rate.IsNil() {
		return nil
	}

	if !isDecPrecisionValid(rate, maxRatePrecisionAllowed) {
		return sdkerrors.Wrap(ErrInvalidInput, "rate precision should not be more than 4 decimal places")
	}

	if rate.LT(sdkmath.LegacyNewDec(0)) || rate.GT(sdkmath.LegacyNewDec(1)) {
		return sdkerrors.Wrap(ErrInvalidInput, "rate is not within acceptable range")
	}

	return nil
}

// ValidateDEXSettings checks that provided DEX settings are valid.
func ValidateDEXSettings(settings DEXSettings) error {
	if settings.UnifiedRefAmount != nil {
		if err := ValidateUnifiedRefAmount(*settings.UnifiedRefAmount); err != nil {
			return err
		}
	}

	return ValidateWhitelistedDenoms(settings.WhitelistedDenoms)
}

// ValidateDEXSettingsAccess checks that provided DEX settings are allowed to be accessed.
func ValidateDEXSettingsAccess(settings DEXSettings, def Definition) error {
	if def.IsFeatureEnabled(Feature_dex_block) {
		return sdkerrors.Wrapf(
			ErrFeatureDisabled,
			"it's prohibited to providy any DEX setting if the %s feature is enabled",
			Feature_dex_block.String(),
		)
	}
	if settings.WhitelistedDenoms != nil && !def.IsFeatureEnabled(Feature_dex_whitelisted_denoms) {
		return sdkerrors.Wrapf(
			ErrFeatureDisabled,
			"it's prohibited to providy any DEX whitelisted denoms if the %s feature is disabled",
			Feature_dex_whitelisted_denoms.String(),
		)
	}
	if settings.UnifiedRefAmount != nil && !def.IsFeatureEnabled(Feature_dex_unified_ref_amount_change) {
		return sdkerrors.Wrapf(
			ErrFeatureDisabled,
			"it's prohibited to providy unified ref amount if the %s feature is disabled",
			Feature_dex_unified_ref_amount_change.String(),
		)
	}

	return nil
}

// ValidateUnifiedRefAmount checks that provided unified ref amount is valid.
func ValidateUnifiedRefAmount(unifiedRefAmount sdkmath.LegacyDec) error {
	if !unifiedRefAmount.IsPositive() {
		return sdkerrors.Wrap(ErrInvalidInput, "unified ref amount must be positive")
	}

	if unifiedRefAmount.GT(MaxUnifiedRefAmount) {
		return sdkerrors.Wrap(ErrInvalidInput, "unified ref amount is too big")
	}

	return nil
}

// ValidateWhitelistedDenoms checks that provided whitelisted denoms are valid.
func ValidateWhitelistedDenoms(whitelistedDenoms []string) error {
	duplicates := lo.FindDuplicates(whitelistedDenoms)
	if len(duplicates) != 0 {
		return sdkerrors.Wrapf(
			ErrInvalidInput, "duplicated denoms in the denoms to trade with list, duplicates: %v", duplicates,
		)
	}
	for _, denom := range whitelistedDenoms {
		if err := sdk.ValidateDenom(denom); err != nil {
			return sdkerrors.Wrapf(
				ErrInvalidInput,
				"invalid denom in the denoms to trade with list, denom: %s, err: %s",
				denom, err,
			)
		}
	}
	return nil
}

// checks that dec precision is limited to the provided value.
func isDecPrecisionValid(dec sdkmath.LegacyDec, prec uint) bool {
	return dec.Mul(sdkmath.LegacyNewDecFromInt(sdkmath.NewInt(int64(math.Pow10(int(prec)))))).IsInteger()
}
