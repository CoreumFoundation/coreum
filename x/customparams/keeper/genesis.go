package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/customparams/types"
)

// InitGenesis initializes the customparams module's state with the provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if err := k.SetStakingParams(ctx, genState.StakingParams); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the customparams module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return &types.GenesisState{
		StakingParams: k.GetStakingParams(ctx),
	}
}
