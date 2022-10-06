package types

import sdk "github.com/cosmos/cosmos-sdk/types"

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
	// AssetSequenceKey defines the store key for the asset id.
	AssetSequenceKey = []byte{0x01}
	// AssetKeyPrefix defines the store key prefix for the asset.
	AssetKeyPrefix = []byte{0x02}
	// AssetFTKeyPrefix defines the key prefix to save the FT asset.
	AssetFTKeyPrefix = append(AssetKeyPrefix, 0x01)
)

// GetAssetFTKey constructs the key for the asset.
func GetAssetFTKey(id uint64) []byte {
	return append(AssetFTKeyPrefix, sdk.Uint64ToBigEndian(id)...)
}
