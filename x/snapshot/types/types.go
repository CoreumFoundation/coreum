package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Mapping interface {
	StoreKey() sdk.StoreKey
	Map(key, value []byte, deleted bool) ([]byte, KeyValuePairs)
}

type KeyValuePair struct {
	Key    []byte
	Value  []byte
	Delete bool
}

type KeyValuePairs []KeyValuePair

type SnapshotRequestInfo struct {
	Prefix          SnapshotPrefix
	Owner           string
	Height          int64
	Description     string
	UserDescription string
}
