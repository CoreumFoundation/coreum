<!--
order: 2
-->

# Keeper

The feemodel module provides a keeper providing these methods:

```go
type Keeper interface {
    // TrackedGas returns gas limits declared by transactions executed so far in current block
    TrackedGas(ctx sdk.Context) int64
	
    // TrackGas increments gas tracked for current block
    TrackGas(ctx sdk.Context, gas int64)

    // SetParams sets the parameters of the model
    SetParams(ctx sdk.Context, params types.Params)

    // GetParams gets the parameters of the model
    GetParams(ctx sdk.Context) types.Params

    // GetShortEMAGas retrieves average gas used by previous blocks, used as a representation of smoothed gas used by latest block
    GetShortEMAGas(ctx sdk.Context) int64

    // SetShortEMAGas sets average gas used by previous blocks, used as a representation of smoothed gas used by latest block
    SetShortEMAGas(ctx sdk.Context, emaGas int64)

    // GetLongEMAGas retrieves long average gas used by previous blocks, used for determining average block load where maximum discount is applied
    GetLongEMAGas(ctx sdk.Context) int64

    // SetLongEMAGas sets long average gas used by previous blocks, used for determining average block load where maximum discount is applied
    SetLongEMAGas(ctx sdk.Context, emaGas int64)

    // GetMinGasPrice returns current minimum gas price required by the network
    GetMinGasPrice(ctx sdk.Context) sdk.DecCoin

    // SetMinGasPrice sets minimum gas price required by the network on current block
    SetMinGasPrice(ctx sdk.Context, minGasPrice sdk.DecCoin)
}
```

From all of these methods only `GetMinGasPrice` should be used by other modules. All the other ones serve internal needs of feemodel module.
