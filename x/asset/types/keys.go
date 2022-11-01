package types

import (
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

const (
	// ModuleName defines the module name
	ModuleName = "asset"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// FungibleTokenKeyPrefix defines the key prefix for the fungible token.
	FungibleTokenKeyPrefix = []byte{0x01}
	// FrozenBalancesPrefix defines the key prefix to track frozen balances
	FrozenBalancesPrefix = []byte{0x02}
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return store.JoinKeysWithLength(FungibleTokenKeyPrefix, []byte(denom))
}

// CreateFrozenBalancesPrefix creates the prefix for an account's balances.
func CreateFrozenBalancesPrefix(addr []byte) []byte {
	return store.JoinKeys(FrozenBalancesPrefix, address.MustLengthPrefix(addr))
}
