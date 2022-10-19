package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
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

func (k Keeper) RequestSnapshot(ctx sdk.Context, info types.SnapshotRequestInfo) (sdk.Int, error) {
	// FIXME (wojtek): validate request

	store := ctx.KVStore(k.storeKey)

	id := sdk.ZeroInt()
	bz := store.Get(types.IDGeneratorKey)
	if bz != nil {
		must.OK(id.Unmarshal(bz))
		id = id.Add(sdk.OneInt())
	}
	store.Set(types.IDGeneratorKey, must.Bytes(id.Marshal()))

	ownerAddress, err := sdk.AccAddressFromBech32(info.Owner)
	if err != nil {
		return sdk.Int{}, errors.WithStack(err)
	}

	request := &types.SnapshotRequest{
		Prefix:          info.Prefix,
		Id:              id,
		Owner:           info.Owner,
		Height:          info.Height,
		Description:     info.Description,
		UserDescription: info.UserDescription,
	}
	requestMarshaled := k.cdc.MustMarshal(request)
	store.Set(types.GetPendingByAccountSnapshotKey(ownerAddress, id), requestMarshaled)
	store.Set(types.GetPendingByBlockSnapshotKey(request.Height, id), requestMarshaled)
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
			Key: types.SnapshotKey{
				Prefix: request.Prefix,
				Index:  snapshotIndex,
			},
			Id:              request.Id,
			Owner:           request.Owner,
			Height:          request.Height,
			Description:     request.Description,
			UserDescription: request.UserDescription,
		}

		owner := sdk.MustAccAddressFromBech32(request.Owner)
		store.Set(types.GetTakenByAccountSnapshotKey(owner, request.Id), k.cdc.MustMarshal(snapshot))
		store.Delete(types.GetPendingByAccountSnapshotKey(owner, request.Id))
		requestStore.Delete(iterator.Key())
	}
}

func (k Keeper) GetPending(ctx sdk.Context, accAddress sdk.AccAddress) []types.SnapshotRequest {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetPendingByAccountPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.SnapshotRequest
	for ; iterator.Valid(); iterator.Next() {
		var request types.SnapshotRequest
		k.cdc.MustUnmarshal(iterator.Value(), &request)
		res = append(res, request)
	}

	return res
}

func (k Keeper) GetSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) []types.Snapshot {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetTakenByAccountPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.Snapshot
	for ; iterator.Valid(); iterator.Next() {
		var snapshot types.Snapshot
		k.cdc.MustUnmarshal(iterator.Value(), &snapshot)
		res = append(res, snapshot)
	}

	return res
}
