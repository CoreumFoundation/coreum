package types

import (
	"regexp"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
	"github.com/samber/lo"
)

var (
	// The length is 31 and this is the reasoning for it.
	// since the class id has {symbol}-{address} format and
	// the address length might be up to 66 symbols and the demon length must be less than 101
	// this leaves room for 33 characters, but we choose 31 to be conservative.
	nftSymbolRegexStr = `^[a-zA-Z][a-zA-Z0-9/:._]{0,30}$`
	nftSymbolRegex    = regexp.MustCompile(nftSymbolRegexStr)
	// the regexp is same as for the nft module.
	nftIDRegexStr = `^[a-zA-Z][a-zA-Z0-9/:._-]{2,100}$`
	nftIDRegex    = regexp.MustCompile(nftIDRegexStr)

	nftClassIDSeparator = "-"
)

// IssueClassSettings is the model which represents the params for the non-fungible token class creation.
type IssueClassSettings struct {
	Issuer      sdk.AccAddress
	Name        string
	Symbol      string
	Description string
	URI         string
	URIHash     string
	Data        *codectypes.Any
	Features    []ClassFeature
	RoyaltyRate sdkmath.LegacyDec
}

// MintSettings is the model which represents the params for the non-fungible token minting.
type MintSettings struct {
	Sender    sdk.AccAddress
	Recipient sdk.AccAddress
	ClassID   string
	ID        string
	URI       string
	URIHash   string
	Data      *codectypes.Any
}

// BuildClassID builds the non-fungible token id string from the symbol and issuer address.
func BuildClassID(symbol string, issuer sdk.AccAddress) string {
	return strings.ToLower(symbol) + nftClassIDSeparator + issuer.String()
}

// DeconstructClassID splits the classID string into the symbol and issuer address.
func DeconstructClassID(classID string) (string, sdk.AccAddress, error) {
	classIDParts := strings.Split(classID, nftClassIDSeparator)
	if len(classIDParts) != 2 {
		return "", nil, sdkerrors.Wrap(ErrInvalidInput, "classID must match format [symbol]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(classIDParts[1])
	if err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidInput, "invalid issuer address in classID,err:%s", err)
	}

	symbol := classIDParts[0]
	if err := ValidateClassSymbol(symbol); err != nil {
		return "", nil, sdkerrors.Wrapf(ErrInvalidInput, "invalid symbol in classID,err:%s", err)
	}

	// ensure that symbol is all lowercase
	if strings.ToLower(symbol) != symbol {
		return "", nil, sdkerrors.Wrapf(ErrInvalidInput, "symbol in classID should be lowercase")
	}

	return symbol, address, nil
}

// ValidateClassSymbol checks the provided non-fungible token class symbol is valid.
func ValidateClassSymbol(symbol string) error {
	if !nftSymbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidInput, "symbol must match regex format '%s'", nftSymbolRegexStr)
	}

	return nil
}

// ValidateClassFeatures verifies that provided features belong to the defined set.
func ValidateClassFeatures(features []ClassFeature) error {
	present := map[ClassFeature]struct{}{}
	for _, f := range features {
		name, exists := ClassFeature_name[int32(f)]
		if !exists {
			return sdkerrors.Wrapf(ErrInvalidInput, "non-existing class feature provided: %d", f)
		}
		if _, exists := present[f]; exists {
			return sdkerrors.Wrapf(ErrInvalidInput, "duplicated class feature: %s", name)
		}
		present[f] = struct{}{}
	}
	return nil
}

// ValidateTokenID checks the provided non-fungible token id is valid.
func ValidateTokenID(id string) error {
	if !nftIDRegex.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidID, "id must match regex format '%s'", nftIDRegexStr)
	}

	return nil
}

// ValidateClassData checks the provided class data field is valid for NFT class.
func ValidateClassData(data *codectypes.Any) error {
	if data != nil {
		if data.TypeUrl != "/"+proto.MessageName((*DataBytes)(nil)) {
			return sdkerrors.Wrapf(ErrInvalidInput, "data field must contain %s type", proto.MessageName((*DataBytes)(nil)))
		}
	}

	return nil
}

// ValidateNFTData checks the provided data field is valid for NFT token.
func ValidateNFTData(data *codectypes.Any) error {
	if data != nil {
		switch data.TypeUrl {
		case "/" + proto.MessageName((*DataBytes)(nil)):
			// no default validation of the data
		case "/" + proto.MessageName((*DataDynamic)(nil)):
			if err := validateDynamicData(data); err != nil {
				return err
			}
		default:
			return sdkerrors.Wrapf(ErrInvalidInput, "data field must contain %s or %s type",
				proto.MessageName((*DataBytes)(nil)), proto.MessageName((*DataDynamic)(nil)))
		}
	}

	return nil
}

// ValidateRoyaltyRate checks the provided non-fungible token royalty rate is valid.
func ValidateRoyaltyRate(rate sdkmath.LegacyDec) error {
	if rate.IsNil() {
		return nil
	}

	if rate.GT(sdkmath.LegacyNewDec(1)) || rate.LT(sdkmath.LegacyNewDec(0)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "royalty rate should be between 0 and 1")
	}

	return nil
}

// CheckFeatureAllowed returns error if feature isn't allowed for the address.
func (nftd ClassDefinition) CheckFeatureAllowed(addr sdk.AccAddress, feature ClassFeature) error {
	// Issuer is allowed to burn even if burning is disabled
	if nftd.IsIssuer(addr) && feature == ClassFeature_burning {
		return nil
	}

	// For all the other cases feature must be enabled
	if !nftd.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(ErrFeatureDisabled, "feature %s is disabled", feature.String())
	}

	// If burning is enabled then everyone may burn
	if feature == ClassFeature_burning {
		return nil
	}

	// Features other than burning may be executed by the issuer only
	if !nftd.IsIssuer(addr) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"address %s is unauthorized to perform %q related operations",
			addr.String(), feature.String(),
		)
	}
	return nil
}

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (nftd ClassDefinition) IsFeatureEnabled(feature ClassFeature) bool {
	return lo.Contains(nftd.Features, feature)
}

// IsIssuer returns true if the addr is the issuer.
func (nftd ClassDefinition) IsIssuer(addr sdk.Address) bool {
	return nftd.Issuer == addr.String()
}

func validateDynamicData(data *codectypes.Any) error {
	var dataDynamic DataDynamic
	if err := dataDynamic.Unmarshal(data.Value); err != nil {
		return sdkerrors.Wrap(ErrInvalidInput, "failed to unmarshal data to DataDynamic")
	}
	if len(dataDynamic.Items) == 0 {
		return sdkerrors.Wrap(ErrInvalidInput, "empty items list")
	}

	for i, item := range dataDynamic.Items {
		usedEditors := make(map[DataEditor]struct{}, 0)
		for _, editor := range item.Editors {
			_, exists := DataEditor_name[int32(editor)]
			if !exists {
				return sdkerrors.Wrapf(ErrInvalidInput, "non-existing data editor provided: %d", editor)
			}
			if _, found := usedEditors[editor]; found {
				return sdkerrors.Wrapf(ErrInvalidInput, "duplicated dynamic data editor, item %d", i)
			}
			usedEditors[editor] = struct{}{}
		}
	}

	return nil
}
