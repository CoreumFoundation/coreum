package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/CoreumFoundation/coreum/pkg/store"
)

const (
	// ModuleName defines the module name
	ModuleName = "snapshot"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	SnapshotKeyPrefix                     = []byte{0x00}
	PendingSnapshotSubkeyPrefix           = []byte{0x00}
	CurrentValueSnapshotIndexSubkeyPrefix = []byte{0x01}
	DataSubkeyPrefix                      = []byte{0x02}

	IDGeneratorKey            = []byte{0x01}
	PendingByAccountKeyPrefix = []byte{0x02}
	PendingByBlockKeyPrefix   = []byte{0x03}
	TakenByAccountKeyPrefix   = []byte{0x04}
)

func GetPendingByAccountPrefix(accAddress sdk.AccAddress) []byte {
	return store.JoinKeysWithLength(PendingByAccountKeyPrefix, accAddress)
}

func GetPendingByBlockPrefix(height int64) []byte {
	return store.JoinKeys(PendingByBlockKeyPrefix, proto.EncodeVarint(uint64(height)))
}

func GetPendingByAccountSnapshotKey(accAddress sdk.AccAddress, snapshotID uint64) []byte {
	return store.JoinKeys(GetPendingByAccountPrefix(accAddress), proto.EncodeVarint(snapshotID))
}

func GetPendingByBlockSnapshotKey(height int64, snapshotID uint64) []byte {
	return store.JoinKeys(GetPendingByBlockPrefix(height), proto.EncodeVarint(snapshotID))
}

func GetTakenByAccountPrefix(accAddress sdk.AccAddress) []byte {
	return store.JoinKeysWithLength(TakenByAccountKeyPrefix, accAddress)
}

func GetTakenByAccountSnapshotKey(accAddress sdk.AccAddress, snapshotID uint64) []byte {
	return store.JoinKeys(GetTakenByAccountPrefix(accAddress), proto.EncodeVarint(snapshotID))
}

func GetSnapshotDataPrefix(snapshotIndex uint64) []byte {
	return store.JoinKeys(DataSubkeyPrefix, proto.EncodeVarint(snapshotIndex))
}

func GetStoreSnapshotsPrefix(storeName string) []byte {
	return store.JoinKeysWithLength(SnapshotKeyPrefix, []byte(storeName))
}

func GetCurrentValueSnapshotIndexSubkey(key []byte) []byte {
	return store.JoinKeys(CurrentValueSnapshotIndexSubkeyPrefix, key)
}
