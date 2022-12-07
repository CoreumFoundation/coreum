package types

import (
	"regexp"
	"strings"

	codetypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/nft"
)

var (
	// TODO(dhil) discuss with team the regex, probably we need uppercase as well, if yes what we do with the uniqueness?
	nftSymbolRegexStr = `^[a-z][a-z0-9]{0,40}$`
	nftSymbolRegex    = regexp.MustCompile(nftSymbolRegexStr)
	// the regexp is same as for the nft module
	nftIDRegexStr = `^[a-zA-Z][a-zA-Z0-9/:-]{2,100}$`
	nftIDRegex    = regexp.MustCompile(nftIDRegexStr)

	nftClassIDSeparator = "-"
)

// CreateNonFungibleTokenClassSettings is the model which represents the params for the non-fungible token class creation.
type CreateNonFungibleTokenClassSettings struct {
	Creator     sdk.AccAddress
	Name        string
	Symbol      string
	Description string
	URI         string
	URIHash     string
	Data        *codetypes.Any
}

// MintNonFungibleTokenSettings is the model which represents the params for the non-fungible token minting.
type MintNonFungibleTokenSettings struct {
	Sender  sdk.AccAddress
	ClassID string
	ID      string
	URI     string
	URIHash string
	Data    *codetypes.Any
}

// BuildNonFungibleTokenClassID builds the non-fungible token id string from the symbol and issuer address.
func BuildNonFungibleTokenClassID(symbol string, creator sdk.AccAddress) string {
	return strings.ToLower(symbol) + nftClassIDSeparator + creator.String()
}

// DeconstructNonFungibleTokenClassID splits the classID string into the symbol and issuer address.
func DeconstructNonFungibleTokenClassID(classID string) (issuer sdk.Address, err error) {
	classIDParts := strings.Split(classID, nftClassIDSeparator)
	if len(classIDParts) != 2 {
		return nil, sdkerrors.Wrap(ErrInvalidDenom, "classID must match format [symbol]-[issuer-address]")
	}

	address, err := sdk.AccAddressFromBech32(classIDParts[len(classIDParts)-1])
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidDenom, "invalid issuer address in classID,err:%s", err)
	}

	return address, nil
}

// ValidateNonFungibleTokenClassSymbol checks the provided non-fungible token class symbol is valid.
func ValidateNonFungibleTokenClassSymbol(symbol string) error {
	if !nftSymbolRegex.MatchString(symbol) {
		return sdkerrors.Wrapf(ErrInvalidSymbol, "symbol must match regex format '%s'", nftSymbolRegexStr)
	}

	return nil
}

// ValidateNonFungibleTokenID checks the provided non-fungible token class symbol is valid.
func ValidateNonFungibleTokenID(id string) error {
	if !nftIDRegex.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidID, "id must match regex format '%s'", nftIDRegexStr)
	}

	if err := nft.ValidateNFTID(id); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, err.Error())
	}

	return nil
}
