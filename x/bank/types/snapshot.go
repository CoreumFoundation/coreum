package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/pkg/store"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

var _ snapshottypes.Mapping = SnapshotMapping{}

var (
	denomsCode = []byte{0x00}
)

func SnapshotName(denom string) []byte {
	return store.JoinKeysWithLength(denomsCode, []byte(denom))
}

func AccountKey(accAddress sdk.AccAddress) []byte {
	return address.MustLengthPrefix(accAddress)
}

type SnapshotMapping struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

func NewSnapshotMapping(
	cdc codec.BinaryCodec,
	bankKey sdk.StoreKey,
) SnapshotMapping {
	return SnapshotMapping{
		cdc:      cdc,
		storeKey: bankKey,
	}
}

func (m SnapshotMapping) StoreKey() sdk.StoreKey {
	return m.storeKey
}

func (m SnapshotMapping) Map(key, value []byte, deleted bool) ([]byte, snapshottypes.KeyValuePairs) {
	if !bytes.HasPrefix(key, banktypes.BalancesPrefix) {
		return nil, nil
	}
	accAddress, denom := decodeAddressDenom(key)
	return SnapshotName(denom), snapshottypes.KeyValuePairs{
		{
			Key:    AccountKey(accAddress),
			Value:  value,
			Delete: deleted,
		},
	}
}

func decodeAddressDenom(key []byte) (sdk.AccAddress, string) {
	key = key[len(banktypes.BalancesPrefix):]
	addressLen := int(key[0])
	key = key[1:]
	return key[:addressLen], string(key[addressLen:])
}
