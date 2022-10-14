package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
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

func (k Keeper) RequestFreeze(ctx sdk.Context, request types.FreezeRequest) error {
	// FIXME (wojtek): validate request

	store := ctx.KVStore(k.storeKey)

	index := sdk.ZeroInt()
	bz := store.Get(types.FreezeRequestsIndexKey)
	if bz != nil {
		must.OK(index.Unmarshal(bz))
		index = index.Add(sdk.OneInt())
	}
	indexMarshaled := must.Bytes(index.Marshal())
	store.Set(types.FreezeRequestsIndexKey, indexMarshaled)

	ownerAddress, err := sdk.AccAddressFromBech32(request.Owner)
	if err != nil {
		return errors.WithStack(err)
	}
	store.Set(types.Join(types.FreezeRequestsKey, address.MustLengthPrefix(ownerAddress), indexMarshaled), must.Bytes(request.Marshal()))
	return nil
}

func (k Keeper) Freeze(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	requestStore := prefix.NewStore(store, types.FreezeRequestsKey)
	frozenStore := prefix.NewStore(store, types.FrozenKey)
	iterator := requestStore.Iterator(nil, nil)
	defer iterator.Close()

	height := ctx.BlockHeight()
	for ; iterator.Valid(); iterator.Next() {
		var request types.FreezeRequest
		must.OK(request.Unmarshal(iterator.Value()))

		if request.Height != height {
			continue
		}

		snapshotIndex := snapshotstore.NewSnapshotKeyStore(store, request.SnapshotID.StoreName).ByName(request.SnapshotID.Name).Freeze()
		frozen := types.FrozenSnapshot{
			SnapshotID:    request.SnapshotID,
			SnapshotIndex: snapshotIndex,
			Owner:         request.Owner,
			Height:        request.Height,
			Name:          request.Name,
			Description:   request.Description,
		}
		frozenStore.Set(iterator.Key(), must.Bytes(frozen.Marshal()))
		requestStore.Delete(iterator.Key())
	}
}

func (k Keeper) GetPendingFreezeRequests(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.FreezeRequest, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.Join(types.FreezeRequestsKey, address.MustLengthPrefix(accAddress)))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.FreezeRequest
	for ; iterator.Valid(); iterator.Next() {
		var request types.FreezeRequest
		must.OK(request.Unmarshal(iterator.Value()))
		res = append(res, request)
	}

	return res, nil
}

func (k Keeper) GetFrozenSnapshots(ctx sdk.Context, accAddress sdk.AccAddress) ([]types.FrozenSnapshot, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.Join(types.FrozenKey, address.MustLengthPrefix(accAddress)))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.FrozenSnapshot
	for ; iterator.Valid(); iterator.Next() {
		var snapshot types.FrozenSnapshot
		must.OK(snapshot.Unmarshal(iterator.Value()))
		res = append(res, snapshot)
	}

	return res, nil
}
