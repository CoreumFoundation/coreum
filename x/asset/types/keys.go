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
	// AssetKeyPrefix defines the store key prefix for the asset.
	AssetKeyPrefix = []byte{0x01}
	// FungibleTokenKeyPrefix defines the key prefix for the fungible token.
	FungibleTokenKeyPrefix = append(AssetKeyPrefix, 0x01)
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return append(FungibleTokenKeyPrefix, []byte(denom)...)
}
