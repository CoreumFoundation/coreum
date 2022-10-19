package types

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

var (
	// FungibleTokenKeyPrefix defines the key prefix for the fungible token.
	FungibleTokenKeyPrefix = []byte{0x01}
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return JoinKeys(FungibleTokenKeyPrefix, []byte(denom))
}

// JoinKeys joins the keys protecting the prefixes from the modification.
func JoinKeys(keys ...[]byte) []byte {
	compositeKey := make([]byte, 0)
	for _, key := range keys {
		compositeKey = append(compositeKey, key...)
	}

	return compositeKey
}
