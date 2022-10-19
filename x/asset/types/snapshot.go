package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

var _ snapshottypes.Transformation = BankTransformation{}

var (
	denomsCode      = []byte{0x00}
	balancesSubcode = []byte{0x00}
)

func BalancesSnapshotName(denom string) []byte {
	return snapshottypes.Join(denomsCode, []byte(denom), balancesSubcode)
}

type BankTransformation struct {
	storeKey sdk.StoreKey
}

func NewBankTransformation(bankKey sdk.StoreKey) BankTransformation {
	return BankTransformation{
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
			Value:  must.Bytes((&AccountBalance{Balance: balance, Address: accAddress.String()}).Marshal()),
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

type SnapshotRequestFungibleToken struct {
	Denom       string
	Owner       string
	Height      int64
	Name        string
	Description string
}
