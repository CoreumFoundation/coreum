package store

import (
	"io"

	"github.com/cosmos/cosmos-sdk/store/cachekv"
	"github.com/cosmos/cosmos-sdk/store/listenkv"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/tracekv"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

// Store holds many branched stores.
// Implements MultiStore.
// NOTE: a Store (and MultiStores in general) should never expose the
// keys for the substores.
type Store struct {
	parent          storetypes.MultiStore
	snapshotKey     storetypes.StoreKey
	snapshotStore   storetypes.KVStore
	transformations map[storetypes.StoreKey][]types.Transformation
}

var _ storetypes.MultiStore = Store{}

// New creates a new multi store wrapper for snapshotting
func New(parent storetypes.MultiStore, snapshotKey storetypes.StoreKey, transformations map[storetypes.StoreKey][]types.Transformation) Store {
	return Store{
		parent:          parent,
		snapshotKey:     snapshotKey,
		snapshotStore:   parent.GetKVStore(snapshotKey),
		transformations: transformations,
	}
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (cms Store) SetTracer(w io.Writer) storetypes.MultiStore {
	cms.parent = cms.parent.SetTracer(w)
	return cms
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (cms Store) SetTracingContext(tc storetypes.TraceContext) storetypes.MultiStore {
	cms.parent = cms.parent.SetTracingContext(tc)
	return cms
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (cms Store) TracingEnabled() bool {
	return cms.parent.TracingEnabled()
}

// AddListeners adds listeners for a specific KVStore
func (cms Store) AddListeners(key storetypes.StoreKey, listeners []storetypes.WriteListener) {
	cms.parent.AddListeners(key, listeners)
}

// ListeningEnabled returns if listening is enabled for a specific KVStore
func (cms Store) ListeningEnabled(key storetypes.StoreKey) bool {
	return cms.parent.ListeningEnabled(key)
}

// GetStoreType returns the type of the store.
func (cms Store) GetStoreType() storetypes.StoreType {
	return cms.parent.GetStoreType()
}

// Write calls Write on each underlying store.
func (cms Store) Write() {
	cms.parent.(storetypes.CacheMultiStore).Write()
}

// Implements CacheWrapper.
func (cms Store) CacheWrap() storetypes.CacheWrap {
	return cms.CacheMultiStore().(storetypes.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (cms Store) CacheWrapWithTrace(_ io.Writer, _ storetypes.TraceContext) storetypes.CacheWrap {
	return cms.CacheWrap()
}

// CacheWrapWithListeners implements the CacheWrapper interface.
func (cms Store) CacheWrapWithListeners(_ storetypes.StoreKey, _ []storetypes.WriteListener) storetypes.CacheWrap {
	return cms.CacheWrap()
}

// Implements MultiStore.
func (cms Store) CacheMultiStore() storetypes.CacheMultiStore {
	return New(cms.parent.CacheMultiStore(), cms.snapshotKey, cms.transformations)
}

// CacheMultiStoreWithVersion implements the MultiStore interface. It will panic
// as an already cached multi-store cannot load previous versions.
//
// TODO: The store implementation can possibly be modified to support this as it
// seems safe to load previous versions (heights).
func (cms Store) CacheMultiStoreWithVersion(_ int64) (storetypes.CacheMultiStore, error) {
	panic("cannot branch cached multi-store with a version")
}

// GetStore returns an underlying Store by key.
func (cms Store) GetStore(key storetypes.StoreKey) storetypes.Store {
	return cms.GetKVStore(key)

}

// GetKVStore returns an underlying KVStore by key.
func (cms Store) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	return newKVStore(cms.parent.GetKVStore(key), NewSnapshotKeyStore(cms.snapshotStore, key.Name()), cms.transformations[key])
}

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
	s.fillSnapshot(key)
	snapshot := s.pendingSnapshot()
	snapshot.set(key, value)
	s.setIndex(key, snapshot.Index())
}

func (s SnapshotNameStore) Delete(key []byte) {
	s.fillSnapshot(key)
	s.pendingSnapshot().delete(key)
	s.deleteIndex(key)
}

func (s SnapshotNameStore) Claim(index sdk.Int, key []byte) error {
	s.fillSnapshot(key)
	snapshot, exists := s.ByIndex(index)
	if !exists {
		return errors.Errorf("snapshot with index %s does not exist", index)
	}
	if snapshot.IsPending() {
		return errors.Errorf("snapshot with index %s is a pending one", index)
	}
	snapshot.delete(key)
	return nil
}

func (s SnapshotNameStore) TakeSnapshot() sdk.Int {
	index := s.pendingIndex()
	s.store.Set(types.PendingSnapshotSubkeyPrefix, must.Bytes(index.Add(sdk.OneInt()).Marshal()))
	return index
}

func (s SnapshotNameStore) ByIndex(index sdk.Int) (SnapshotDataStore, bool) {
	pendingIndex := s.pendingIndex()
	if index.GT(pendingIndex) {
		return SnapshotDataStore{}, false
	}
	return NewSnapshotDataStore(index, index.Equal(pendingIndex), s.store), true
}

func (s SnapshotNameStore) pendingSnapshot() SnapshotDataStore {
	snapshot, _ := s.ByIndex(s.pendingIndex())
	return snapshot
}

func (s SnapshotNameStore) latestSnapshotByKey(key []byte) (SnapshotDataStore, bool) {
	var index sdk.Int
	bz := s.store.Get(types.GetCurrentValueSnapshotIndexSubkey(key))
	if bz == nil {
		return SnapshotDataStore{}, false
	}
	if err := index.Unmarshal(bz); err != nil {
		panic(err)
	}

	return s.ByIndex(index)
}

func (s SnapshotNameStore) pendingIndex() sdk.Int {
	index := sdk.ZeroInt()
	bz := s.store.Get(types.PendingSnapshotSubkeyPrefix)
	if bz != nil {
		if err := index.Unmarshal(bz); err != nil {
			panic(err)
		}
	}
	return index
}

func (s SnapshotNameStore) deleteIndex(key []byte) {
	s.store.Delete(types.GetCurrentValueSnapshotIndexSubkey(key))
}

func (s SnapshotNameStore) setIndex(key []byte, index sdk.Int) {
	s.store.Set(types.GetCurrentValueSnapshotIndexSubkey(key), must.Bytes(index.Marshal()))
}

func (s SnapshotNameStore) fillSnapshot(key []byte) {
	currentValueSnapshotStore, exists := s.latestSnapshotByKey(key)
	if !exists || currentValueSnapshotStore.IsPending() {
		return
	}
	pendingSnapshotStore := s.pendingSnapshot()
	currentValue := currentValueSnapshotStore.Get(key)
	for index := currentValueSnapshotStore.Index().Add(sdk.OneInt()); index.LTE(pendingSnapshotStore.Index()); index = index.Add(sdk.OneInt()) {
		snapshotStore, _ := s.ByIndex(index)
		snapshotStore.set(key, currentValue)
	}
	s.setIndex(key, pendingSnapshotStore.Index())
}

func NewSnapshotDataStore(index sdk.Int, isPending bool, parentStore storetypes.KVStore) SnapshotDataStore {
	return SnapshotDataStore{
		index:     index,
		store:     prefix.NewStore(parentStore, types.GetSnapshotDataPrefix(index)),
		isPending: isPending,
	}
}

type SnapshotDataStore struct {
	index     sdk.Int
	store     storetypes.KVStore
	isPending bool
}

func (s SnapshotDataStore) Index() sdk.Int {
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

func newKVStore(parent storetypes.KVStore, snapshotKeySubstore SnapshotKeyStore, transformations []types.Transformation) kvStore {
	return kvStore{
		parent:              parent,
		snapshotKeySubstore: snapshotKeySubstore,
		transformations:     transformations,
	}
}

type kvStore struct {
	parent              storetypes.KVStore
	snapshotKeySubstore SnapshotKeyStore
	transformations     []types.Transformation
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
	for _, t := range s.transformations {
		snapshotName, keyValuePairs := t.Transform(key, value, deleted)
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
