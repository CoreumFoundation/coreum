package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/bytesop"
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

var (
	// AssetKeyPrefix defines the store key prefix for the asset.
	AssetKeyPrefix = []byte{0x01}
	// FungibleTokenKeyPrefix defines the key prefix for the fungible token.
	FungibleTokenKeyPrefix = append(AssetKeyPrefix, 0x01)

	AirdropIDGeneratorKey = []byte{0x02}
	AirdropKeyPrefix      = []byte{0x03}
)

// GetFungibleTokenKey constructs the key for the fungible token.
func GetFungibleTokenKey(denom string) []byte {
	return bytesop.Join(FungibleTokenKeyPrefix, bytesop.WithLength([]byte(denom)))
}

func GetAirdropDenomPrefix(denom string) []byte {
	return bytesop.Join(AirdropKeyPrefix, bytesop.WithLength([]byte(denom)))
}

func GetAirdropKey(denom string, airdropID sdk.Int) []byte {
	return bytesop.Join(GetAirdropDenomPrefix(denom), bytesop.WithLength(must.Bytes(airdropID.Marshal())))
}
