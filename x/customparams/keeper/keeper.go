package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/x/customparams/types"
)

// Keeper is customparams module Keeper.
type Keeper struct {
	stakingParamSpace paramtypes.Subspace
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(stakingParamSpace paramtypes.Subspace) Keeper {
	// set KeyTable if it has not already been set
	if !stakingParamSpace.HasKeyTable() {
		stakingParamSpace = stakingParamSpace.WithKeyTable(types.StakingParamKeyTable())
	}

	return Keeper{
		stakingParamSpace: stakingParamSpace,
	}
}

// GetStakingParams returns the set of staking parameters.
func (k Keeper) GetStakingParams(ctx sdk.Context) types.StakingParams {
	var stakingParams types.StakingParams
	k.stakingParamSpace.GetParamSet(ctx, &stakingParams)
	return stakingParams
}

// SetStakingParams sets the module staking parameters to the param space.
func (k Keeper) SetStakingParams(ctx sdk.Context, params types.StakingParams) {
	k.stakingParamSpace.SetParamSet(ctx, &params)
}
