package types

import (
	"regexp"
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/x/nft"
)

var (
	nftSymbolRegexStr = `^[a-zA-Z][a-zA-Z0-9]{0,40}$`
	nftSymbolRegex    = regexp.MustCompile(nftSymbolRegexStr)
	// the regexp is same as for the nft module.
	nftIDRegexStr = `^[a-zA-Z][a-zA-Z0-9/:-]{2,100}$`
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
	RoyaltyRate sdk.Dec
}

// MintSettings is the model which represents the params for the non-fungible token minting.
type MintSettings struct {
	Sender  sdk.AccAddress
	ClassID string
	ID      string
	URI     string
	URIHash string
	Data    *codectypes.Any
}

// BuildClassID builds the non-fungible token id string from the symbol and issuer address.
func BuildClassID(symbol string, issuer sdk.AccAddress) string {
	return strings.ToLower(symbol) + nftClassIDSeparator + issuer.String()
}

// DeconstructClassID splits the classID string into the symbol and issuer address.
func DeconstructClassID(classID string) (issuer sdk.Address, err error) {
	classIDParts := strings.Split(classID, nftClassIDSeparator)
	if len(classIDParts) != 2 {
		return nil, sdkerrors.Wrap(ErrInvalidInput, "classID must match format [symbol]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(classIDParts[1])
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidInput, "invalid issuer address in classID,err:%s", err)
	}

	return address, nil
}

// ValidateClassSymbol checks the provided non-fungible token class symbol is valid.
func ValidateClassSymbol(symbol string) error {
	if !nftSymbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidInput, "symbol must match regex format '%s'", nftSymbolRegexStr)
	}

	return nil
}

// ValidateTokenID checks the provided non-fungible token class symbol is valid.
func ValidateTokenID(id string) error {
	if !nftIDRegex.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidID, "id must match regex format '%s'", nftIDRegexStr)
	}

	if err := nft.ValidateNFTID(id); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, err.Error())
	}

	return nil
}

// ValidateData checks the provided data field is valid for NFT class or token.
func ValidateData(data *codectypes.Any) error {
	if data != nil {
		if len(data.Value) > MaxDataSize {
			return sdkerrors.Wrapf(ErrInvalidInput, "invalid data, it's allowed to use %d bytes", MaxDataSize)
		}
		if data.TypeUrl != "/"+proto.MessageName((*DataBytes)(nil)) {
			return sdkerrors.Wrapf(ErrInvalidInput, "data field must contain %s type", proto.MessageName((*DataBytes)(nil)))
		}
	}

	return nil
}

// ValidateRoyaltyRate checks the provided non-fungible token royalty rate is valid.
func ValidateRoyaltyRate(rate sdk.Dec) error {
	if rate.IsNil() {
		return nil
	}

	if rate.GT(sdk.NewDec(1)) || rate.LT(sdk.NewDec(0)) {
		return sdkerrors.Wrapf(ErrInvalidInput, "royalty rate should be between 0 and 1")
	}

	return nil
}

// CheckFeatureAllowed returns error if feature isn't allowed for the address.
func (nftd ClassDefinition) CheckFeatureAllowed(addr sdk.AccAddress, feature ClassFeature) error {
	if nftd.IsFeatureAllowed(addr, feature) {
		return nil
	}

	if !nftd.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(ErrFeatureDisabled, "feature %s is disabled", feature.String())
	}

	return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %s is unauthorized to perform %q related operations", addr.String(), feature.String())
}

// IsFeatureAllowed returns true if feature is allowed for the address.
func (nftd ClassDefinition) IsFeatureAllowed(addr sdk.Address, feature ClassFeature) bool {
	featureEnabled := nftd.IsFeatureEnabled(feature)
	// issuer can use any enabled feature and burning even if it is disabled
	if nftd.IsIssuer(addr) {
		return featureEnabled || feature == ClassFeature_burning
	}

	// non-issuer can use only burning and only if it is enabled
	return featureEnabled && feature == ClassFeature_burning
}

// IsFeatureEnabled returns true if feature is enabled for a fungible token.
func (nftd ClassDefinition) IsFeatureEnabled(feature ClassFeature) bool {
	return lo.Contains(nftd.Features, feature)
}

// IsIssuer returns true if the addr is the issuer.
func (nftd ClassDefinition) IsIssuer(addr sdk.Address) bool {
	return nftd.Issuer == addr.String()
}
