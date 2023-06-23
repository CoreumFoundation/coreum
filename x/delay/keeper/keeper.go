package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/delay/types"
)

// Keeper is delay module Keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey sdk.StoreKey
	router   types.Router
	registry codectypes.InterfaceRegistry
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	router types.Router,
	registry codectypes.InterfaceRegistry,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		router:   router,
		registry: registry,
	}
}

// Router returns router.
func (k Keeper) Router() types.Router {
	return k.router
}

// DelayExecution stores an item to be executed later.
func (k Keeper) DelayExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, delay time.Duration) error {
	return k.StoreDelayedExecution(ctx, id, data, ctx.BlockTime().Add(delay))
}

// StoreDelayedExecution stores delayed execution item using absolute time.
func (k Keeper) StoreDelayedExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, t time.Time) error {
	key, err := types.CreateDelayedItemKey(id, t)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "delayed item is already stored under the key, id: %s", id)
	}

	dataAny, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		return err
	}

	b, err := k.cdc.Marshal(dataAny)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "marshaling delayed item failed: %s", err.Error())
	}
	store.Set(key, b)
	return nil
}

// ExecuteDelayedItems executes delayed logic.
func (k Keeper) ExecuteDelayedItems(ctx sdk.Context) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedItemKeyPrefix)

	// messages will be returned from this iterator in the execution time ascending order
	iter := store.Iterator(nil, nil)

	blockTime := ctx.BlockTime()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		execTime, _, err := types.ExtractTimeAndIDFromDelayedItemKey(key)
		if err != nil {
			return err
		}

		// due to the order of items returned by the iterator, if we find that execution time is after
		// the current block time, then there is no reason to iterate further
		if execTime.After(blockTime) {
			return nil
		}

		dataAny := &codectypes.Any{}
		if err := k.cdc.Unmarshal(iter.Value(), dataAny); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "decoding delayed message failed: %s", err.Error())
		}

		var data codec.ProtoMarshaler
		if err := k.cdc.UnpackAny(dataAny, &data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking delayed message failed: %s", err.Error())
		}

		handler, err := k.router.Handler(data)
		if err != nil {
			return err
		}
		if err := handler(ctx, data); err != nil {
			return err
		}

		store.Delete(key)
	}
	return nil
}

// ImportDelayedItems imports delayed items.
func (k Keeper) ImportDelayedItems(ctx sdk.Context, items []types.DelayedItem) error {
	for _, i := range items {
		var data codec.ProtoMarshaler
		if err := k.registry.UnpackAny(i.Data, &data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking delayed message failed: %s", err.Error())
		}

		if err := k.StoreDelayedExecution(ctx, i.Id, data, i.ExecutionTime); err != nil {
			return err
		}
	}
	return nil
}

// ExportDelayedItems exports delayed items.
func (k Keeper) ExportDelayedItems(ctx sdk.Context) ([]types.DelayedItem, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedItemKeyPrefix)
	delayedItems := []types.DelayedItem{}
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.MaxLimit}, func(key, value []byte) error {
		executionTime, id, err := types.ExtractTimeAndIDFromDelayedItemKey(key)
		if err != nil {
			return err
		}

		data := &codectypes.Any{}
		if err := k.cdc.Unmarshal(value, data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking delayed message failed: %s", err.Error())
		}

		delayedItems = append(delayedItems, types.DelayedItem{
			Id:            id,
			ExecutionTime: executionTime,
			Data:          data,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return delayedItems, nil
}
