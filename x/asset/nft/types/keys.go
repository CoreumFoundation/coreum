package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

const (
	// ModuleName defines the module name
	ModuleName = "assetnft"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// NFTClassKeyPrefix defines the key prefix for the non-fungible token class definition.
	NFTClassKeyPrefix = []byte{0x01}
	// NFTFreezingKeyPrefix defines the key prefix to track frozen NFTs.
	NFTFreezingKeyPrefix = []byte{0x02}
)

// CreateClassKey constructs the key for the non-fungible token class.
func CreateClassKey(classID string) []byte {
	return store.JoinKeys(NFTClassKeyPrefix, []byte(classID))
}

// CreateFreezingKey constructs the key for the freezing of non-fungible token.
func CreateFreezingKey(classID, nftID string) ([]byte, error) {
	return store.JoinKeysWithLength([]byte(classID), []byte(nftID))
}

// ParseFreezingKey parses freezing key back to class id and nft id
func ParseFreezingKey(key []byte) (classID, nftID string, err error) {
	parsedKeys, err := store.ParseLengthPrefixedKeys(key)
	if err != nil {
		return
	}
	if len(parsedKeys) != 2 {
		err = sdkerrors.Wrapf(ErrInvalidKey, "freezing key must be composed to 2 length prefixed keys")
		return
	}
	classID = string(parsedKeys[0])
	nftID = string(parsedKeys[1])
	return classID, nftID, nil
}
