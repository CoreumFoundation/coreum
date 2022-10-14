package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type KeyValuePair struct {
	Key    []byte
	Value  []byte
	Delete bool
}

type KeyValuePairs []KeyValuePair

type Transformation interface {
	StoreKey() sdk.StoreKey
	Transform(key, value []byte, deleted bool) ([]byte, KeyValuePairs)
}
