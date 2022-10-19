package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
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
	SnapshotsKey                    = []byte{0x00}
	PendingSnapshotSubkey           = []byte{0x00}
	CurrentValueSnapshotIndexSubkey = []byte{0x01}
	DataSubkey                      = []byte{0x02}

	IDGeneratorKey      = []byte{0x01}
	PendingByAccountKey = []byte{0x02}
	PendingByBlockKey   = []byte{0x03}
	TakenByAccountKey   = []byte{0x04}
)

func PendingByAccountPrefix(accAddress sdk.AccAddress) []byte {
	return Join(PendingByAccountKey, address.MustLengthPrefix(accAddress))
}

func PendingByBlockPrefix(height int64) []byte {
	return Join(PendingByBlockKey, must.Bytes(sdk.NewInt(height).Marshal()))
}

func AccountPendingSnapshotKey(accAddress sdk.AccAddress, snapshotID sdk.Int) []byte {
	return Join(PendingByAccountPrefix(accAddress), must.Bytes(snapshotID.Marshal()))
}

func BlockPendingSnapshotKey(height int64, snapshotID sdk.Int) []byte {
	return Join(PendingByBlockPrefix(height), must.Bytes(snapshotID.Marshal()))
}

func TakenByAccountPrefix(accAddress sdk.AccAddress) []byte {
	return Join(TakenByAccountKey, address.MustLengthPrefix(accAddress))
}

func AccountTakenSnapshotKey(accAddress sdk.AccAddress, snapshotID sdk.Int) []byte {
	return Join(TakenByAccountPrefix(accAddress), must.Bytes(snapshotID.Marshal()))
}

func SnapshotDataPrefix(snapshotIndex sdk.Int) []byte {
	return Join(DataSubkey, must.Bytes(snapshotIndex.Marshal()))
}

func StoreSnapshotsPrefix(storeName string) []byte {
	return Join(SnapshotsKey, []byte(storeName))
}

func SubkeyForCurrentValueSnapshotIndex(key []byte) []byte {
	return Join(CurrentValueSnapshotIndexSubkey, key)
}

func Join(keyComponents ...[]byte) []byte {
	var totalLength int
	for _, v := range keyComponents {
		totalLength += len(v)
	}

	res := make([]byte, 0, totalLength)
	for _, v := range keyComponents {
		res = append(res, v...)
	}
	return res
}
