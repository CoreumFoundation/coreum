package store

import (
	"io"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

var _ storetypes.CommitKVStore = kvStore{}

func newKVStore(parent storetypes.KVStore, snapshotKeySubstore SnapshotKeyStore, mappings []types.Mapping) kvStore {
	parentCommit, _ := parent.(storetypes.CommitKVStore)
	return kvStore{
		parent:              parent,
		parentCommit:        parentCommit,
		snapshotKeySubstore: snapshotKeySubstore,
		mappings:            mappings,
	}
}

type kvStore struct {
	parent              storetypes.KVStore
	parentCommit        storetypes.CommitKVStore
	snapshotKeySubstore SnapshotKeyStore
	mappings            []types.Mapping
}

func (s kvStore) Commit() storetypes.CommitID {
	return s.parentCommit.Commit()
}

func (s kvStore) LastCommitID() storetypes.CommitID {
	return s.parentCommit.LastCommitID()
}

func (s kvStore) SetPruning(options storetypes.PruningOptions) {
	s.parentCommit.SetPruning(options)
}

func (s kvStore) GetPruning() storetypes.PruningOptions {
	return s.parentCommit.GetPruning()
}

func (s kvStore) Get(key []byte) []byte {
	return s.parent.Get(key)
}

func (s kvStore) Has(key []byte) bool {
	return s.parent.Has(key)
}

func (s kvStore) Set(key, value []byte) {
	s.parent.Set(key, value)
	s.onWrite(key, value, false)
}

func (s kvStore) Delete(key []byte) {
	s.parent.Delete(key)
	s.onWrite(key, nil, true)
}

func (s kvStore) Iterator(start, end []byte) storetypes.Iterator {
	return s.parent.Iterator(start, end)
}

func (s kvStore) ReverseIterator(start, end []byte) storetypes.Iterator {
	return s.parent.ReverseIterator(start, end)
}

func (s kvStore) GetStoreType() storetypes.StoreType {
	return s.parent.GetStoreType()
}

// CacheWrap implements CacheWrapper.
func (s kvStore) CacheWrap() storetypes.CacheWrap {
	return cachekv.NewStore(s)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (s kvStore) CacheWrapWithTrace(w io.Writer, tc storetypes.TraceContext) storetypes.CacheWrap {
	return cachekv.NewStore(tracekv.NewStore(s, w, tc))
}

// CacheWrapWithListeners implements the CacheWrapper interface.
func (s kvStore) CacheWrapWithListeners(storeKey storetypes.StoreKey, listeners []storetypes.WriteListener) storetypes.CacheWrap {
	return cachekv.NewStore(listenkv.NewStore(s, storeKey, listeners))
}

func (s kvStore) onWrite(key, value []byte, deleted bool) {
	for _, t := range s.mappings {
		snapshotName, keyValuePairs := t.Map(key, value, deleted)
		if len(snapshotName) == 0 || len(keyValuePairs) == 0 {
			return
		}

		store := s.snapshotKeySubstore.ByName(snapshotName)
		for _, kv := range keyValuePairs {
			if kv.Delete {
				store.Delete(kv.Key)
			} else {
				store.Set(kv.Key, kv.Value)
			}
		}
	}
}
