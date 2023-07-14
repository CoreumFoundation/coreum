package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

const (
	// ModuleName defines the module name.
	ModuleName = "assetnft"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// QuerierRoute defines the module's query routing key.
	QuerierRoute = ModuleName

	// RouterKey is the message route for module.
	RouterKey = ModuleName
)

// Store key prefixes.
var (
	// NFTClassKeyPrefix defines the key prefix for the non-fungible token class definition.
	NFTClassKeyPrefix = []byte{0x01}
	// NFTFreezingKeyPrefix defines the key prefix to track frozen NFTs.
	NFTFreezingKeyPrefix = []byte{0x02}
	// NFTWhitelistingKeyPrefix defines the key prefix to track whitelisted account.
	NFTWhitelistingKeyPrefix = []byte{0x03}
	// NFTBurningKeyPrefix defines the key prefix to track burnt NFTs.
	NFTBurningKeyPrefix = []byte{0x04}
)

// CreateClassKey constructs the key for the non-fungible token class.
func CreateClassKey(classID string) ([]byte, error) {
	symbol, issuer, err := DeconstructClassID(classID)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidKey, "can't build class key from classID, classID:%s, err:%s", classID, err)
	}
	// use keys in the reverse order to query by the issuer
	classKey, err := store.JoinKeysWithLength(issuer, []byte(symbol))
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidKey, "can't join NFT class key with length, issuer:%s, symbol:%s, err:%s", issuer, symbol, err)
	}

	return store.JoinKeys(NFTClassKeyPrefix, classKey), nil
}

// CreateIssuerClassPrefix constructs the key for the non-fungible token class for the specific issuer.
func CreateIssuerClassPrefix(issuer sdk.AccAddress) ([]byte, error) {
	issuerKey, err := store.JoinKeysWithLength(issuer)
	if err != nil {
		return nil, sdkerrors.Wrapf(ErrInvalidKey, "can't join NFT class issuer key with length, issuer:%s, err:%s", issuer, err)
	}
	return store.JoinKeys(NFTClassKeyPrefix, issuerKey), nil
}

// CreateFreezingKey constructs the key for the freezing of non-fungible token.
func CreateFreezingKey(classID, nftID string) ([]byte, error) {
	compositeKey, err := store.JoinKeysWithLength([]byte(classID), []byte(nftID))
	if err != nil {
		return nil, err
	}

	return store.JoinKeys(NFTFreezingKeyPrefix, compositeKey), nil
}

// ParseFreezingKey parses freezing key back to class id and nft id.
func ParseFreezingKey(key []byte) (string, string, error) {
	parsedKeys, err := store.ParseLengthPrefixedKeys(key)
	if err != nil {
		return "", "", err
	}
	if len(parsedKeys) != 2 {
		err = sdkerrors.Wrapf(ErrInvalidKey, "freezing key must be composed to 2 length prefixed keys")
		return "", "", err
	}
	return string(parsedKeys[0]), string(parsedKeys[1]), nil
}

// CreateWhitelistingKey constructs the key for the whitelisting of non-fungible token.
func CreateWhitelistingKey(classID, nftID string, account sdk.AccAddress) ([]byte, error) {
	compositeKey, err := store.JoinKeysWithLength([]byte(classID), []byte(nftID), account)
	if err != nil {
		return nil, err
	}

	return store.JoinKeys(NFTWhitelistingKeyPrefix, compositeKey), nil
}

// ParseWhitelistingKey parses freezing key back to class id and nft id.
func ParseWhitelistingKey(key []byte) (string, string, sdk.AccAddress, error) {
	parsedKeys, err := store.ParseLengthPrefixedKeys(key)
	if err != nil {
		return "", "", nil, err
	}
	if len(parsedKeys) != 3 {
		err = sdkerrors.Wrapf(ErrInvalidKey, "whitelisting key must be composed of 3 length prefixed keys")
		return "", "", nil, err
	}
	return string(parsedKeys[0]), string(parsedKeys[1]), parsedKeys[2], nil
}

// CreateBurningKey constructs the key for the burning of non-fungible token.
func CreateBurningKey(classID, nftID string) ([]byte, error) {
	compositeKey, err := store.JoinKeysWithLength([]byte(classID), []byte(nftID))
	if err != nil {
		return nil, err
	}

	return store.JoinKeys(NFTBurningKeyPrefix, compositeKey), nil
}

// CreateClassBurningKey constructs the key for the burning of non-fungible token.
func CreateClassBurningKey(classID string) ([]byte, error) {
	compositeKey, err := store.JoinKeysWithLength([]byte(classID))
	if err != nil {
		return nil, err
	}

	return store.JoinKeys(NFTBurningKeyPrefix, compositeKey), nil
}

// ParseBurningKey parses burning key back to class id and nft id.
func ParseBurningKey(key []byte) (string, string, error) {
	parsedKeys, err := store.ParseLengthPrefixedKeys(key)
	if err != nil {
		return "", "", err
	}
	if len(parsedKeys) != 2 {
		err = sdkerrors.Wrapf(ErrInvalidKey, "burning key must be composed to 2 length prefixed keys")
		return "", "", err
	}
	return string(parsedKeys[0]), string(parsedKeys[1]), nil
}
