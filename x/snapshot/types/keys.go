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

	IDGeneratorKey     = []byte{0x01}
	PendingRequestsKey = []byte{0x02}
	TakenKey           = []byte{0x03}
)

func AccountPendingSnapshotsPrefix(accAddress sdk.AccAddress) []byte {
	return Join(PendingRequestsKey, address.MustLengthPrefix(accAddress))
}

func AccountPendingSnapshotKey(accAddress sdk.AccAddress, index sdk.Int) []byte {
	return Join(AccountPendingSnapshotsPrefix(accAddress), must.Bytes(index.Marshal()))
}

func AccountTakenSnapshotsPrefix(accAddress sdk.AccAddress) []byte {
	return Join(TakenKey, address.MustLengthPrefix(accAddress))
}

func AccountTakenSnapshotKey(accAddress sdk.AccAddress, index sdk.Int) []byte {
	return Join(AccountTakenSnapshotsPrefix(accAddress), must.Bytes(index.Marshal()))
}

func SnapshotDataPrefix(index sdk.Int) []byte {
	return Join(DataSubkey, must.Bytes(index.Marshal()))
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
