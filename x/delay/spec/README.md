# x/delay

## Abstract

This document descibes the functionality of the `delay` module. It is responsible for delayed execution of the predefined logic. Developers may specify a data structure and corresponding handler (logic). Then, this data structure might be created at any time (e.g. when transaction message is executed) and stored to be executed later at particular time, when new block with the matching block time is created.

## State

State managed by the module:

- DelayedMessages: `0x01 | -> any`

## Keeper

The `delay` module provides a keeper providing these methods:

```go
type Keeper interface {
// Router returns router.
func (k Keeper) Router() types.Router

// DelayExecution stores an item to be executed later.
func (k Keeper) DelayExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, delay time.Duration) error

// StoreDelayedExecution stores delayed execution item using absolute time.
func (k Keeper) StoreDelayedExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, t time.Time) error

// ExecuteDelayedItems executes delayed logic. It executes all the previously stored delayed items having the execution time
// equal to or earlier than the current block time.
func (k Keeper) ExecuteDelayedItems(ctx sdk.Context) error

// ImportDelayedItems imports delayed items. Used for importing genesis state only.
func (k Keeper) ImportDelayedItems(ctx sdk.Context, items []types.DelayedItem) error

// ExportDelayedItems exports delayed items. Used for exporting genesis state only.
func (k Keeper) ExportDelayedItems(ctx sdk.Context) ([]types.DelayedItem, error)
}
```
