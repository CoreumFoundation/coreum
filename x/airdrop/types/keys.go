package types

import (
	"github.com/gogo/protobuf/proto"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

const (
	// ModuleName defines the module name
	ModuleName = "airdrop"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	AirdropKeyPrefix = []byte{0x00}
)

func GetAirdropDenomPrefix(denom string) []byte {
	return store.JoinKeysWithLength(AirdropKeyPrefix, []byte(denom))
}

func GetAirdropKey(denom string, airdropID uint64) []byte {
	return store.JoinKeysWithLength(GetAirdropDenomPrefix(denom), proto.EncodeVarint(airdropID))
}
