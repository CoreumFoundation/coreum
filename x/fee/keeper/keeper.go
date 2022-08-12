package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

var _ Keeper = (*BaseKeeper)(nil)

// Keeper defines a module interface that facilitates the transfer of coins
// between accounts.
type Keeper interface {
	TrackedGas(ctx sdk.Context) int64
	TrackGas(ctx sdk.Context, gas int64)
	GetShortAverageGas(ctx sdk.Context) int64
	SetShortAverageGas(ctx sdk.Context, averageGas int64)
	GetLongAverageGas(ctx sdk.Context) int64
	SetLongAverageGas(ctx sdk.Context, averageGas int64)
	GetMinGasPrice(ctx sdk.Context) sdk.Coin
	SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin)
}

// BaseKeeper manages transfers between accounts. It implements the Keeper interface.
type BaseKeeper struct {
	feeDenom          string
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
	feeDenom string,
	storeKey sdk.StoreKey,
	transientStoreKey sdk.StoreKey,
) BaseKeeper {
	return BaseKeeper{
		feeDenom:          feeDenom,
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
	}
}

var (
	gasTrackingKey       = []byte{0x00}
	gasPriceKey          = []byte{0x01}
	currentAverageGasKey = []byte{0x02}
	averageGasKey        = []byte{0x03}
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

// GetShortAverageGas retrieves average gas used by previous blocks, used as a representation of smoothed gas used by latest block
func (k BaseKeeper) GetShortAverageGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(currentAverageGasKey)

	if bz == nil {
		return 0
	}

	currentAverageGas := sdk.NewInt(0)
	if err := currentAverageGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return currentAverageGas.Int64()
}

// SetShortAverageGas sets average gas used by previous blocks, used as a representation of smoothed gas used by latest block
func (k BaseKeeper) SetShortAverageGas(ctx sdk.Context, currentAverageGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(currentAverageGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(currentAverageGasKey, bz)
}

// GetLongAverageGas retrieves long average gas used by previous blocks, used for determining average block load where maximum discount is applied
func (k BaseKeeper) GetLongAverageGas(ctx sdk.Context) int64 {
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

// SetLongAverageGas sets long average gas used by previous blocks, used for determining average block load where maximum discount is applied
func (k BaseKeeper) SetLongAverageGas(ctx sdk.Context, averageGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(averageGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(averageGasKey, bz)
}

// GetMinGasPrice returns current minimum gas price required by the network
func (k BaseKeeper) GetMinGasPrice(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(gasPriceKey)
	if bz == nil {
		panic(errors.New("minimum gas price is not set"))
	}
	var minGasPrice sdk.Coin
	if err := minGasPrice.Unmarshal(bz); err != nil {
		panic(err)
	}
	return minGasPrice
}

// SetMinGasPrice sets minimum gas price required by the network on current block
func (k BaseKeeper) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz, err := minGasPrice.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(gasPriceKey, bz)
}
