package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	snapshotstore "github.com/CoreumFoundation/coreum/x/snapshot/store"
	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
}

func NewKeeper(storeKey sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: storeKey,
	}
}

func (k Keeper) RequestSnapshot(ctx sdk.Context, info types.SnapshotRequestInfo) error {
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
		return errors.WithStack(err)
	}

	request := types.SnapshotRequest{
		Prefix:          info.Prefix,
		Id:              id,
		Owner:           info.Owner,
		Height:          info.Height,
		Description:     info.Description,
		UserDescription: info.UserDescription,
	}
	store.Set(types.AccountPendingSnapshotKey(ownerAddress, id), must.Bytes(request.Marshal()))
	return nil
}

func (k Keeper) TakeSnapshots(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	requestStore := prefix.NewStore(store, types.PendingRequestsKey)
	frozenStore := prefix.NewStore(store, types.TakenKey)
	iterator := requestStore.Iterator(nil, nil)
	defer iterator.Close()

	height := ctx.BlockHeight()
	for ; iterator.Valid(); iterator.Next() {
		var request types.SnapshotRequest
		must.OK(request.Unmarshal(iterator.Value()))

		if request.Height != height {
			continue
		}

		snapshotIndex := snapshotstore.NewSnapshotKeyStore(store, request.Prefix.StoreName).ByName(request.Prefix.Name).TakeSnapshot()
		snapshot := types.Snapshot{
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
		frozenStore.Set(iterator.Key(), must.Bytes(snapshot.Marshal()))
		requestStore.Delete(iterator.Key())
	}
}

func (k Keeper) GetPending(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.SnapshotRequest, error) {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AccountPendingSnapshotsPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.SnapshotRequest
	for ; iterator.Valid(); iterator.Next() {
		var request types.SnapshotRequest
		must.OK(request.Unmarshal(iterator.Value()))
		res = append(res, request)
	}

	return res, nil
}

func (k Keeper) GetSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.Snapshot, error) {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AccountTakenSnapshotsPrefix(accAddress))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.Snapshot
	for ; iterator.Valid(); iterator.Next() {
		var snapshot types.Snapshot
		must.OK(snapshot.Unmarshal(iterator.Value()))
		res = append(res, snapshot)
	}

	return res, nil
}
