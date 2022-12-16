package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// FungibleTokenSymbolKeyPrefix defines the key prefix for the fungible token by Symbol.
	FungibleTokenSymbolKeyPrefix = []byte{0x02}
	// FrozenBalancesKeyPrefix defines the key prefix to track frozen balances
	FrozenBalancesKeyPrefix = []byte{0x03}
	// GlobalFreezeKeyPrefix defines the key prefix to track global freezing of a Fungible Token.
	GlobalFreezeKeyPrefix = []byte{0x04}
	// WhitelistedBalancesKeyPrefix defines the key prefix to track whitelisted balances
	WhitelistedBalancesKeyPrefix = []byte{0x05}
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return store.JoinKeysWithLength(FungibleTokenKeyPrefix, []byte(denom))
}

// CreateFrozenBalancesPrefix creates the prefix for an account's frozen balances.
func CreateFrozenBalancesPrefix(addr []byte) []byte {
	return store.JoinKeys(FrozenBalancesKeyPrefix, address.MustLengthPrefix(addr))
}

// CreateGlobalFreezePrefix creates the prefix for fungible token global freeze key.
func CreateGlobalFreezePrefix(denom string) []byte {
	return store.JoinKeys(GlobalFreezeKeyPrefix, []byte(denom))
}

// CreateWhitelistedBalancesPrefix creates the prefix for an account's whitelisted balances.
func CreateWhitelistedBalancesPrefix(addr []byte) []byte {
	return store.JoinKeys(WhitelistedBalancesKeyPrefix, address.MustLengthPrefix(addr))
}

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the prefix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
//
// If invalid key is passed, AddressFromBalancesStore returns ErrInvalidKey.
func AddressFromBalancesStore(key []byte) (sdk.AccAddress, error) {
	if len(key) == 0 {
		return nil, ErrInvalidKey
	}
	addrLen := key[0]
	bound := int(addrLen)
	if len(key)-1 < bound {
		return nil, ErrInvalidKey
	}
	return key[1 : bound+1], nil
}

// CreateSymbolPrefix creates the prefix for an ft symbol.
func CreateSymbolPrefix(addr []byte) []byte {
	return store.JoinKeys(FungibleTokenSymbolKeyPrefix, addr)
}
