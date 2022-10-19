package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/pkg/bytesop"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

var _ snapshottypes.Transformation = BankTransformation{}

var (
	denomsCode      = []byte{0x00}
	balancesSubcode = []byte{0x00}
)

func BalancesSnapshotName(denom string) []byte {
	return bytesop.Join(denomsCode, bytesop.WithLength([]byte(denom)), balancesSubcode)
}

type BankTransformation struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

func NewBankTransformation(
	cdc codec.BinaryCodec,
	bankKey sdk.StoreKey,
) BankTransformation {
	return BankTransformation{
		cdc:      cdc,
		storeKey: bankKey,
	}
}

func (bt BankTransformation) StoreKey() sdk.StoreKey {
	return bt.storeKey
}

func (bt BankTransformation) Transform(key, value []byte, deleted bool) ([]byte, snapshottypes.KeyValuePairs) {
	if !bytes.HasPrefix(key, banktypes.BalancesPrefix) {
		return nil, nil
	}
	accAddress, denom := decodeAddressDenom(key)
	var balance sdk.Coin
	must.OK(balance.Unmarshal(value))
	return BalancesSnapshotName(denom), snapshottypes.KeyValuePairs{
		{
			Key:    address.MustLengthPrefix(accAddress),
			Value:  bt.cdc.MustMarshal(&AccountBalance{Balance: balance, Address: accAddress.String()}),
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
