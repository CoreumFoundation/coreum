package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	snapshotstore "github.com/CoreumFoundation/coreum/x/snapshot/store"
	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}

func (k Keeper) RequestSnapshot(ctx sdk.Context, info types.SnapshotRequestInfo) (uint64, error) {
	// FIXME (wojtek): validate request

	store := ctx.KVStore(k.storeKey)

	var id uint64
	bz := store.Get(types.IDGeneratorKey)
	if bz != nil {
		id, _ = proto.DecodeVarint(bz)
		id++
	}
	store.Set(types.IDGeneratorKey, proto.EncodeVarint(id))

	ownerAddress, err := sdk.AccAddressFromBech32(info.Owner)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	request := &types.SnapshotRequest{
		Info: types.SnapshotInfo{
			Id:              id,
			Owner:           info.Owner,
			Height:          info.Height,
			Description:     info.Description,
			UserDescription: info.UserDescription,
		},
		Prefix: info.Prefix,
	}
	requestMarshaled := k.cdc.MustMarshal(request)
	store.Set(types.GetPendingByAccountSnapshotKey(ownerAddress, id), requestMarshaled)
	store.Set(types.GetPendingByBlockSnapshotKey(info.Height, id), requestMarshaled)
	return id, nil
}

func (k Keeper) TakeSnapshots(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	requestStore := prefix.NewStore(store, types.GetPendingByBlockPrefix(ctx.BlockHeight()))
	iterator := requestStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var request types.SnapshotRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)

		snapshotIndex := snapshotstore.NewSnapshotKeyStore(store, request.Prefix.StoreName).ByName(request.Prefix.Name).TakeSnapshot()
		snapshot := &types.Snapshot{
			Info: request.Info,
			Key: types.SnapshotKey{
				Prefix: request.Prefix,
				Index:  snapshotIndex,
			},
		}

		owner := sdk.MustAccAddressFromBech32(request.Info.Owner)
		store.Set(types.GetTakenByAccountSnapshotKey(owner, request.Info.Id), k.cdc.MustMarshal(snapshot))
		store.Delete(types.GetPendingByAccountSnapshotKey(owner, request.Info.Id))
		requestStore.Delete(iterator.Key())
	}
}

func (k Keeper) GetPending(ctx sdk.Context, accAddress sdk.AccAddress) []types.SnapshotInfo {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetPendingByAccountPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.SnapshotInfo
	for ; iterator.Valid(); iterator.Next() {
		var request types.SnapshotRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)
		res = append(res, request.Info)
	}

	return res
}

func (k Keeper) GetSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) []types.SnapshotInfo {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetTakenByAccountPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.SnapshotInfo
	for ; iterator.Valid(); iterator.Next() {
		var snapshot types.Snapshot
		k.cdc.MustUnmarshal(iterator.Value(), &snapshot)
		res = append(res, snapshot.Info)
	}

	return res
}

func (k Keeper) GetSnapshot(ctx sdk.Context, owner sdk.AccAddress, snapshotID uint64) (types.Snapshot, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTakenByAccountSnapshotKey(owner, snapshotID))
	if bz == nil {
		return types.Snapshot{}, errors.New("snapshot does not exist")
	}
	var snapshot types.Snapshot
	k.cdc.MustUnmarshal(bz, &snapshot)

	return snapshot, nil
}

func (k Keeper) GetValueFromSnapshot(ctx sdk.Context, snapshotKey types.SnapshotKey, key []byte) ([]byte, error) {
	return snapshotstore.NewSnapshotKeyStore(ctx.KVStore(k.storeKey), snapshotKey.Prefix.StoreName).ByName(snapshotKey.Prefix.Name).GetFromSnapshot(snapshotKey.Index, key)
}

func (k Keeper) ClaimFromSnapshot(ctx sdk.Context, snapshotKey types.SnapshotKey, key []byte) error {
	return snapshotstore.NewSnapshotKeyStore(ctx.KVStore(k.storeKey), snapshotKey.Prefix.StoreName).ByName(snapshotKey.Prefix.Name).ClaimFromSnapshot(snapshotKey.Index, key)
}
