package types

import "github.com/CoreumFoundation/coreum/pkg/store"

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
)

// CreateClassKey constructs the key for the non-fungible token class.
func CreateClassKey(classID string) []byte {
	return store.JoinKeysWithLength(NFTClassKeyPrefix, []byte(classID))
}
