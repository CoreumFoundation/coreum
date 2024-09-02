package dex

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// InitGenesis initializes the dex module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	// TODO(dex): implement for missing pars
}

// ExportGenesis returns the dex module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// TODO(dex): implement for missing pars
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
