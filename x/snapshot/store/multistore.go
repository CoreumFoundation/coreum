package store

import (
	"io"

	snapshotstypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protoio "github.com/gogo/protobuf/io"
	dbm "github.com/tendermint/tm-db"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type MultiStore struct {
	parent       storetypes.MultiStore
	parentCommit storetypes.CommitMultiStore
	snapshotKey  storetypes.StoreKey
	mappings     map[storetypes.StoreKey][]types.Mapping
}

var _ storetypes.CommitMultiStore = MultiStore{}

func New(parent storetypes.MultiStore, snapshotKey storetypes.StoreKey, mappings []types.Mapping) MultiStore {
	ms := map[sdk.StoreKey][]types.Mapping{}
	for _, t := range mappings {
		ms[t.StoreKey()] = append(ms[t.StoreKey()], t)
	}
	return newWithMappings(parent, snapshotKey, ms)
}

func newWithMappings(parent storetypes.MultiStore, snapshotKey storetypes.StoreKey, mappings map[storetypes.StoreKey][]types.Mapping) MultiStore {
	parentCommit, _ := parent.(storetypes.CommitMultiStore)
	return MultiStore{
		parent:       parent,
		parentCommit: parentCommit,
		snapshotKey:  snapshotKey,
		mappings:     mappings,
	}
}

func (cms MultiStore) MountStoreWithDB(key storetypes.StoreKey, typ storetypes.StoreType, db dbm.DB) {
	cms.parentCommit.MountStoreWithDB(key, typ, db)
}

func (cms MultiStore) GetCommitStore(key storetypes.StoreKey) storetypes.CommitStore {
	return cms.GetCommitKVStore(key)
}

func (cms MultiStore) GetCommitKVStore(key storetypes.StoreKey) storetypes.CommitKVStore {
	store := cms.parentCommit.GetCommitKVStore(key)
	if len(cms.mappings[key]) == 0 {
		return store
	}
	return newKVStore(store, NewSnapshotKeyStore(cms.parent.GetKVStore(cms.snapshotKey), key.Name()), cms.mappings[key])
}

func (cms MultiStore) LoadLatestVersion() error {
	return cms.parentCommit.LoadLatestVersion()
}

func (cms MultiStore) LoadLatestVersionAndUpgrade(upgrades *storetypes.StoreUpgrades) error {
	return cms.parentCommit.LoadLatestVersionAndUpgrade(upgrades)
}

func (cms MultiStore) LoadVersionAndUpgrade(ver int64, upgrades *storetypes.StoreUpgrades) error {
	return cms.parentCommit.LoadVersionAndUpgrade(ver, upgrades)
}

func (cms MultiStore) LoadVersion(ver int64) error {
	return cms.parentCommit.LoadVersion(ver)
}

func (cms MultiStore) SetInterBlockCache(cache storetypes.MultiStorePersistentCache) {
	cms.parentCommit.SetInterBlockCache(cache)
}

func (cms MultiStore) SetInitialVersion(version int64) error {
	return cms.parentCommit.SetInitialVersion(version)
}

func (cms MultiStore) SetIAVLCacheSize(size int) {
	cms.parentCommit.SetIAVLCacheSize(size)
}

func (cms MultiStore) SetIAVLDisableFastNode(disable bool) {
	cms.parentCommit.SetIAVLDisableFastNode(disable)
}

func (cms MultiStore) RollbackToVersion(version int64) error {
	return cms.RollbackToVersion(version)
}

func (cms MultiStore) Commit() storetypes.CommitID {
	return cms.parentCommit.Commit()
}

func (cms MultiStore) LastCommitID() storetypes.CommitID {
	return cms.parentCommit.LastCommitID()
}

func (cms MultiStore) SetPruning(options storetypes.PruningOptions) {
	cms.parentCommit.SetPruning(options)
}

func (cms MultiStore) GetPruning() storetypes.PruningOptions {
	return cms.parentCommit.GetPruning()
}

func (cms MultiStore) Snapshot(height uint64, protoWriter protoio.Writer) error {
	return cms.parentCommit.Snapshot(height, protoWriter)
}

func (cms MultiStore) Restore(height uint64, format uint32, protoReader protoio.Reader) (snapshotstypes.SnapshotItem, error) {
	return cms.parentCommit.Restore(height, format, protoReader)
}

func (cms MultiStore) SetTracer(w io.Writer) storetypes.MultiStore {
	cms.parent = cms.parent.SetTracer(w)
	return cms
}

func (cms MultiStore) SetTracingContext(tc storetypes.TraceContext) storetypes.MultiStore {
	cms.parent = cms.parent.SetTracingContext(tc)
	return cms
}

func (cms MultiStore) TracingEnabled() bool {
	return cms.parent.TracingEnabled()
}

func (cms MultiStore) AddListeners(key storetypes.StoreKey, listeners []storetypes.WriteListener) {
	cms.parent.AddListeners(key, listeners)
}

func (cms MultiStore) ListeningEnabled(key storetypes.StoreKey) bool {
	return cms.parent.ListeningEnabled(key)
}

func (cms MultiStore) GetStoreType() storetypes.StoreType {
	return cms.parent.GetStoreType()
}

func (cms MultiStore) Write() {
	cms.parent.(storetypes.CacheMultiStore).Write()
}

func (cms MultiStore) CacheWrap() storetypes.CacheWrap {
	return cms.CacheMultiStore().(storetypes.CacheWrap)
}

func (cms MultiStore) CacheWrapWithTrace(_ io.Writer, _ storetypes.TraceContext) storetypes.CacheWrap {
	return cms.CacheWrap()
}

func (cms MultiStore) CacheWrapWithListeners(writer storetypes.StoreKey, listeners []storetypes.WriteListener) storetypes.CacheWrap {
	return cms.CacheWrap()
}

func (cms MultiStore) CacheMultiStore() storetypes.CacheMultiStore {
	return newWithMappings(cms.parent.CacheMultiStore(), cms.snapshotKey, cms.mappings)
}

func (cms MultiStore) CacheMultiStoreWithVersion(version int64) (storetypes.CacheMultiStore, error) {
	substore, err := cms.parent.CacheMultiStoreWithVersion(version)
	if err != nil {
		return nil, err
	}
	return newWithMappings(substore, cms.snapshotKey, cms.mappings), nil
}

func (cms MultiStore) GetStore(key storetypes.StoreKey) storetypes.Store {
	return cms.GetKVStore(key)

}

func (cms MultiStore) GetKVStore(key storetypes.StoreKey) storetypes.KVStore {
	store := cms.parent.GetKVStore(key)
	if len(cms.mappings[key]) == 0 {
		return store
	}
	return newKVStore(store, NewSnapshotKeyStore(cms.parent.GetKVStore(cms.snapshotKey), key.Name()), cms.mappings[key])
}
