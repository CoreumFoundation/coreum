package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/x/wstaking/types"
)

// Keeper is wstaking module Keeper.
type Keeper struct {
	paramSpace paramtypes.Subspace
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(paramSubspace paramtypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !paramSubspace.HasKeyTable() {
		paramSubspace = paramSubspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		paramSpace: paramSubspace,
	}
}

// GetParams returns the total set of module parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the module parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
