package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// ParamSubspace represents a subscope of methods exposed by param module to store and retrieve parameters.
type ParamSubspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}

// Keeper is a fee model keeper.
type Keeper struct {
	paramSubspace     ParamSubspace
	storeKey          sdk.StoreKey
	transientStoreKey sdk.StoreKey
}

// NewKeeper returns a new keeper object providing storage options required by fee model.
func NewKeeper(
	paramSubspace ParamSubspace,
	storeKey sdk.StoreKey,
	transientStoreKey sdk.StoreKey,
) Keeper {
	return Keeper{
		paramSubspace:     paramSubspace,
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
	}
}

// TrackedGas returns gas limits declared by transactions executed so far in current block.
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

// TrackGas increments gas tracked for current block.
func (k Keeper) TrackGas(ctx sdk.Context, gas int64) {
	tStore := ctx.TransientStore(k.transientStoreKey)
	bz, err := sdk.NewInt(k.TrackedGas(ctx) + gas).Marshal()
	if err != nil {
		panic(err)
	}
	tStore.Set(gasTrackingKey, bz)
}

// SetParams sets the parameters of the model.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the parameters of the model.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// GetShortEMAGas retrieves average gas used by previous blocks, used as a representation of smoothed gas used by latest block.
func (k Keeper) GetShortEMAGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(shortEMAGasKey)

	if bz == nil {
		return 0
	}

	currentEMAGas := sdk.NewInt(0)
	if err := currentEMAGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return currentEMAGas.Int64()
}

// SetShortEMAGas sets average gas used by previous blocks, used as a representation of smoothed gas used by latest block.
func (k Keeper) SetShortEMAGas(ctx sdk.Context, emaGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(emaGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(shortEMAGasKey, bz)
}

// GetLongEMAGas retrieves long average gas used by previous blocks, used for determining average block load where maximum discount is applied.
func (k Keeper) GetLongEMAGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(longEMAGasKey)

	if bz == nil {
		return 0
	}

	emaGas := sdk.NewInt(0)
	if err := emaGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return emaGas.Int64()
}

// SetLongEMAGas sets long average gas used by previous blocks, used for determining average block load where maximum discount is applied.
func (k Keeper) SetLongEMAGas(ctx sdk.Context, emaGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdk.NewInt(emaGas).Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(longEMAGasKey, bz)
}

// GetMinGasPrice returns current minimum gas price required by the network.
func (k Keeper) GetMinGasPrice(ctx sdk.Context) sdk.DecCoin {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(gasPriceKey)
	if bz == nil {
		// This is really a panic condition because it means that genesis initialization was not done correctly
		panic("min gas price not set")
	}
	var minGasPrice sdk.DecCoin
	if err := minGasPrice.Unmarshal(bz); err != nil {
		panic(err)
	}
	return minGasPrice
}

// SetMinGasPrice sets minimum gas price required by the network on current block.
func (k Keeper) SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.DecCoin) {
	store := ctx.KVStore(k.storeKey)
	bz, err := minGasPrice.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set(gasPriceKey, bz)
}

// CalculateEdgeGasPriceAfterBlocks returns the smallest and highest possible values for min gas price in future blocks.
func (k Keeper) CalculateEdgeGasPriceAfterBlocks(ctx sdk.Context, after uint32) (sdk.DecCoin, sdk.DecCoin, error) {
	shortEMABlockLength := k.GetParams(ctx).Model.ShortEmaBlockLength
	if after > shortEMABlockLength {
		return sdk.DecCoin{}, sdk.DecCoin{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "after blocks must be lower than or equal to %d", shortEMABlockLength)
	}

	// if no after value is provided shortEMABlockLength is taken as default value
	if after == 0 {
		after = shortEMABlockLength
	}

	params := k.GetParams(ctx)
	shortEMA := k.GetShortEMAGas(ctx)
	longEMA := k.GetLongEMAGas(ctx)

	maxShortEMA := shortEMA
	minShortEMA := shortEMA

	maxLongEMA := longEMA
	minLongEMA := longEMA

	model := types.NewModel(params.Model)
	minGasPrice := model.CalculateNextGasPrice(shortEMA, longEMA)

	lowMinGasPrice := minGasPrice
	highMinGasPrice := minGasPrice
	minBlockGas := int64(0)
	maxBlockGas := params.Model.MaxBlockGas

	for i := uint32(0); i < after; i++ {
		maxShortEMA = types.CalculateEMA(maxShortEMA, maxBlockGas,
			params.Model.ShortEmaBlockLength)
		maxLongEMA = types.CalculateEMA(maxLongEMA, params.Model.MaxBlockGas,
			params.Model.LongEmaBlockLength)
		maxLoadMinGasPrice := model.CalculateNextGasPrice(maxShortEMA, maxLongEMA)

		minShortEMA = types.CalculateEMA(minShortEMA, minBlockGas,
			params.Model.ShortEmaBlockLength)
		minLongEMA = types.CalculateEMA(minLongEMA, minBlockGas,
			params.Model.LongEmaBlockLength)
		minLoadMinGasPrice := model.CalculateNextGasPrice(minShortEMA, minLongEMA)

		highMinGasPrice = sdk.MaxDec(highMinGasPrice, sdk.MaxDec(maxLoadMinGasPrice, minLoadMinGasPrice))
		lowMinGasPrice = sdk.MinDec(lowMinGasPrice, sdk.MinDec(maxLoadMinGasPrice, minLoadMinGasPrice))
	}

	denom := k.GetMinGasPrice(ctx).Denom
	return sdk.NewDecCoinFromDec(denom, lowMinGasPrice),
		sdk.NewDecCoinFromDec(denom, highMinGasPrice),
		nil
}
