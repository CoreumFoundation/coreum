package keeper

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/gogoproto/proto"

	"github.com/CoreumFoundation/coreum/v4/x/delay/types"
)

// Keeper is delay module Keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	router   types.Router
	registry codectypes.InterfaceRegistry
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
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

// DelayExecution stores an item to be executed in delay time.
func (k Keeper) DelayExecution(ctx sdk.Context, id string, data proto.Message, delay time.Duration) error {
	return k.StoreDelayedExecution(ctx, id, data, ctx.BlockTime().Add(delay))
}

// ExecuteAfter stores an item to be executed after specified time.
func (k Keeper) ExecuteAfter(ctx sdk.Context, id string, data proto.Message, time time.Time) error {
	return k.StoreDelayedExecution(ctx, id, data, time)
}

// ExecuteAfterBlock stores an item to be executed after specified block.
func (k Keeper) ExecuteAfterBlock(ctx sdk.Context, id string, data proto.Message, height uint64) error {
	return k.StoreBlockExecution(ctx, id, data, height)
}

// RemoveExecuteAtBlock removes an item to be executed at specified block.
func (k Keeper) RemoveExecuteAtBlock(ctx sdk.Context, id string, height uint64) error {
	key, err := types.CreateBlockItemKey(id, height)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	return nil
}

// RemoveExecuteAfter removes an item to be executed at after specified time.
func (k Keeper) RemoveExecuteAfter(ctx sdk.Context, id string, time time.Time) error {
	key, err := types.CreateDelayedItemKey(id, time)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	return nil
}

// StoreDelayedExecution stores delayed execution item using absolute time.
func (k Keeper) StoreDelayedExecution(ctx sdk.Context, id string, data proto.Message, time time.Time) error {
	if !k.router.Has(data) {
		return sdkerrors.Wrapf(
			types.ErrInvalidData,
			"the router does not support this type, id: %s, data: %s",
			id, proto.MessageName(data),
		)
	}
	key, err := types.CreateDelayedItemKey(id, time)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "delayed item is already stored under the key, id: %s", id)
	}

	dataAny, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "failed to construct new Any, err: %s", err)
	}

	b, err := k.cdc.Marshal(dataAny)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "marshaling delayed item failed: %s", err.Error())
	}
	store.Set(key, b)
	return nil
}

// StoreBlockExecution stores block execution item using block height.
func (k Keeper) StoreBlockExecution(ctx sdk.Context, id string, data proto.Message, height uint64) error {
	if !k.router.Has(data) {
		return sdkerrors.Wrapf(
			types.ErrInvalidData,
			"the router does not support this type, id: %s, data: %s",
			id, proto.MessageName(data),
		)
	}

	key, err := types.CreateBlockItemKey(id, height)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "block item is already stored under the key, id: %s", id)
	}

	dataAny, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "failed to construct new Any, err: %s", err)
	}

	b, err := k.cdc.Marshal(dataAny)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "marshaling block item failed: %s", err.Error())
	}
	store.Set(key, b)
	return nil
}

// ExecuteAllItems executes delayed and block items for the current block time and height.
func (k Keeper) ExecuteAllItems(ctx sdk.Context) error {
	if err := k.ExecuteDelayedItems(ctx); err != nil {
		return err
	}

	return k.ExecuteBlockItems(ctx)
}

// ExecuteDelayedItems executes delayed logic.
func (k Keeper) ExecuteDelayedItems(ctx sdk.Context) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedItemKeyPrefix)

	// messages will be returned from this iterator in the execution time ascending order
	iter := store.Iterator(nil, nil)

	blockTime := ctx.BlockTime()
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		execTime, _, err := types.DecodeDelayedItemKey(key)
		if err != nil {
			return err
		}

		// due to the order of items returned by the iterator, if we find that execution time is after
		// the current block time, then there is no reason to iterate further
		if execTime.After(blockTime) {
			return nil
		}

		if err := k.executeMessage(ctx, iter.Value()); err != nil {
			return err
		}

		store.Delete(key)
	}
	return nil
}

// ImportDelayedItems imports delayed items.
func (k Keeper) ImportDelayedItems(ctx sdk.Context, items []types.DelayedItem) error {
	for _, i := range items {
		var data proto.Message
		if err := k.registry.UnpackAny(i.Data, &data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking of execution message failed: %s", err.Error())
		}

		if err := k.StoreDelayedExecution(ctx, i.ID, data, i.ExecutionTime); err != nil {
			return err
		}
	}
	return nil
}

// ExportDelayedItems exports delayed items.
//
//nolint:dupl // there is not duplication the code is similar in terms of structure, but different in terms of logic
func (k Keeper) ExportDelayedItems(ctx sdk.Context) ([]types.DelayedItem, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedItemKeyPrefix)
	delayedItems := make([]types.DelayedItem, 0)
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.PaginationMaxLimit}, func(key, value []byte) error {
		executionTime, id, err := types.DecodeDelayedItemKey(key)
		if err != nil {
			return err
		}

		data := &codectypes.Any{}
		if err := k.cdc.Unmarshal(value, data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking of delayed message failed: %s", err.Error())
		}

		delayedItems = append(delayedItems, types.DelayedItem{
			ID:            id,
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

// ExecuteBlockItems executes block logic.
func (k Keeper) ExecuteBlockItems(ctx sdk.Context) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BlockItemKeyPrefix)

	// messages will be returned from this iterator in the execution time ascending order
	iter := store.Iterator(nil, nil)

	currentHeight := uint64(ctx.BlockHeight())
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()

		height, _, err := types.DecodeBlockItemKey(key)
		if err != nil {
			return err
		}

		// stop execution if the block height is greater than the current block height
		if height >= currentHeight {
			return nil
		}

		if err := k.executeMessage(ctx, iter.Value()); err != nil {
			return err
		}

		store.Delete(key)
	}
	return nil
}

// ImportBlockItems imports block items.
func (k Keeper) ImportBlockItems(ctx sdk.Context, items []types.BlockItem) error {
	for _, i := range items {
		var data proto.Message
		if err := k.registry.UnpackAny(i.Data, &data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking of execution message failed: %s", err.Error())
		}

		if err := k.StoreBlockExecution(ctx, i.ID, data, i.Height); err != nil {
			return err
		}
	}
	return nil
}

// ExportBlockItems exports block items.
//
//nolint:dupl // there is not duplication the code is similar in terms of structure, but different in terms of logic
func (k Keeper) ExportBlockItems(ctx sdk.Context) ([]types.BlockItem, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.BlockItemKeyPrefix)
	blockItems := make([]types.BlockItem, 0)
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.PaginationMaxLimit}, func(key, value []byte) error {
		height, id, err := types.DecodeBlockItemKey(key)
		if err != nil {
			return err
		}

		data := &codectypes.Any{}
		if err := k.cdc.Unmarshal(value, data); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking of block message failed: %s", err.Error())
		}

		blockItems = append(blockItems, types.BlockItem{
			ID:     id,
			Height: height,
			Data:   data,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return blockItems, nil
}

func (k Keeper) executeMessage(ctx sdk.Context, messageData []byte) error {
	dataAny := &codectypes.Any{}
	if err := k.cdc.Unmarshal(messageData, dataAny); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "decoding of execution message failed: %s", err.Error())
	}

	var data proto.Message
	if err := k.cdc.UnpackAny(dataAny, &data); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidData, "unpacking of execution message failed: %s", err.Error())
	}

	handler, err := k.router.Handler(data)
	if err != nil {
		return err
	}

	return handler(ctx, data)
}
