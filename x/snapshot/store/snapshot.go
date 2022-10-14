package store

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

func NewSnapshotKeyStore(snapshotStore storetypes.KVStore, storeName string) SnapshotKeyStore {
	return SnapshotKeyStore{
		store: prefix.NewStore(snapshotStore, types.GetStoreSnapshotsPrefix(storeName)),
	}
}

type SnapshotKeyStore struct {
	store storetypes.KVStore
}

func (s SnapshotKeyStore) ByName(shanpshotName []byte) SnapshotNameStore {
	return SnapshotNameStore{
		store: prefix.NewStore(s.store, shanpshotName),
	}
}

type SnapshotNameStore struct {
	store storetypes.KVStore
}

func (s SnapshotNameStore) Get(key []byte) []byte {
	snapshot, exists := s.latestSnapshotByKey(key)
	if !exists {
		return nil
	}
	return snapshot.Get(key)
}

func (s SnapshotNameStore) Has(key []byte) bool {
	_, exists := s.latestSnapshotByKey(key)
	return exists
}

func (s SnapshotNameStore) Set(key, value []byte) {
	snapshot := s.pendingSnapshot()
	s.fillSnapshot(snapshot.Index(), key)
	snapshot.set(key, value)
}

func (s SnapshotNameStore) Delete(key []byte) {
	snapshot := s.pendingSnapshot()
	if snapshot.Index() > 0 {
		s.fillSnapshot(snapshot.Index()-1, key)
	}
	snapshot.delete(key)
	s.deleteIndex(key)
}

func (s SnapshotNameStore) ClaimFromSnapshot(index uint64, key []byte) error {
	snapshot, exists := s.ByIndex(index)
	if !exists {
		return errors.Errorf("snapshot with index %s does not exist", index)
	}
	if snapshot.IsPending() {
		return errors.Errorf("snapshot with index %s is a pending one", index)
	}
	s.fillSnapshot(snapshot.Index()+1, key)
	if !snapshot.Has(key) {
		return errors.New("key does not exist in snapshot")
	}
	snapshot.delete(key)
	return nil
}

func (s SnapshotNameStore) GetFromSnapshot(index uint64, key []byte) ([]byte, error) {
	if index > s.pendingIndex() {
		return nil, errors.Errorf("snapshot with index %s does not exist", index)
	}
	snapshot, exists := s.latestSnapshotByKey(key)
	if exists && snapshot.Index() <= index {
		return snapshot.Get(key), nil
	}

	snapshot, _ = s.ByIndex(index)
	return snapshot.Get(key), nil
}

func (s SnapshotNameStore) TakeSnapshot() uint64 {
	index := s.pendingIndex()
	s.store.Set(types.PendingSnapshotSubkeyPrefix, proto.EncodeVarint(index+1))
	return index
}

func (s SnapshotNameStore) ByIndex(index uint64) (SnapshotDataStore, bool) {
	pendingIndex := s.pendingIndex()
	if index > pendingIndex {
		return SnapshotDataStore{}, false
	}
	return NewSnapshotDataStore(index, index == pendingIndex, s.store), true
}

func (s SnapshotNameStore) pendingSnapshot() SnapshotDataStore {
	snapshot, _ := s.ByIndex(s.pendingIndex())
	return snapshot
}

func (s SnapshotNameStore) latestSnapshotByKey(key []byte) (SnapshotDataStore, bool) {
	var index uint64
	bz := s.store.Get(types.GetCurrentValueSnapshotIndexSubkey(key))
	if bz == nil {
		return SnapshotDataStore{}, false
	}
	index, _ = proto.DecodeVarint(bz)
	return s.ByIndex(index)
}

func (s SnapshotNameStore) pendingIndex() uint64 {
	var index uint64
	bz := s.store.Get(types.PendingSnapshotSubkeyPrefix)
	if bz != nil {
		index, _ = proto.DecodeVarint(bz)
	}
	return index
}

func (s SnapshotNameStore) deleteIndex(key []byte) {
	s.store.Delete(types.GetCurrentValueSnapshotIndexSubkey(key))
}

func (s SnapshotNameStore) setIndex(key []byte, index uint64) {
	s.store.Set(types.GetCurrentValueSnapshotIndexSubkey(key), proto.EncodeVarint(index))
}

func (s SnapshotNameStore) fillSnapshot(upToIndex uint64, key []byte) {
	currentValueSnapshotStore, exists := s.latestSnapshotByKey(key)
	if !exists || currentValueSnapshotStore.IsPending() {
		return
	}
	if upToIndex > currentValueSnapshotStore.Index() {
		currentValue := currentValueSnapshotStore.Get(key)
		for index := currentValueSnapshotStore.Index() + 1; index <= upToIndex; index++ {
			snapshotStore, _ := s.ByIndex(index)
			snapshotStore.set(key, currentValue)
		}
		s.setIndex(key, upToIndex)
	}
}

func NewSnapshotDataStore(index uint64, isPending bool, parentStore storetypes.KVStore) SnapshotDataStore {
	return SnapshotDataStore{
		index:     index,
		store:     prefix.NewStore(parentStore, types.GetSnapshotDataPrefix(index)),
		isPending: isPending,
	}
}

type SnapshotDataStore struct {
	index     uint64
	store     storetypes.KVStore
	isPending bool
}

func (s SnapshotDataStore) Index() uint64 {
	return s.index
}

func (s SnapshotDataStore) IsPending() bool {
	return s.isPending
}

func (s SnapshotDataStore) Get(key []byte) []byte {
	return s.store.Get(key)
}

func (s SnapshotDataStore) Has(key []byte) bool {
	return s.store.Has(key)
}

func (s SnapshotDataStore) set(key, value []byte) {
	s.store.Set(key, value)
}

func (s SnapshotDataStore) delete(key []byte) {
	s.store.Delete(key)
}
