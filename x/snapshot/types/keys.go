package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/bytesop"
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
	return bytesop.Join(PendingByAccountKeyPrefix, address.MustLengthPrefix(accAddress))
}

func GetPendingByBlockPrefix(height int64) []byte {
	return bytesop.Join(PendingByBlockKeyPrefix, must.Bytes(sdk.NewInt(height).Marshal()))
}

func GetPendingByAccountSnapshotKey(accAddress sdk.AccAddress, snapshotID sdk.Int) []byte {
	return bytesop.Join(GetPendingByAccountPrefix(accAddress), bytesop.WithLength(must.Bytes(snapshotID.Marshal())))
}

func GetPendingByBlockSnapshotKey(height int64, snapshotID sdk.Int) []byte {
	return bytesop.Join(GetPendingByBlockPrefix(height), bytesop.WithLength(must.Bytes(snapshotID.Marshal())))
}

func GetTakenByAccountPrefix(accAddress sdk.AccAddress) []byte {
	return bytesop.Join(TakenByAccountKeyPrefix, address.MustLengthPrefix(accAddress))
}

func GetTakenByAccountSnapshotKey(accAddress sdk.AccAddress, snapshotID sdk.Int) []byte {
	return bytesop.Join(GetTakenByAccountPrefix(accAddress), bytesop.WithLength(must.Bytes(snapshotID.Marshal())))
}

func GetSnapshotDataPrefix(snapshotIndex sdk.Int) []byte {
	return bytesop.Join(DataSubkeyPrefix, bytesop.WithLength(must.Bytes(snapshotIndex.Marshal())))
}

func GetStoreSnapshotsPrefix(storeName string) []byte {
	return bytesop.Join(SnapshotKeyPrefix, []byte(storeName))
}

func GetCurrentValueSnapshotIndexSubkey(key []byte) []byte {
	return bytesop.Join(CurrentValueSnapshotIndexSubkeyPrefix, key)
}
