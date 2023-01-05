package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, definition := range genState.ClassDefinitions {
		k.SetClassDefinition(ctx, definition)
	}
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	classDefinitions, _, err := k.GetClassDefinitions(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		ClassDefinitions: classDefinitions,
		Params:           k.GetParams(ctx),
	}
}
