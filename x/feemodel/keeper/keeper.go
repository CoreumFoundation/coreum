package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

// Keeper is a fee model keeper.
type Keeper struct {
	storeKey          storetypes.StoreKey
	transientStoreKey storetypes.StoreKey
	cdc               codec.BinaryCodec
	authority         string
}

// NewKeeper returns a new keeper object providing storage options required by fee model.
func NewKeeper(
	storeKey storetypes.StoreKey,
	transientStoreKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority string,
) Keeper {
	return Keeper{
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
		cdc:               cdc,
		authority:         authority,
	}
}

// TrackedGas returns gas limits declared by transactions executed so far in current block.
func (k Keeper) TrackedGas(ctx sdk.Context) int64 {
	tStore := ctx.TransientStore(k.transientStoreKey)

	gasUsed := sdkmath.NewInt(0)
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
	bz, err := sdkmath.NewInt(k.TrackedGas(ctx) + gas).Marshal()
	if err != nil {
		panic(err)
	}
	tStore.Set(gasTrackingKey, bz)
}

// SetParams sets the parameters of the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(paramsKey, bz)
	return nil
}

// GetParams gets the parameters of the module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(paramsKey)
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// UpdateParams is a governance operation that sets parameters of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return k.SetParams(ctx, params)
}

// GetShortEMAGas retrieves average gas used by previous blocks, used as a representation of smoothed gas used by latest block.
func (k Keeper) GetShortEMAGas(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(shortEMAGasKey)

	if bz == nil {
		return 0
	}

	currentEMAGas := sdkmath.NewInt(0)
	if err := currentEMAGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return currentEMAGas.Int64()
}

// SetShortEMAGas sets average gas used by previous blocks, used as a representation of smoothed gas used by latest block.
func (k Keeper) SetShortEMAGas(ctx sdk.Context, emaGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdkmath.NewInt(emaGas).Marshal()
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

	emaGas := sdkmath.NewInt(0)
	if err := emaGas.Unmarshal(bz); err != nil {
		panic(err)
	}
	return emaGas.Int64()
}

// SetLongEMAGas sets long average gas used by previous blocks, used for determining average block load where maximum discount is applied.
func (k Keeper) SetLongEMAGas(ctx sdk.Context, emaGas int64) {
	store := ctx.KVStore(k.storeKey)

	bz, err := sdkmath.NewInt(emaGas).Marshal()
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
		return sdk.DecCoin{}, sdk.DecCoin{}, sdkerrors.Wrapf(cosmoserrors.ErrInvalidRequest, "after blocks must be lower than or equal to %d", shortEMABlockLength)
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
