package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	TrackedGas(ctx sdk.Context) int64
	TrackGas(ctx sdk.Context, gas int64)
	GetAverageGas(ctx sdk.Context) int64
	SetAverageGas(ctx sdk.Context, averageGas int64)
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	storeKey          sdk.StoreKey
	transientStoreKey sdk.StoreKey
}

// NewBaseKeeper returns a new BaseKeeper object with a given codec, dedicated
// store key, an AccountKeeper implementation, and a parameter Subspace used to
// store and fetch module parameters. The BaseKeeper also accepts a
// blocklist map. This blocklist describes the set of addresses that are not allowed
// to receive funds through direct and explicit actions, for example, by using a MsgSend or
// by using a SendCoinsFromModuleToAccount execution.
func NewBaseKeeper(
	storeKey sdk.StoreKey,
	transientStoreKey sdk.StoreKey,
) BaseKeeper {
	return BaseKeeper{
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
	}
}

var (
	gasTrackingKey = []byte{0x00}
	averageGasKey  = []byte{0x01}
)

// TrackedGas returns gas limits declared by transactions executed so far in current block
func (k BaseKeeper) TrackedGas(ctx sdk.Context) int64 {
	tStore := ctx.TransientStore(k.transientStoreKey)

	gasUsed := sdk.NewInt(0)
	bz := tStore.Get(gasTrackingKey)

	if bz != nil {
		if err := gasUsed.Unmarshal(bz); err != nil {
			panic(err)
		}
	}

	return gasUsed.Int64()
}

// TrackGas increments gas tracked for current block
func (k BaseKeeper) TrackGas(ctx sdk.Context, gas int64) {
	tStore := ctx.TransientStore(k.transientStoreKey)
	bz, err := sdk.NewInt(k.TrackedGas(ctx) + gas).Marshal()
	if err != nil {
		panic(err)
	}
	tStore.Set(gasTrackingKey, bz)
}

// GetAverageGas retrieves latest average gas used by previous blocks
func (k BaseKeeper) GetAverageGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(averageGasKey)

	if bz == nil {
		return 0
	}

	averageGas := sdk.NewInt(0)
	if err := averageGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return averageGas.Int64()
}

// SetAverageGas sets latest average gas used by previous blocks
func (k BaseKeeper) SetAverageGas(ctx sdk.Context, averageGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(averageGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(averageGasKey, bz)
}
