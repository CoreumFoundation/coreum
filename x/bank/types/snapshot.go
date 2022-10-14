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

var _ snapshottypes.Transformation = SnapshotTransformation{}

var (
	denomsCode = []byte{0x00}
)

func SnapshotName(denom string) []byte {
	return store.JoinKeysWithLength(denomsCode, []byte(denom))
}

type SnapshotTransformation struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

func NewSnapshotTransformation(
	cdc codec.BinaryCodec,
	bankKey sdk.StoreKey,
) SnapshotTransformation {
	return SnapshotTransformation{
		cdc:      cdc,
		storeKey: bankKey,
	}
}

func (t SnapshotTransformation) StoreKey() sdk.StoreKey {
	return t.storeKey
}

func (t SnapshotTransformation) Transform(key, value []byte, deleted bool) ([]byte, snapshottypes.KeyValuePairs) {
	if !bytes.HasPrefix(key, banktypes.BalancesPrefix) {
		return nil, nil
	}
	accAddress, denom := decodeAddressDenom(key)
	return SnapshotName(denom), snapshottypes.KeyValuePairs{
		{
			Key:    address.MustLengthPrefix(accAddress),
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
