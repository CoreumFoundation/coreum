package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// Keeper manages transfers between accounts. It implements the Keeper interface.
type Keeper struct {
	paramSubspace     paramtypes.Subspace
	feeDenom          string
	storeKey          sdk.StoreKey
	transientStoreKey sdk.StoreKey
}

// NewKeeper returns a new keeper object providing storage options required by fee model.
func NewKeeper(
	paramSubspace paramtypes.Subspace,
	feeDenom string,
	storeKey sdk.StoreKey,
	transientStoreKey sdk.StoreKey,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSubspace.HasKeyTable() {
		paramSubspace = paramSubspace.WithKeyTable(paramtypes.NewKeyTable().RegisterParamSet(&types.Model{}))
	}

	return Keeper{
		paramSubspace:     paramSubspace,
		feeDenom:          feeDenom,
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
	}
}

// TrackedGas returns gas limits declared by transactions executed so far in current block
func (k Keeper) TrackedGas(ctx sdk.Context) int64 {
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
func (k Keeper) TrackGas(ctx sdk.Context, gas int64) {
	tStore := ctx.TransientStore(k.transientStoreKey)
	bz, err := sdk.NewInt(k.TrackedGas(ctx) + gas).Marshal()
	if err != nil {
		panic(err)
	}
	tStore.Set(gasTrackingKey, bz)
}

// SetModel sets the parameters of the model
func (k Keeper) SetModel(ctx sdk.Context, model types.Model) {
	k.paramSubspace.SetParamSet(ctx, &model)
}

// GetModel gets the parameters of the model
func (k Keeper) GetModel(ctx sdk.Context) types.Model {
	var model types.Model
	k.paramSubspace.GetParamSet(ctx, &model)
	return model
}

// GetShortAverageGas retrieves average gas used by previous blocks, used as a representation of smoothed gas used by latest block
func (k Keeper) GetShortAverageGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(shortAverageGasKey)

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
func (k Keeper) SetShortAverageGas(ctx sdk.Context, currentAverageGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(currentAverageGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(shortAverageGasKey, bz)
}

// GetLongAverageGas retrieves long average gas used by previous blocks, used for determining average block load where maximum discount is applied
func (k Keeper) GetLongAverageGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(longAverageGasKey)

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
func (k Keeper) SetLongAverageGas(ctx sdk.Context, averageGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(averageGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(longAverageGasKey, bz)
}

// GetMinGasPrice returns current minimum gas price required by the network
func (k Keeper) GetMinGasPrice(ctx sdk.Context) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(gasPriceKey)
	if bz == nil {
		return sdk.NewCoin(k.feeDenom, k.GetModel(ctx).InitialGasPrice)
	}
	var minGasPrice sdk.Coin
	if err := minGasPrice.Unmarshal(bz); err != nil {
		panic(err)
	}
	return minGasPrice
}

// SetMinGasPrice sets minimum gas price required by the network on current block
func (k Keeper) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	bz, err := minGasPrice.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(gasPriceKey, bz)
}
